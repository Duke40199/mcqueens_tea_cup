package discord

import (
	"fmt"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/service"
)

func (h *Handler) HandleOBRanking(cc *CommandContext) error {
	optMap := cc.OptMap

	// 1. Parse inputs / presentation defaults
	limit := 1000 // fetch all records by default if no limit specified
	if raw, ok := optMap["limit"]; ok {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return NewUserError("Limit must be a number")
		}
		limit = n
	}

	areaInput := strings.ToLower(optMap["area"])
	if areaInput == "" {
		areaInput = "all"
	}
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}
	finalArea := optMap["area"]
	areaName := finalArea
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}

	finalRound := optMap["round"]
	if finalRound == "" {
		finalRound = "all"
	}

	// 2. Orchestrate via the domain service
	view, err := h.OBRankingService.GetRanking(cc.Ctx, finalRound, finalArea, limit)
	if err != nil {
		return fmt.Errorf("fetching online battle rankings: %w", err)
	}
	if len(view.Entries) == 0 {
		return cc.Edit(fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName))
	}

	// 3. Render
	cc.SendPages(buildOBRankingPages(areaName, finalRound, view))
	return nil
}

// buildOBRankingPages formats a resolved OB ranking view into Discord message
// pages, capping each page at 10 entries or ~1900 characters.
func buildOBRankingPages(areaName, round string, view *entity.OBRankingView) []string {
	header := fmt.Sprintf("# Initial D Rankings (Online Battle)\n"+
		"### 🌎 : %s | Round: %s\n"+
		"### Calculated at: %s (JST time, local time coming soon.)\n\n", areaName, round, view.CalcDate)

	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	currentMessage.WriteString(header)

	for _, e := range view.Entries {
		var entry string
		if e.IsPride {
			entry = fmt.Sprintf("%s. **%s** — %s —`%s`\n", e.Rank, e.Name, e.RankName, e.Point)
		} else {
			entry = fmt.Sprintf("%s. **%s** — %s — %s —`%s`\n", e.Rank, e.Name, e.RankName, e.StarCount, e.Point)
		}

		// Split if 10 items OR length > 1900
		if itemsInChunk >= 10 || currentMessage.Len()+len(entry) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header) // Add header to every page for clarity
			itemsInChunk = 0
		}

		currentMessage.WriteString(entry)
		itemsInChunk++
	}

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	return pages
}

func (h *Handler) HandleOBMeta(cc *CommandContext) error {
	optMap := cc.OptMap

	// 1. Parse inputs / presentation defaults
	limit := 1000 // sample size: analyze top 1000 players by default
	if raw, ok := optMap["limit"]; ok {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return NewUserError("Limit must be a number")
		}
		limit = n
	}

	areaInput := strings.ToLower(optMap["area"])
	if areaInput == "" {
		areaInput = "all"
	}
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}
	finalArea := optMap["area"]
	areaName := finalArea
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}

	// 2. Orchestrate via the domain service
	view, err := h.OBMetaService.GetMeta(cc.Ctx, finalArea, limit)
	if err != nil {
		return fmt.Errorf("fetching online battle meta: %w", err)
	}
	if view.TotalSampled == 0 {
		return cc.Edit(fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName))
	}

	// 3. Render (shared with the OB meta sync cron)
	cc.SendPages(service.FormatOBMetaPages(view))
	return nil
}
