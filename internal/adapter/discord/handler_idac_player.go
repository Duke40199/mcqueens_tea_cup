package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain/entity"
)

func (h *Handler) HandlePlayerInfo(cc *CommandContext) error {
	// 1. Validate & Parse Inputs
	playerInfo, hasPlayerInfo := cc.OptMap["user"]
	if !hasPlayerInfo {
		return NewUserError("Missing player info")
	}
	// 2. Resolve player credentials, then fetch the profile via the service
	playerName, playerArea, _, err := h.PlayerService.ResolvePlayer(cc.Ctx, playerInfo, "")
	if err != nil {
		return NewUserError("Cannot find player info.")
	}
	profile, err := h.PlayerService.GetPlayerProfile(cc.Ctx, playerName, playerArea)
	if err != nil {
		return fmt.Errorf("fetching player profile: %w", err)
	}
	if profile == nil {
		return NewUserError("Cannot find player grade.")
	}
	// 3. Build Response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"# **IGN:** %s | Area: %s\n"+
			"### **Account Grade**:\n"+
			"# %s%s\n"+
			"### **Online Battle Rank:**\n"+
			"# %s %s\n", playerName, "VN", profile.GradeName, profile.GradeNum, profile.OBRankName, profile.OBStarCount))
	finalContent := sb.String()
	embed := &discordgo.MessageEmbed{
		Title:       "__**Player Profile**__",
		Description: finalContent,
	}
	_, err = cc.Session.InteractionResponseEdit(cc.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

func (h *Handler) HandlePlayerTournamentInfo(cc *CommandContext) error {
	// 1. Validate & Parse Inputs
	areaInput, isInputArea := cc.OptMap["area"]
	if !isInputArea {
		return NewUserError("Area input is required.")
	}
	areaCode := entity.AreaAliases[areaInput]

	// 2. Orchestrate via the domain service
	view, err := h.PlayerService.GetTournamentInfo(cc.Ctx, areaCode)
	if err != nil {
		return fmt.Errorf("fetching tournament info: %w", err)
	}
	if len(view.Players) == 0 {
		return NewUserError("Player map not found")
	}

	// 3. Render
	areaName := entity.AreaDisplayNameByCode[areaInput]
	if areaName == "" {
		areaName = areaInput
	}
	cc.SendPages(buildTournamentInfoPages(areaName, view))
	return nil
}

// buildTournamentInfoPages formats a resolved tournament view into Discord
// message pages, capping each page at 10 players or ~1900 characters.
func buildTournamentInfoPages(areaName string, view *entity.TournamentInfoView) []string {
	header := fmt.Sprintf("## Player Info (Tournament Mode)\n"+
		"### 🌎 : %s | Round: %s\n"+"### Calculated at: %s.\n\n", areaName, "all", view.CalcDate)

	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	currentMessage.WriteString(header)

	playerCount := 0
	for _, player := range view.Players {
		playerCount++
		grade := player.Grade
		if grade == "" {
			grade = "N/A"
		}
		obRank := player.OBRank
		if obRank == "" {
			obRank = "N/A"
		}
		playerEntry := fmt.Sprintf("%d. **%s** — **%s%s** `(%d pts)` — **%s %s** `(%d pts)`\n",
			playerCount, player.Name, grade, player.GradeNum, player.GradeExp, obRank, player.OBRankNum, player.OBRankExp)

		// Split if 10 items OR length > 1900
		if itemsInChunk >= 10 || currentMessage.Len()+len(playerEntry) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header) // Add header to every page for clarity
			itemsInChunk = 0
		}
		currentMessage.WriteString(playerEntry)
		itemsInChunk++
	}
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	return pages
}
