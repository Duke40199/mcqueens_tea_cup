package discord

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"McQueens_Tea_Cup/internal/domain/entity"
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
	case "track":
		query := focused.StringValue()
		choices = h.SearchTrackChoiceQuery(query)
	case "car":
		query := focused.StringValue()
		choices = h.SearchCarChoiceQuery(query)
	case "spec":
		choices = h.handleCarSpecAutocomplete(subCmd)
	case "area-select":
		query := focused.StringValue()
		choices = h.SearchAreaChoiceQuery(query)
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
	caserTitle := cases.Title(language.English)
	caserUpper := cases.Upper(language.English)
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, car := range h.CarChoices {
		if query != "" &&
			!strings.Contains(car.SearchBlob, query) {
			continue
		}
		carDisplayName := fmt.Sprintf("%s %s (%s)", caserTitle.String(car.Maker), car.Name, caserUpper.String((car.ModelCode)))
		choices = append(choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  carDisplayName,
				Value: strings.ToLower(fmt.Sprintf("%s %s (%s)", car.Maker, car.Name, car.ModelCode)),
			},
		)
		if len(choices) >= 25 {
			break
		}
	}
	return choices
}

func (h *Handler) SearchTrackChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	query = strings.ToLower(strings.TrimSpace(query))
	var choices []*discordgo.ApplicationCommandOptionChoice
	for trackName, courseID := range h.TrackChoices {
		if query != "" && !strings.Contains(strings.ToLower(trackName), query) {
			continue
		}
		choices = append(choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  trackName,
				Value: courseID,
			},
		)
	}
	sort.Slice(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})
	if len(choices) > 25 {
		choices = choices[:25]
	}
	return choices
}

func (h *Handler) SearchAreaChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	query = strings.ToLower(strings.TrimSpace(query))
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, area := range h.AreaMetadata {
		if query != "" && !strings.Contains(strings.ToLower(area.Name), query) {
			continue
		}
		choices = append(choices,
			&discordgo.ApplicationCommandOptionChoice{
				Name:  area.Name,
				Value: area.ALlNetCode,
			},
		)
	}
	sort.Slice(choices, func(i, j int) bool {
		return choices[i].Name < choices[j].Name
	})
	if len(choices) > 25 {
		choices = choices[:25]
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

func (h *Handler) handleCarSpecAutocomplete(subCmd *discordgo.ApplicationCommandInteractionDataOption) []*discordgo.ApplicationCommandOptionChoice {
	var (
		selectedCarName string
		query           string
	)
	for _, opt := range subCmd.Options {
		switch opt.Name {
		case "car":
			selectedCarName = opt.StringValue()
		case "spec":
			if opt.Focused {
				query = strings.ToLower(
					strings.TrimSpace(opt.StringValue()),
				)
			}
		}
	}
	if selectedCarName == "" {
		return []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  "Select a car first",
				Value: "none",
			},
		}
	}
	var choices []*discordgo.ApplicationCommandOptionChoice
	var foundCar *entity.CarMetadata
	for _, car := range h.CarChoices {
		if strings.Contains(selectedCarName, strings.ToLower(fmt.Sprintf("%s %s (%s)", car.Maker, car.Name, car.ModelCode))) {
			foundCar = car
			break
		}
	}
	if foundCar != nil {
		for idx, specName := range foundCar.SpecNames {
			if idx >= len(foundCar.SpecIDs) {
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
					Value: foundCar.SpecIDs[idx],
				},
			)
			if len(choices) >= 25 {
				break
			}
		}
	} else {
		return []*discordgo.ApplicationCommandOptionChoice{
			{
				Name:  "Car not found!",
				Value: "none",
			},
		}
	}

	return choices
}
