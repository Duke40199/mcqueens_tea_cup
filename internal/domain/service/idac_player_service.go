package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

var reMention = regexp.MustCompile(`^<@!?(\d+)>$`)

type IDACPlayerService struct {
	// clients
	segaClient port.SegaIDACClient
	// repos
	aliasRepo        database.AliasRepository
	rankingCfgRepo   database.RankingCfgRepository
	obRankingCfgRepo database.OBRankingCfgRepository
}

func NewIDACPlayerService(
	segaClient port.SegaIDACClient,
	aliasRepo database.AliasRepository,
	rankingCfgRepo database.RankingCfgRepository,
	obRankingCfgRepo database.OBRankingCfgRepository,
) port.IDACPlayerService {
	return &IDACPlayerService{
		segaClient:       segaClient,
		aliasRepo:        aliasRepo,
		rankingCfgRepo:   rankingCfgRepo,
		obRankingCfgRepo: obRankingCfgRepo,
	}
}

// ResolvePlayer resolves a player reference (a Discord mention/ID or a custom
// tag) plus an optional manual area into a concrete (ign, area) pair using the
// alias store. The returned bool reports whether an alias was matched. An empty
// area with no error signals "area missing" to the caller.
func (s *IDACPlayerService) ResolvePlayer(ctx context.Context, input, manualArea string) (string, string, bool, error) {
	cleanInput := strings.TrimSpace(input)

	lookupKey := cleanInput
	if match := reMention.FindStringSubmatch(cleanInput); len(match) > 1 {
		lookupKey = match[1] // extract ID from <@123>
	}

	var aliasIgn, aliasArea string
	var foundAlias bool

	if lookupKey != "" {
		// Discord ID -> strict lookup
		playerAlias, isFound, err := s.aliasRepo.GetByAliasKey(lookupKey)
		if err != nil {
			return "", "", false, err
		}
		if !isFound {
			return "", "", false, fmt.Errorf("couldn't find a matching alias")
		}
		foundAlias = true
		aliasIgn = playerAlias.Ign
		aliasArea = playerAlias.Area
	} else {
		areaCode := entity.AreaAliases[manualArea]
		// text input -> custom tag lookup (case-insensitive)
		val, ok, err := s.aliasRepo.GetByIgnAndAreaCode(strings.ToLower(cleanInput), areaCode)
		if !ok || err != nil {
			return "", "", false, fmt.Errorf("couldn't find a matching alias")
		}
		aliasIgn = val.Ign
		aliasArea = val.Area
		foundAlias = true
	}

	finalIgn := cleanInput
	finalArea := aliasArea
	if foundAlias {
		finalIgn = aliasIgn
	}
	if finalArea == "" {
		return "", "", false, nil // signal "missing area"
	}
	return finalIgn, finalArea, foundAlias, nil
}

// GetPlayerProfile resolves a player's account grade and Online Battle rank from
// the Sega client and the ranking config repos. A nil view with a nil error
// means the player grade could not be found.
func (s *IDACPlayerService) GetPlayerProfile(ctx context.Context, ign, area string) (*entity.PlayerProfileView, error) {
	playerGrade, err := s.segaClient.GetPlayerGradeByIGN(ign, area)
	if err != nil {
		return nil, err
	}
	if playerGrade == nil {
		return nil, nil
	}

	playerGradeCfgs, err := s.rankingCfgRepo.GetPlayerGradeBySegaIDs(ctx, playerGrade.GradeID, playerGrade.NumberIcon)
	if err != nil {
		return nil, err
	}

	view := &entity.PlayerProfileView{}
	for _, cfg := range playerGradeCfgs {
		switch cfg.Type {
		case "RANK_NUMBER":
			view.GradeNum = cfg.Name
		case "GRADE":
			if cfg.Emoji != nil {
				view.GradeName = *cfg.Emoji
			} else {
				view.GradeName = cfg.Name
			}
		}
	}

	obRankingRes, err := s.segaClient.GetOBRankingByIGN(ign, "all", area)
	if err != nil {
		return nil, err
	}
	if obRankingRes != nil {
		obRankingCfgMap, err := s.obRankingCfgRepo.GetRankingCfgMap()
		if err != nil {
			return nil, err
		}
		if cfg, ok := obRankingCfgMap[obRankingRes.OnlineBattleRankId]; ok {
			view.OBRankName = cfg.Name
		}
		view.OBStarCount = obRankingRes.GetDisplayStarCount()
	}
	return view, nil
}

// GetTournamentInfo aggregates Online Battle ranking and account-grade data for
// an area into a per-player view, sorted by Online Battle points descending.
func (s *IDACPlayerService) GetTournamentInfo(ctx context.Context, areaCode string) (*entity.TournamentInfoView, error) {
	listGradeCfg, err := s.rankingCfgRepo.GetListPlayerGradeCfg(ctx)
	if err != nil {
		return nil, err
	}
	gradeCfgMap := make(map[string]*entity.PlayerGradeCfg, len(listGradeCfg))
	for _, cfg := range listGradeCfg {
		gradeCfgMap[cfg.SegaID] = cfg
	}

	obRankingRes, err := s.segaClient.GetListOBRanking("all", areaCode)
	if err != nil {
		return nil, err
	}
	listSegaPlayerGrade, err := s.segaClient.GetListPlayerGrade(areaCode)
	if err != nil {
		return nil, err
	}
	obRankingCfgMap, err := s.obRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		return nil, err
	}

	// Aggregate by player name across both Sega sources.
	mapFoundPlayers := make(map[string]entity.PlayerTournamentInfo)
	for _, obRanking := range obRankingRes.Records {
		foundPlayer, ok := mapFoundPlayers[obRanking.Name]
		if !ok {
			foundPlayer = entity.PlayerTournamentInfo{Name: obRanking.Name}
		}
		if foundOBRankCfg, ok := obRankingCfgMap[obRanking.OnlineBattleRankId]; ok {
			foundPlayer.OBRank = foundOBRankCfg.Name
			foundPlayer.OBRankNum = obRanking.GetDisplayStarCount()
			exp, _ := strconv.Atoi(obRanking.Point)
			foundPlayer.OBRankExp = exp
		}
		mapFoundPlayers[obRanking.Name] = foundPlayer
	}
	for _, segaPlayerGrade := range listSegaPlayerGrade.Records {
		foundPlayer, ok := mapFoundPlayers[segaPlayerGrade.Name]
		if !ok {
			foundPlayer = entity.PlayerTournamentInfo{Name: segaPlayerGrade.Name}
		}
		foundPlayer.GradeExp = segaPlayerGrade.GradeExp
		if foundGrade, ok := gradeCfgMap[segaPlayerGrade.GradeID]; ok {
			foundPlayer.Grade = foundGrade.Name
		}
		if foundGradeNum, ok := gradeCfgMap[segaPlayerGrade.NumberIcon]; ok {
			foundPlayer.GradeNum = foundGradeNum.Name
		}
		mapFoundPlayers[segaPlayerGrade.Name] = foundPlayer
	}

	players := make([]*entity.PlayerTournamentInfo, 0, len(mapFoundPlayers))
	for _, player := range mapFoundPlayers {
		players = append(players, &player)
	}
	sort.Slice(players, func(i, j int) bool {
		return players[i].OBRankExp > players[j].OBRankExp
	})

	return &entity.TournamentInfoView{
		CalcDate: obRankingRes.CalcDate,
		Players:  players,
	}, nil
}
