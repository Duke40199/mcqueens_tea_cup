package discord

import (
	"fmt"
	"time"

	idac_domain "McQueens_Tea_Cup/internal/domain/idac"

	"github.com/bwmarrin/discordgo"
)

// sendPagination sends a paginated message with Next/Prev buttons
func (h *Handler) SendPagination(i *discordgo.InteractionCreate, pages []string) {
	// If only 1 page, just send it without buttons
	if len(pages) == 1 {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &pages[0],
		})
		return
	}

	// Current Page Index
	pageIndex := 0

	// Helper to create buttons based on current page
	getComponents := func(current int) []discordgo.MessageComponent {
		prevDisabled := current == 0
		nextDisabled := current == len(pages)-1

		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "◀️ Previous",
						Style:    discordgo.PrimaryButton,
						CustomID: "pagination_prev",
						Disabled: prevDisabled,
					},
					discordgo.Button{
						Label:    fmt.Sprintf("Page %d/%d", current+1, len(pages)),
						Style:    discordgo.SecondaryButton,
						CustomID: "pagination_status",
						Disabled: true, // Just a label
					},
					discordgo.Button{
						Label:    "Next ▶️",
						Style:    discordgo.PrimaryButton,
						CustomID: "pagination_next",
						Disabled: nextDisabled,
					},
				},
			},
		}
	}

	// 1. Send the FIRST page
	components := getComponents(pageIndex)
	msg, err := h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &pages[0],
		Components: &components,
	})
	if err != nil {
		return
	}

	// 2. Register a Handler for Button Clicks
	// We use a closure so we can access 'pageIndex' and 'pages' safely
	// We also need a "stop" channel to kill the listener after timeout
	stop := make(chan struct{})

	// Create the handler function
	cleanup := h.Session.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		// Filter: Must be Button click, match Message ID, and match User ID (optional security)
		if ic.Type != discordgo.InteractionMessageComponent ||
			ic.Message.ID != msg.ID ||
			ic.Member.User.ID != i.Member.User.ID {
			return
		}

		// Handle Buttons
		switch ic.MessageComponentData().CustomID {
		case "pagination_prev":
			if pageIndex > 0 {
				pageIndex--
			}
		case "pagination_next":
			if pageIndex < len(pages)-1 {
				pageIndex++
			}
		}

		// Update the Message
		newComps := getComponents(pageIndex)
		s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    pages[pageIndex],
				Components: newComps,
			},
		})
	})
	// 3. Cleanup Routine (Timeout after 2 minutes)
	// We run this in a goroutine so we don't block
	go func() {
		select {
		case <-time.After(2 * time.Minute):
			// Remove buttons after timeout
			h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Components: &[]discordgo.MessageComponent{}, // Empty components clears them
			})
			cleanup() // Remove the event handler
		case <-stop:
			cleanup()
		}
	}()
}

func (h *Handler) HandleAutoComplete(i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if data.Name != "idac" {
		return
	}

	// Find the subcommand options
	// structure: idac -> [time-attack] -> [track, variant, area...]
	if len(data.Options) == 0 {
		return
	}
	subCmd := data.Options[0]

	var selectedTrack string

	// 1. Find what the user has currently selected for "track"
	for _, opt := range subCmd.Options {
		if opt.Name == "track" {
			selectedTrack = opt.StringValue()
		}
	}

	// 2. Generate choices for "variant"
	var choices []*discordgo.ApplicationCommandOptionChoice

	if variants, ok := idac_domain.TrackRegistry[selectedTrack]; ok {
		for _, v := range variants {
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  v.Name, // Display: "Downhill"
				Value: v.ID,   // Value passed to handler: "course-12"
			})
		}
	} else {
		// Default hint if no track selected yet
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  "Select a track first",
			Value: "none",
		})
	}

	// 3. Send choices back to Discord
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})

}
