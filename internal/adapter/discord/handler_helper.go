package discord

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"McQueens_Tea_Cup/internal/domain/entity"
	idac_domain "McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/pkg/logger"
	"McQueens_Tea_Cup/pkg/tracer"
)

// SendPagination and its button handling now live in pagination.go.

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
		// Store-location commands need the AllNet area code.
		query := focused.StringValue()
		choices = h.SearchAreaChoiceQuery(query)
	case "country-select", "area", "country", "area1", "area2":
		// Ranking / TA / OB commands need Sega's area code (e.g. "area-57").
		query := focused.StringValue()
		choices = h.SearchCountryChoiceQuery(query)
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
		logger.Error(tracer.NewContext(context.Background()), "autocomplete response error", err)
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

// areaChoiceLabel renders an area for the autocomplete dropdown, suffixing the ISO
// country code (area_code) when available, e.g. "Vietnam (VNM)".
func areaChoiceLabel(area entity.IDACAreaMetadata) string {
	iso := strings.ToUpper(strings.TrimSpace(area.AreaCode))
	if iso == "" {
		return area.Name
	}
	return fmt.Sprintf("%s (%s)", area.Name, iso)
}

// areaMatchesQuery reports whether an area matches a (lowercased) autocomplete
// query by name or ISO code, so users can type either "viet" or "vnm".
func areaMatchesQuery(area entity.IDACAreaMetadata, query string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(area.Name), query) ||
		strings.Contains(strings.ToLower(area.AreaCode), query)
}

// maxAutocompleteChoices is Discord's hard limit on autocomplete results.
const maxAutocompleteChoices = 25

// areaSortRank orders areas for the dropdown: World/All first, then countries,
// then prefectures.
func areaSortRank(area entity.IDACAreaMetadata) int {
	switch {
	case area.SegaAreaCode == "area-all":
		return 0
	case strings.EqualFold(area.AreaType, "COUNTRY"):
		return 1
	default: // PREFECTURE (and anything else)
		return 2
	}
}

// sortedAreaMatches returns the areas matching query, ordered World first, then
// COUNTRY (A–Z), then PREFECTURE (A–Z), capped to Discord's choice limit.
func (h *Handler) sortedAreaMatches(query string) []entity.IDACAreaMetadata {
	query = strings.ToLower(strings.TrimSpace(query))
	var matches []entity.IDACAreaMetadata
	for _, area := range h.AreaMetadata {
		if areaMatchesQuery(area, query) {
			matches = append(matches, area)
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if ri, rj := areaSortRank(matches[i]), areaSortRank(matches[j]); ri != rj {
			return ri < rj
		}
		return strings.ToLower(matches[i].Name) < strings.ToLower(matches[j].Name)
	})
	if len(matches) > maxAutocompleteChoices {
		matches = matches[:maxAutocompleteChoices]
	}
	return matches
}

func (h *Handler) SearchAreaChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, area := range h.sortedAreaMatches(query) {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  areaChoiceLabel(area),
			Value: area.ALLNetCode,
		})
	}
	return choices
}

func (h *Handler) SearchCountryChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, area := range h.sortedAreaMatches(query) {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  areaChoiceLabel(area),
			Value: area.SegaAreaCode,
		})
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
