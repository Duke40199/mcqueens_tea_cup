package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/config"

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

	log.Printf("📊 Starting OBMeta Sync to channel %s...", s.MetaCfg.ChannelID)

	var pages []string
	var err error
	maxPollingDuration := 14 * time.Minute
	pollingInterval := 1 * time.Minute
	startTime := time.Now()
	detectionTime := ""

	for {
		// 1. Get formatted pages (this also fetches from Sega internally)
		pages, err = s.MetaLogic.GetOBMetaPages(ctx, 1000, "all")
		if err != nil {
			return "", fmt.Errorf("failed to get meta pages: %w", err)
		}

		// Validation: check if data is fresh
		if len(pages) > 0 {
			msg := pages[0]
			calcDateStart := strings.Index(msg, "Calculated at: ") + len("Calculated at: ")
			calcDateEnd := strings.Index(msg[calcDateStart:], " (JST)")
			if calcDateStart > -1 && calcDateEnd > -1 {
				calcDate := msg[calcDateStart : calcDateStart+calcDateEnd]
				if s.MetaLogic.IsDataFresh(calcDate) {
					log.Printf("✅ Data is fresh (CalcDate: %s). Proceeding...", calcDate)
					detectionTime = time.Now().In(time.FixedZone("JST", 9*60*60)).Format("2006/01/02 15:04:05")
					break
				}

				if time.Since(startTime) > maxPollingDuration {
					log.Printf("❌ Max polling duration reached. Sega is significantly late. Using latest available data.")
					detectionTime = time.Now().In(time.FixedZone("JST", 9*60*60)).Format("2006/01/02 15:04:05")
					break
				}

				log.Printf("⚠️ Sega is late (CalcDate: %s). Polling again in %v...", calcDate, pollingInterval)
			}
		}

		select {
		case <-time.After(pollingInterval):
			// continue loop
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}

	// 2. Fetch existing bot messages to edit
	messages, err := s.Session.ChannelMessages(s.MetaCfg.ChannelID, 50, "", "", "")
	if err != nil {
		return "", fmt.Errorf("failed to fetch messages: %w", err)
	}

	var botMessages []*discordgo.Message
	for _, m := range messages {
		if m.Author.ID == s.Session.State.User.ID {
			botMessages = append(botMessages, m)
		}
	}

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
				log.Printf("⚠️ Warning: could not edit message %s: %v", botMessages[i].ID, err)
				// Fallback: if edit fails, try sending a new one?
				// For now just log it.
			}
		} else {
			// Send new message
			_, err := s.Session.ChannelMessageSend(s.MetaCfg.ChannelID, page)
			if err != nil {
				log.Printf("❌ Error sending meta page: %v", err)
			}
		}
	}

	// 4. Delete leftover old messages if new pages are fewer than old messages
	if len(botMessages) > len(pages) {
		for i := len(pages); i < len(botMessages); i++ {
			err := s.Session.ChannelMessageDelete(s.MetaCfg.ChannelID, botMessages[i].ID)
			if err != nil {
				log.Printf("⚠️ Warning: could not delete leftover message %s: %v", botMessages[i].ID, err)
			}
		}
	}

	log.Printf("✅ OBMeta Sync Completed")
	return detectionTime, nil
}
