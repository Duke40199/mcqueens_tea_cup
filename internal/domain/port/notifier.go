package port

import (
	"context"

	"McQueens_Tea_Cup/internal/domain/entity"
)

// Notifier abstracts the chat transport (Discord) that the sync services publish
// to, so the domain layer never imports discordgo directly.
type Notifier interface {
	// BotMessages returns up to limit of the bot's own recent messages in the
	// channel, newest first.
	BotMessages(ctx context.Context, channelID string, limit int) ([]entity.ChannelMessage, error)

	// SyncPages reconciles the channel's bot messages to exactly `pages`: the
	// `existing` messages (as returned by BotMessages) are edited in chronological
	// order, extra pages are sent as new messages, and any leftover messages are
	// deleted.
	SyncPages(ctx context.Context, channelID string, existing []entity.ChannelMessage, pages []string) error
}
