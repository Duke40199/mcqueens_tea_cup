package discord

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"McQueens_Tea_Cup/internal/domain"

	"github.com/bwmarrin/discordgo"
)

type DiscordNotifier struct {
	session *discordgo.Session
}

func NewDiscordNotifier(token string) (*DiscordNotifier, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	if err := dg.Open(); err != nil {
		return nil, err
	}

	// Log user details
	if u, err := dg.User("@me"); err == nil {
		log.Printf("Connected to Discord as %s", u.Username)
	}

	return &DiscordNotifier{session: dg}, nil
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

// StartCommands registers the commands with Discord and sets up the listener
func (d *DiscordNotifier) StartCommands() error {
	// 1. Define Commands
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show current bot status",
		},
		{
			Name:        "nuhuh",
			Description: "Your opinion is in the trash",
		},
		{
			Name:        "mckween",
			Description: "desuwa",
		},
	}

	// 2. Define Handlers
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "I am running and monitoring your feeds! 📡",
				},
			})
		},
		"nuhuh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "https://c.tenor.com/SCfWfZvA8_0AAAAd/tenor.gif",
				},
			})
		},

		"mckween": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			resp, err := http.Get("https://c.tenor.com/9HXPPljXLxUAAAAC/tenor.gif")
			if err != nil {
				log.Printf("Error fetching wink gif: %v", err)
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Desuwa~",
					Files: []*discordgo.File{
						{
							Name:   "desuwa.gif",
							Reader: resp.Body,
						},
					},
				},
			})
		},
	}

	// 3. Register Handler to Router
	d.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// 4. Register Commands with Discord API (Bulk Overwrite)
	// Note: We use empty string "" for GuildID to register globally.
	// Global commands take up to 1 hour to update. For instant testing, put your Guild ID there.
	log.Println("Registering commands...")
	_, err := d.session.ApplicationCommandBulkOverwrite(d.session.State.User.ID, "", commands)
	return err
}
func truncate(text string, length int) string {
	runes := []rune(text)
	if len(runes) <= length {
		return text
	}
	return string(runes[:length]) + "..."
}
