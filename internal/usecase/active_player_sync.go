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

	for {
		activePlayersByArea = make(map[string]areaActivity)
		freshDataFound := false

		// 4. Check each area
		jstLoc, _ := time.LoadLocation("Asia/Tokyo")
		for _, area := range areas {
			resp, err := s.SegaClient.GetListOBRanking(roundStr, area.AreaCode)
			if err != nil {
				log.Printf("⚠️ Error fetching ranking for %s (%s): %v", area.AreaName, area.AreaCode, err)
				continue
			}

			if resp == nil || len(resp.Records) == 0 {
				continue
			}

			// Validation: check if this data is fresh
			if s.MetaLogic.IsDataFresh(resp.CalcDate) {
				freshDataFound = true
				if detectionTime == "" {
					detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
				}
			} else {
				log.Printf("✅ ActivePlayerSyncService: Data is fresh (CalcDate: %s). Proceeding...", resp.CalcDate)
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

		if freshDataFound || time.Since(startTime) > maxPollingDuration {
			if !freshDataFound {
				log.Printf("❌ Max polling duration reached for Active Players. Results might be 15m late.")
				detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05") // Fallback detection time
			} else {
				log.Printf("✅ Fresh data found for Active Player Sync. Proceeding...")
			}
			break
		}

		log.Printf("⚠️ No fresh data found yet for any area. Polling again in %v...", pollingInterval)

		select {
		case <-time.After(pollingInterval):
			// continue loop
		case <-ctx.Done():
			return "", ctx.Err()
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
					section.WriteString(fmt.Sprintf("- `%s` — %s — %s\n", p.Record.Name, rankName, p.LocalTime))
				} else {
					section.WriteString(fmt.Sprintf("- `%s` — %s %s — %s\n", p.Record.Name, rankName, p.Record.GetDisplayStarCount(), p.LocalTime))
				}
			} else {
				section.WriteString(fmt.Sprintf("- `%s` — %s\n", p.Record.Name, p.LocalTime))
			}
		}
		section.WriteString("\n")

		if currentMessage.Len()+section.Len() > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
		}
		currentMessage.WriteString(section.String())
	}

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	// 5. Fetch existing bot messages to edit
	messages, err := s.Session.ChannelMessages(s.Config.ChannelID, 50, "", "", "")
	if err != nil {
		return "", fmt.Errorf("failed to fetch messages: %w", err)
	}

	var botMessages []*discordgo.Message
	for _, m := range messages {
		if m.Author.ID == s.Session.State.User.ID {
			botMessages = append(botMessages, m)
		}
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
