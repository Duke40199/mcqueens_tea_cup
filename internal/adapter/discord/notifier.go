package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/pkg/logger"
)

// DiscordNotifier implements port.Notifier over a discordgo session. It keeps all
// discordgo-specific message plumbing out of the domain services.
type DiscordNotifier struct {
	session *discordgo.Session
}

func NewDiscordNotifier(s *discordgo.Session) port.Notifier {
	return &DiscordNotifier{session: s}
}

// BotMessages fetches up to limit recent messages in the channel and keeps only
// those authored by the bot, preserving Discord's newest-first order.
func (n *DiscordNotifier) BotMessages(_ context.Context, channelID string, limit int) ([]entity.ChannelMessage, error) {
	messages, err := n.session.ChannelMessages(channelID, limit, "", "", "")
	if err != nil {
		return nil, err
	}

	botID := ""
	if n.session.State != nil && n.session.State.User != nil {
		botID = n.session.State.User.ID
	}

	var out []entity.ChannelMessage
	for _, m := range messages {
		if m.Author != nil && m.Author.ID == botID {
			out = append(out, entity.ChannelMessage{ID: m.ID, Content: m.Content})
		}
	}
	return out, nil
}

// SyncPages reconciles the channel to exactly `pages`. `existing` is the slice
// returned by BotMessages (newest first); it is reversed to chronological order so
// page 1 maps to the oldest message. Per-message failures are logged but don't
// abort the whole sync. This logic previously lived (duplicated) inside both sync
// services.
func (n *DiscordNotifier) SyncPages(ctx context.Context, channelID string, existing []entity.ChannelMessage, pages []string) error {
	ordered := make([]entity.ChannelMessage, len(existing))
	for i, m := range existing {
		ordered[len(existing)-1-i] = m
	}

	// Edit existing messages in place; send new ones for any extra pages.
	for i, page := range pages {
		if i < len(ordered) {
			if _, err := n.session.ChannelMessageEdit(channelID, ordered[i].ID, page); err != nil {
				logger.Error(ctx, fmt.Sprintf("could not edit message %s", ordered[i].ID), err)
			}
		} else {
			if _, err := n.session.ChannelMessageSend(channelID, page); err != nil {
				logger.Error(ctx, "error sending page", err)
			}
		}
	}

	// Delete leftovers when there are fewer pages than existing messages.
	for i := len(pages); i < len(ordered); i++ {
		if err := n.session.ChannelMessageDelete(channelID, ordered[i].ID); err != nil {
			logger.Error(ctx, fmt.Sprintf("could not delete leftover message %s", ordered[i].ID), err)
		}
	}
	return nil
}
