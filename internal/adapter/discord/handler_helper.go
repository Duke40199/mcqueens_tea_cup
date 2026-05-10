package discord

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	idac_domain "McQueens_Tea_Cup/internal/domain/entity"
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

func (h *Handler) HandleAutoComplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if data.Name != string(CommandNameIDAC) {
		return
	}
	if len(data.Options) == 0 {
		return
	}
	subCmd := data.Options[0]
	var focused *discordgo.ApplicationCommandInteractionDataOption
	// Find focused option
	for _, opt := range subCmd.Options {
		if opt.Focused {
			focused = opt
			break
		}
	}
	if focused == nil {
		return
	}
	var choices []*discordgo.ApplicationCommandOptionChoice
	switch focused.Name {
	case "variant":
		choices = h.handleVariantAutocomplete(subCmd)
	case "car":
		query := focused.StringValue()
		choices = h.SearchCarChoiceQuery(query)
	case "spec":
		choices = h.handleCarSpecAutocomplete(subCmd)
	}
	err := s.InteractionRespond(
		i.Interaction,
		&discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		},
	)

	if err != nil {
		log.Println("autocomplete response error:", err)
	}
}

func (h *Handler) SearchCarChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	query = strings.ToLower(strings.TrimSpace(query))
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, car := range h.CarChoices {
		if query != "" &&
			!strings.Contains(car.SearchBlob, query) {
			continue
		}
		choices = append(choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  fmt.Sprintf("%s %s (%s)", car.Maker, car.Name, car.ModelCode),
				Value: strconv.Itoa(int(car.SegaCarID)),
			},
		)

		if len(choices) >= 25 {
			break
		}
	}
	return choices
}

func (h *Handler) handleVariantAutocomplete(subCmd *discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandOptionChoice {
	var selectedTrack string
	for _, opt := range subCmd.Options {
		if opt.Name == "track" {
			selectedTrack = opt.StringValue()
		}
	}
	var choices []*discordgo.ApplicationCommandOptionChoice
	if variants, ok := idac_domain.TrackRegistry[selectedTrack]; ok {
		for _, v := range variants {
			choices = append(choices,
				&discordgo.ApplicationCommandOptionChoice{
					Name:  v.Name,
					Value: v.ID,
				},
			)
		}
	} else {
		choices = append(choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  "Select a track first",
				Value: "none",
			},
		)
	}
	return choices
}

func (h *Handler) handleCarSpecAutocomplete(
	subCmd *discordgo.ApplicationCommandInteractionDataOption,
) []*discordgo.ApplicationCommandOptionChoice {

	var (
		selectedCarID string
		query         string
	)
	for _, opt := range subCmd.Options {
		switch opt.Name {
		case "car":
			selectedCarID = opt.StringValue()
		case "spec":
			if opt.Focused {
				query = strings.ToLower(
					strings.TrimSpace(opt.StringValue()),
				)
			}
		}
	}
	if selectedCarID == "" {
		return []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  "Select a car first",
				Value: "none",
			},
		}
	}
	carID, err := strconv.ParseInt(selectedCarID, 10, 64)
	if err != nil {
		return nil
	}
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, car := range h.CarChoices {
		if car.SegaCarID != carID {
			continue
		}
		for idx, specName := range car.SpecNames {
			if idx >= len(car.SpecIDs) {
				continue
			}
			if query != "" &&
				!strings.Contains(
					strings.ToLower(specName),
					query,
				) {
				continue
			}
			choices = append(choices,
				&discordgo.ApplicationCommandOptionChoice{
					Name:  specName,
					Value: car.SpecIDs[idx],
				},
			)
			if len(choices) >= 25 {
				break
			}
		}
		break
	}
	return choices
}
