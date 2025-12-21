package discord

import (
	"fmt"
	"strings"

	"McQueens_Tea_Cup/internal/domain"

	"github.com/bwmarrin/discordgo"
)

type DiscordNotifier struct {
	session   *discordgo.Session
	channelID string
}

func NewDiscordNotifier(s *discordgo.Session, channelID string) *DiscordNotifier {
	return &DiscordNotifier{
		session:   s,
		channelID: channelID, // Store channel ID here if your config has it
	}
}

func (d *DiscordNotifier) Close() error {
	return d.session.Close()
}

func (d *DiscordNotifier) Send(channelID string, item domain.Item, feedTitle string, transEN, transVN string) error {
	// Reconstruct the message body logic
	msgBody := fmt.Sprintf("**%s**\n\n", feedTitle)

	itemTitle := strings.TrimSpace(item.Title)
	contentBody := strings.TrimSpace(item.Content)

	if itemTitle != "" && !strings.Contains(contentBody, itemTitle) && len(itemTitle) < 100 {
		msgBody += fmt.Sprintf("**%s**\n", itemTitle)
	}

	msgBody += truncate(contentBody, 800)

	if transEN != "" {
		msgBody += fmt.Sprintf("\n\n🇬🇧 **Translation:**\n%s", truncate(transEN, 800))
	}
	if transVN != "" {
		msgBody += fmt.Sprintf("\n\n🇻🇳 **Bản dịch:**\n%s", truncate(transVN, 800))
	}

	msgBody += fmt.Sprintf("\n\n%s", item.Link)
	if item.ImageURL != "" {
		msgBody += "\n" + item.ImageURL
	}

	_, err := d.session.ChannelMessageSend(channelID, msgBody)
	return err
}

func truncate(text string, length int) string {
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length]) + "..."
}
