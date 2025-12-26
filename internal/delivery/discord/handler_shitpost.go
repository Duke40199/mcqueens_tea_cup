package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"McQueens_Tea_Cup/internal/domain"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleMckween(i *discordgo.InteractionCreate) {
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Desuwa~",
			Files: []*discordgo.File{
				{Name: "desuwa.gif", Reader: bytes.NewReader(desuwaGif)},
			},
		},
	})
}

func (h *Handler) HandleMiemebell(i *discordgo.InteractionCreate) {
	var data domain.MieMeBell
	// Assuming miemebellJson is accessible here (package level or struct field)
	if err := json.Unmarshal(miemebellJson, &data); err != nil {
		log.Println("Error parsing JSON:", err)
		return
	}
	// Use h.Rand if you want to mock random in tests, otherwise global rand is fine for simple bots
	randomIndex := rand.Intn(len(data.Blocks))

	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: data.Blocks[randomIndex],
		},
	})
}

func (h *Handler) HandleNuhuh(i *discordgo.InteractionCreate) {
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: "Brewing tea..."},
	})

	options := i.ApplicationCommandData().Options
	var content string
	for _, opt := range options {
		if opt.Name == "user" {
			targetID := opt.Value.(string)
			if targetID == h.Session.State.User.ID {
				content = "Nice try, but ***nuh-uh***"
			} else {
				content = fmt.Sprintf("<@%s>", targetID)
			}
		}
	}

	var contentPtr *string
	if content != "" {
		contentPtr = &content
	}

	h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: contentPtr,
		Files: []*discordgo.File{
			{Name: "nuhuh.gif", Reader: bytes.NewReader(nuhuhGif)},
		},
	})
}
