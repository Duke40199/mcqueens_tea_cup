package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
)

type MetaSyncService struct {
	Notifier  port.Notifier
	MetaLogic *MetaLogicService
	MetaCfg   config.MetaSyncConfig
}

func NewMetaSyncService(notifier port.Notifier, logic *MetaLogicService, cfg config.MetaSyncConfig) *MetaSyncService {
	return &MetaSyncService{
		Notifier:  notifier,
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

	botMessages, err := s.Notifier.BotMessages(ctx, s.MetaCfg.ChannelID, 50)
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

	// Reconcile the channel to the freshly rendered pages (edit / send / delete).
	if err := s.Notifier.SyncPages(ctx, s.MetaCfg.ChannelID, botMessages, pages); err != nil {
		logger.Error(ctx, "failed to sync meta pages", err)
	}

	logger.Info(ctx, "OBMeta sync completed")
	return detectionTime, nil
}
