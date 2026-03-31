package discord

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
)

var (
	cfsCounter int
	cfsMutex   sync.Mutex
)

func init() {
	// Load initial counter state
	b, err := os.ReadFile("cfs_counter.txt")
	if err == nil {
		cfsCounter, _ = strconv.Atoi(string(bytes.TrimSpace(b)))
	}
}

func (h *Handler) HandleAnonymousCommand(i *discordgo.InteractionCreate) {

	// 1. Extract the message they want to send anonymously
	messageContent := i.ApplicationCommandData().Options[0].StringValue()
	var discordID string
	if i.Member != nil {
		// Command was used in a server (Guild)
		discordID = i.Member.User.ID
	} else if i.User != nil {
		// Command was used in a Direct Message
		discordID = i.User.ID
	}
	// 2. Respond to the interaction ephemerally
	// Because this is ephemeral, the prompt "User used /anon" is hidden from the public!
	err := h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Your confession has been sent in secret!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		// handle error
		return
	}
	// 3. Increment counter and format tag
	newID, err := h.cfsStateRepo.CreateCfsState(discordID, messageContent)
	if err != nil {
		h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error updating cfs state",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	// 4. Send a completely separate standard message to the channel.
	// This will just look like the bot is speaking on its own.
	_, err = h.Session.ChannelMessageSend(i.ChannelID, fmt.Sprintf("#cfs%04d: %s", newID, messageContent))
	if err != nil {
		h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error sending confession",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}
