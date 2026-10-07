package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/pkg/logger"

	"github.com/bwmarrin/discordgo"
)

type MetaSyncService struct {
	Session   *discordgo.Session
	MetaLogic *MetaLogicService
	MetaCfg   config.MetaSyncConfig
}

func NewMetaSyncService(s *discordgo.Session, logic *MetaLogicService, cfg config.MetaSyncConfig) *MetaSyncService {
	return &MetaSyncService{
		Session:   s,
		MetaLogic: logic,
		MetaCfg:   cfg,
	}
}

func (s *MetaSyncService) Sync(ctx context.Context) (string, error) {
	if s.MetaCfg.ChannelID == "" {
		return "", fmt.Errorf("META_CHANNEL_ID not configured")
	}

	logger.Info(ctx, fmt.Sprintf("starting OBMeta sync to channel %s", s.MetaCfg.ChannelID))

	var pages []string
	var err error
	maxPollingDuration := 14 * time.Minute
	pollingInterval := 1 * time.Minute
	startTime := time.Now()
	detectionTime := ""

	// 1. Fetch existing state from Discord
	lastReportedTimeStr := ""
	lastDetectedHeaderPrefix := "_Detected at: "

	messages, err := s.Session.ChannelMessages(s.MetaCfg.ChannelID, 50, "", "", "")
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
						logger.Info(ctx, fmt.Sprintf("found existing state in Discord, last reported: %s", lastReportedTimeStr))
					}
				}
			}
		}
	}

	// Polling Phase
	jstLoc := time.FixedZone("JST", 9*60*60)
	for {
		// Stop polling after the max duration regardless of the data state. This
		// MUST live at the top of the loop (not nested inside the "data found"
		// checks below): during a Sega outage the response has no "Calculated at:"
		// header, so a nested break would never fire and the loop would spin every
		// minute until ctx is cancelled.
		if time.Since(startTime) > maxPollingDuration {
			logger.Warn(ctx, "max polling duration reached, using latest available data")
			detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
			break
		}

		// Get formatted pages (this also fetches from Sega internally)
		pages, err = s.MetaLogic.GetOBMetaPages(ctx, 1000, "all")
		if err != nil {
			return "", fmt.Errorf("failed to get meta pages: %w", err)
		}

		// Validation: check if data is fresh. Test the raw Index result for the
		// header before adding the prefix length — otherwise calcDateStart is 14
		// even when the header is absent, which both defeats the guard and risks a
		// slice-out-of-range on msg[calcDateStart:].
		if len(pages) > 0 {
			msg := pages[0]
			if headerIdx := strings.Index(msg, "Calculated at: "); headerIdx > -1 {
				calcDateStart := headerIdx + len("Calculated at: ")
				if calcDateEnd := strings.Index(msg[calcDateStart:], " (JST)"); calcDateEnd > -1 {
					calcDate := msg[calcDateStart : calcDateStart+calcDateEnd]

					// State-Aware Refresh Check
					isNewerThanDiscord := true
					if lastReportedTimeStr != "" {
						respTime, _ := time.ParseInLocation("2006/01/02 15:04:05", calcDate, jstLoc)
						discordTime, _ := time.ParseInLocation("2006/01/02 15:04:05", lastReportedTimeStr, jstLoc)
						isNewerThanDiscord = respTime.After(discordTime)
					}

					if s.MetaLogic.IsDataFresh(calcDate) && isNewerThanDiscord {
						logger.Info(ctx, fmt.Sprintf("data is fresh & newer (calcDate: %s), proceeding", calcDate))
						detectionTime = time.Now().In(jstLoc).Format("2006/01/02 15:04:05")
						break
					}

					if !isNewerThanDiscord {
						logger.Info(ctx, fmt.Sprintf("data (%s) already reported in Discord, waiting", calcDate))
					} else {
						logger.Warn(ctx, fmt.Sprintf("sega is late (calcDate: %s), polling again", calcDate))
					}
				}
			}
		}

		select {
		case <-time.After(pollingInterval):
			// continue loop
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// botMessages already fetched at beginning for state

	// Reverse botMessages to get them in chronological order (oldest first)
	for i, j := 0, len(botMessages)-1; i < j; i, j = i+1, j-1 {
		botMessages[i], botMessages[j] = botMessages[j], botMessages[i]
	}

	// 3. Edit existing messages or send new ones
	for i, page := range pages {
		if i < len(botMessages) {
			// Edit existing message
			_, err := s.Session.ChannelMessageEdit(s.MetaCfg.ChannelID, botMessages[i].ID, page)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("could not edit message %s", botMessages[i].ID), err)
				// Fallback: if edit fails, try sending a new one?
				// For now just log it.
			}
		} else {
			// Send new message
			_, err := s.Session.ChannelMessageSend(s.MetaCfg.ChannelID, page)
			if err != nil {
				logger.Error(ctx, "error sending meta page", err)
			}
		}
	}

	// 4. Delete leftover old messages if new pages are fewer than old messages
	if len(botMessages) > len(pages) {
		for i := len(pages); i < len(botMessages); i++ {
			err := s.Session.ChannelMessageDelete(s.MetaCfg.ChannelID, botMessages[i].ID)
			if err != nil {
				logger.Error(ctx, fmt.Sprintf("could not delete leftover message %s", botMessages[i].ID), err)
			}
		}
	}

	logger.Info(ctx, "OBMeta sync completed")
	return detectionTime, nil
}
