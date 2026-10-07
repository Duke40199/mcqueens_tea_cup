package discord

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleAnonymousCommand(i *discordgo.InteractionCreate) {
	// 1. Extract the message they want to send anonymously.
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		return
	}
	messageContent := data.Options[0].StringValue()

	var discordID string
	if user := interactionUser(i); user != nil {
		discordID = user.ID
	}

	// 2. Acknowledge privately first (deferred + ephemeral). This keeps us inside
	// Discord's 3-second window and, crucially, lets us report the REAL outcome
	// after the DB write and channel post below. The previous code confirmed
	// success before doing any work and then tried to respond a second time on
	// error (rejected by Discord as "already acknowledged").
	if err := h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		log.Printf("cfs: failed to defer interaction: %v", err)
		return
	}

	// 3. Persist the confession (this also allocates the sequential id).
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	newID, err := h.CfsStateRepo.CreateCfsState(ctx, discordID, messageContent)
	if err != nil {
		log.Printf("cfs: failed to create cfs state: %v", err)
		h.editEphemeralReply(i, "⚠️ Something went wrong saving your confession. Please try again.")
		return
	}

	// 4. Post the confession to the channel as a standalone bot message.
	if _, err := h.Session.ChannelMessageSend(i.ChannelID, fmt.Sprintf("#cfs%04d: %s", newID, messageContent)); err != nil {
		log.Printf("cfs: failed to send confession message: %v", err)
		h.editEphemeralReply(i, "⚠️ Your confession was saved but couldn't be posted. Please contact an admin.")
		return
	}

	// 5. Only now confirm success to the author.
	h.editEphemeralReply(i, "Your confession has been sent in secret!")
}

// editEphemeralReply replaces the deferred ephemeral response with the given text.
func (h *Handler) editEphemeralReply(i *discordgo.InteractionCreate, content string) {
	if _, err := h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	}); err != nil {
		log.Printf("cfs: failed to edit response: %v", err)
	}
}
