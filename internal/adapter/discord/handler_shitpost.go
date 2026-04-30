package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"McQueens_Tea_Cup/internal/domain/entity"

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
	var data entity.MieMeBell
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

func splitMessage(s string, maxLength int) []string {
	// 1. Fast path: if it fits, return immediately
	if len(s) <= maxLength {
		return []string{s}
	}

	var chunks []string
	runes := []rune(s) // Convert to runes to handle Emoji/Unicode correctly

	for len(runes) > 0 {
		// If remaining text fits, add it and we are done
		if len(runes) <= maxLength {
			chunks = append(chunks, string(runes))
			break
		}

		// 2. Determine the best cut point
		// Start assuming we have to do a hard cut at maxLength
		cut := maxLength

		// 3. Search backwards for a safe split point
		foundBestSplit := false

		// Priority A: Look for Newlines (\n)
		// We scan from the limit backwards. The first newline we find is the
		// "latest" possible paragraph break that fits in the message.
		for i := cut - 1; i >= 0; i-- {
			if runes[i] == '\n' {
				cut = i + 1 // Split *after* the newline
				foundBestSplit = true
				break
			}
		}

		// Priority B: Look for Spaces (if no newline found)
		if !foundBestSplit {
			for i := cut - 1; i >= 0; i-- {
				if runes[i] == ' ' {
					cut = i + 1 // Split *after* the space
					foundBestSplit = true
					break
				}
			}
		}

		// (If neither is found, we stick to the hard cut at maxLength
		// which handles giant strings of non-stop characters)

		// 4. Create the chunk
		chunk := string(runes[:cut])
		chunks = append(chunks, chunk)

		// 5. Advance the slice
		runes = runes[cut:]
	}

	return chunks
}
