package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
)

const segaTimeLayout = "2006/01/02 15:04:05"

type ActivePlayerSyncService struct {
	Notifier         port.Notifier
	SegaClient       port.SegaIDACClient
	AreaRepo         port.AreaRepository
	OBRankingCfgRepo port.OBRankingCfgRepository
	MetaLogic        *MetaLogicService
	Config           config.ActivePlayersSyncConfig
}

func NewActivePlayerSyncService(notifier port.Notifier, client port.SegaIDACClient, areaRepo port.AreaRepository, obRepo port.OBRankingCfgRepository, logic *MetaLogicService, cfg config.ActivePlayersSyncConfig) *ActivePlayerSyncService {
	return &ActivePlayerSyncService{
		Notifier:         notifier,
		SegaClient:       client,
		AreaRepo:         areaRepo,
		OBRankingCfgRepo: obRepo,
		MetaLogic:        logic,
		Config:           cfg,
	}
}

func (s *ActivePlayerSyncService) Sync(ctx context.Context) (string, error) {
	if s.Config.ChannelID == "" {
		return "", fmt.Errorf("ACTIVE_PLAYERS_CHANNEL_ID not configured")
	}

	// 1. Get active areas from DB
	areas, err := s.AreaRepo.GetOBActiveAreas(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch active areas: %w", err)
	}

	if len(areas) == 0 {
		logger.Warn(ctx, "no areas found for active player sync")
		return "", nil
	}

	// 2. Get rank configs
	obRankingCfgMap, err := s.OBRankingCfgRepo.GetRankingCfgMap(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get ranking cfg map: %w", err)
	}

	// 3. Get current round
	currentRound, err := s.SegaClient.GetCurrentRound(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get current round: %w", err)
	}
	roundStr := fmt.Sprintf("%d", currentRound)

	type playerActivity struct {
		Record    entity.OBRankingRecord
		LocalTime string
	}
	type areaActivity struct {
		GMT     string
		Players []playerActivity
	}

	var activePlayersByArea map[string]areaActivity
	maxPollingDuration := 14 * time.Minute
	pollingInterval := 1 * time.Minute
	startTime := time.Now()
	detectionTime := ""

	// 4. Fetch existing messages to extract state (Source of Truth)
	lastReportedTimeStr := ""
	lastDetectedHeaderPrefix := "_Detected at: "

	botMessages, err := s.Notifier.BotMessages(ctx, s.Config.ChannelID, 50)
	if err != nil {
		logger.Warn(ctx, fmt.Sprintf("could not fetch existing messages: %v", err))
		botMessages = nil
	}
	for _, m := range botMessages {
		if lastReportedTimeStr == "" && strings.Contains(m.Content, lastDetectedHeaderPrefix) {
			start := strings.Index(m.Content, lastDetectedHeaderPrefix) + len(lastDetectedHeaderPrefix)
			end := strings.Index(m.Content[start:], " (JST)")
			if end > -1 {
				lastReportedTimeStr = m.Content[start : start+end]
				logger.Info(ctx, fmt.Sprintf("found existing state in Discord, last reported: %s", lastReportedTimeStr))
			}
		}
	}

	// 5. Polling Phase (Canary Check)
	// Use a fixed offset instead of LoadLocation("Asia/Tokyo")
	jstLoc := time.FixedZone("JST", 9*60*60)
	canaryArea := areas[0]

	for {
		resp, err := s.SegaClient.GetListOBRanking(ctx, roundStr, canaryArea.AreaCode)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("polling error for canary %s", canaryArea.AreaName), err)
		} else if resp != nil {
			// State-Aware Refresh Check
			isNewerThanDiscord := true
			if lastReportedTimeStr != "" {
				// resp.CalcDate vs lastReportedTimeStr
				respTime, _ := time.ParseInLocation(segaTimeLayout, resp.CalcDate, jstLoc)
				discordTime, _ := time.ParseInLocation(segaTimeLayout, lastReportedTimeStr, jstLoc)
				isNewerThanDiscord = respTime.After(discordTime)
			}

			// Validation: Fresh by schedule AND newer than existing Discord state
			if s.MetaLogic.IsDataFresh(resp.CalcDate) && isNewerThanDiscord {
				logger.Info(ctx, fmt.Sprintf("fresh & newer data detected via canary (%s: %s), proceeding to full sync", canaryArea.AreaName, resp.CalcDate))
				detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
				break
			}

			if !isNewerThanDiscord {
				logger.Info(ctx, fmt.Sprintf("sega data (%s) already reported in Discord (%s), waiting for next block", resp.CalcDate, lastReportedTimeStr))
			} else {
				logger.Warn(ctx, fmt.Sprintf("sega is late (canary %s: %s), polling again in %v", canaryArea.AreaName, resp.CalcDate, pollingInterval))
			}
		}

		if time.Since(startTime) > maxPollingDuration {
			logger.Warn(ctx, "max polling duration reached, sega is significantly late, using latest available data")
			detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
			break
		}

		select {
		case <-time.After(pollingInterval):
			// continue loop
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// 6. Processing Phase (Fetch all areas)
	activePlayersByArea = make(map[string]areaActivity)
	for _, area := range areas {
		resp, err := s.SegaClient.GetListOBRanking(ctx, roundStr, area.AreaCode)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("error fetching ranking for %s (%s)", area.AreaName, area.AreaCode), err)
			continue
		}

		if resp == nil || len(resp.Records) == 0 {
			continue
		}

		calcTime, err := time.ParseInLocation(segaTimeLayout, resp.CalcDate, jstLoc)
		if err != nil {
			continue
		}

		localLoc, err := time.LoadLocation(area.Timezone)
		if err != nil {
			localLoc = jstLoc
		}

		_, offsetSeconds := time.Now().In(localLoc).Zone()
		offsetHours := offsetSeconds / 3600
		gmtStr := fmt.Sprintf("GMT%+d", offsetHours)

		var players []playerActivity
		for _, record := range resp.Records {
			updateTime, err := time.ParseInLocation(segaTimeLayout, record.UpdateDate, jstLoc)
			if err != nil {
				continue
			}

			if calcTime.Sub(updateTime) <= 15*time.Minute {
				localTimeStr := updateTime.In(localLoc).Format("15:04:05")
				players = append(players, playerActivity{
					Record:    record,
					LocalTime: localTimeStr,
				})
			}
		}

		if len(players) > 0 {
			activePlayersByArea[area.AreaName] = areaActivity{
				GMT:     gmtStr,
				Players: players,
			}
		}
	}

	if len(activePlayersByArea) == 0 {
		return detectionTime, nil // No active players, but still return when we checked/detected
	}

	// 5. Format pages (Sorted by Area Name)
	areaNames := make([]string, 0, len(activePlayersByArea))
	for name := range activePlayersByArea {
		areaNames = append(areaNames, name)
	}
	sort.Strings(areaNames)

	var pages []string
	var currentMessage strings.Builder
	var normalPlayerCount int
	var prideCount int
	header := "📡 **Active SEA OB Players (Live Update)**\n" +
		fmt.Sprintf("_Detected at: %s (JST)_\n", detectionTime) +
		fmt.Sprintf("_Refreshed every %d minutes_\n\n", s.Config.IntervalMinutes)
	currentMessage.WriteString(header)
	for _, areaName := range areaNames {
		activity := activePlayersByArea[areaName]
		var section strings.Builder
		section.WriteString(fmt.Sprintf("📍 **%s — (%s)**\n", areaName, activity.GMT))
		for _, p := range activity.Players {
			rankName := obRankingCfgMap[p.Record.OnlineBattleRankId].Name
			isPride := false
			if rankName == "" {
				rankName = obRankingCfgMap[p.Record.PrideId].Name
				isPride = true
			}
			if rankName != "" {
				if isPride {
					prideCount++
					section.WriteString(fmt.Sprintf("- `%s` — %s — (%d) — %s\n", p.Record.Name, rankName, p.Record.PridePoint, p.LocalTime))
				} else {
					normalPlayerCount++
					section.WriteString(fmt.Sprintf("- `%s` — %s %s — %s\n", p.Record.Name, rankName, p.Record.GetDisplayStarCount(), p.LocalTime))
				}
			} else {
				section.WriteString(fmt.Sprintf("- `%s` — %s\n", p.Record.Name, p.LocalTime))
			}
		}
		if currentMessage.Len()+section.Len() > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
		}
		currentMessage.WriteString("\n")
		currentMessage.WriteString(section.String())
	}
	// default add page for footer
	footer := "\n 📊 ***Player Counts***\n" +
		fmt.Sprintf("**PRIDE players: %d**\n", prideCount) +
		fmt.Sprintf("**Non-PRIDE players: %d**\n", normalPlayerCount)
	if currentMessage.Len()+len(footer) > 1900 {
		pages = append(pages, currentMessage.String())
		currentMessage.Reset()
	}
	currentMessage.WriteString(footer)

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	// Reconcile the channel to the freshly rendered pages (edit / send / delete).
	if err := s.Notifier.SyncPages(ctx, s.Config.ChannelID, botMessages, pages); err != nil {
		logger.Error(ctx, "failed to sync active player pages", err)
	}

	logger.Info(ctx, fmt.Sprintf("active player sync completed for %d areas", len(activePlayersByArea)))
	return detectionTime, nil
}
