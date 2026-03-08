package usecase

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

const segaTimeLayout = "2006/01/02 15:04:05"

type ActivePlayerSyncService struct {
	Session          *discordgo.Session
	SegaClient       entity.SegaClient
	AreaRepo         postgres.AreaRepository
	OBRankingCfgRepo postgres.OBRankingCfgRepository
	MetaLogic        *MetaLogicService
	Config           config.ActivePlayersSyncConfig
}

func NewActivePlayerSyncService(s *discordgo.Session, client entity.SegaClient, areaRepo postgres.AreaRepository, obRepo postgres.OBRankingCfgRepository, logic *MetaLogicService, cfg config.ActivePlayersSyncConfig) *ActivePlayerSyncService {
	return &ActivePlayerSyncService{
		Session:          s,
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
		log.Println("⚠️ No areas found for active player sync")
		return "", nil
	}

	// 2. Get rank configs
	obRankingCfgMap, err := s.OBRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		return "", fmt.Errorf("failed to get ranking cfg map: %w", err)
	}

	// 3. Get current round
	currentRound, err := s.SegaClient.GetCurrentRound()
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

	messages, err := s.Session.ChannelMessages(s.Config.ChannelID, 50, "", "", "")
	var botMessages []*discordgo.Message
	if err == nil {
		for _, m := range messages {
			if m.Author.ID == s.Session.State.User.ID {
				botMessages = append(botMessages, m)
				if lastReportedTimeStr == "" && strings.Contains(m.Content, lastDetectedHeaderPrefix) {
					start := strings.Index(m.Content, lastDetectedHeaderPrefix) + len(lastDetectedHeaderPrefix)
					end := strings.Index(m.Content[start:], " (JST)")
					if end > -1 {
						lastReportedTimeStr = m.Content[start : start+end]
						log.Printf("📥 Found existing state in Discord. Last reported: %s", lastReportedTimeStr)
					}
				}
			}
		}
	}

	// 5. Polling Phase (Canary Check)
	jstLoc, _ := time.LoadLocation("Asia/Tokyo")
	canaryArea := areas[0]

	for {
		resp, err := s.SegaClient.GetListOBRanking(roundStr, canaryArea.AreaCode)
		if err != nil {
			log.Printf("⚠️ Polling Error for Canary %s: %v", canaryArea.AreaName, err)
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
				log.Printf("✅ Fresh & Newer data detected via Canary (%s: %s). Proceeding to full sync...", canaryArea.AreaName, resp.CalcDate)
				detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
				break
			}

			if !isNewerThanDiscord {
				log.Printf("😴 Sega Data (%s) is already reported in Discord (%s). Waiting for next block...", resp.CalcDate, lastReportedTimeStr)
			} else {
				log.Printf("⚠️ Sega is late (Canary %s: %s). Polling again in %v...", canaryArea.AreaName, resp.CalcDate, pollingInterval)
			}
		}

		if time.Since(startTime) > maxPollingDuration {
			log.Printf("❌ Max polling duration reached. Sega is significantly late. Using latest available data.")
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
		resp, err := s.SegaClient.GetListOBRanking(roundStr, area.AreaCode)
		if err != nil {
			log.Printf("⚠️ Error fetching ranking for %s (%s): %v", area.AreaName, area.AreaCode, err)
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
		fmt.Sprintf("_Refreshed every %d minutes_\n\n", s.Config.Interval)
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
					section.WriteString(fmt.Sprintf("- `%s` — %s — %s\n", p.Record.Name, rankName, p.LocalTime))
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
			currentMessage.WriteString(header)
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
		currentMessage.WriteString(header)
	}
	currentMessage.WriteString(footer)

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	// Reverse to get chronological order (oldest first)
	for i, j := 0, len(botMessages)-1; i < j; i, j = i+1, j-1 {
		botMessages[i], botMessages[j] = botMessages[j], botMessages[i]
	}

	// 6. Edit or Send
	for i, page := range pages {
		if i < len(botMessages) {
			_, err := s.Session.ChannelMessageEdit(s.Config.ChannelID, botMessages[i].ID, page)
			if err != nil {
				log.Printf("⚠️ Warning: could not edit message %s: %v", botMessages[i].ID, err)
			}
		} else {
			_, err := s.Session.ChannelMessageSend(s.Config.ChannelID, page)
			if err != nil {
				log.Printf("❌ Error sending active players page: %v", err)
			}
		}
	}

	// 7. Prune leftover
	if len(botMessages) > len(pages) {
		for i := len(pages); i < len(botMessages); i++ {
			err := s.Session.ChannelMessageDelete(s.Config.ChannelID, botMessages[i].ID)
			if err != nil {
				log.Printf("⚠️ Warning: could not delete leftover message %s: %v", botMessages[i].ID, err)
			}
		}
	}

	log.Printf("✅ Active Player Sync Completed for %d areas", len(activePlayersByArea))
	return detectionTime, nil
}
