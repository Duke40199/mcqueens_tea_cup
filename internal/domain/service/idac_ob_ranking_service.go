package service

import (
	"context"
	"strconv"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
)

type IDACOBRankingService struct {
	// clients
	segaClient port.SegaIDACClient
	// repos
	obRankingCfgRepo database.OBRankingCfgRepository
}

func NewIDACOBRankingService(
	segaClient port.SegaIDACClient,
	obRankingCfgRepo database.OBRankingCfgRepository,
) port.IDACOBRankingService {
	return &IDACOBRankingService{
		segaClient:       segaClient,
		obRankingCfgRepo: obRankingCfgRepo,
	}
}

// GetRanking fetches the Online Battle ranking for an area/round, resolves each
// record's rank name and points against the ranking config, and returns at most
// `limit` entries (limit <= 0 means "all"). An empty view (no entries) is a
// valid, non-error result; upstream failures are returned as errors.
func (s *IDACOBRankingService) GetRanking(ctx context.Context, round, area string, limit int) (*entity.OBRankingView, error) {
	resp, err := s.segaClient.GetListOBRanking(round, area)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.Records) == 0 {
		return &entity.OBRankingView{}, nil
	}

	cfgMap, err := s.obRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(resp.Records) {
		limit = len(resp.Records)
	}

	view := &entity.OBRankingView{
		CalcDate: resp.CalcDate,
		Entries:  make([]entity.OBRankingEntry, 0, limit),
	}
	for j := 0; j < limit; j++ {
		r := resp.Records[j]
		// A record maps to a normal Online Battle rank when the config has an
		// entry for its OnlineBattleRankId; otherwise it is a Pride player and
		// we fall back to the Pride rank name and points.
		if rankName := cfgMap[r.OnlineBattleRankId].Name; rankName != "" {
			view.Entries = append(view.Entries, entity.OBRankingEntry{
				Rank:      r.Rank,
				Name:      r.Name,
				RankName:  rankName,
				Point:     r.Point,
				StarCount: r.GetDisplayStarCount(),
				IsPride:   false,
			})
			continue
		}
		view.Entries = append(view.Entries, entity.OBRankingEntry{
			Rank:     r.Rank,
			Name:     r.Name,
			RankName: cfgMap[r.PrideId].Name,
			Point:    strconv.Itoa(r.PridePoint),
			IsPride:  true,
		})
	}
	return view, nil
}
