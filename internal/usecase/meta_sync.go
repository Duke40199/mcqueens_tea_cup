package usecase

import (
	"context"
	"fmt"
	"log"

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

func (s *MetaSyncService) Sync(ctx context.Context) error {
	if s.MetaCfg.ChannelID == "" {
		return fmt.Errorf("META_CHANNEL_ID not configured")
	}

	log.Printf("📊 Starting OBMeta Sync to channel %s...", s.MetaCfg.ChannelID)

	// 1. Get formatted pages
	pages, err := s.MetaLogic.GetOBMetaPages(ctx, 1000, "all")
	if err != nil {
		return fmt.Errorf("failed to get meta pages: %w", err)
	}

	// 2. Fetch existing bot messages to edit
	messages, err := s.Session.ChannelMessages(s.MetaCfg.ChannelID, 50, "", "", "")
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
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
	return nil
}
