package discord

import (
	"McQueens_Tea_Cup/internal/domain/entity"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleOBRanking(i *discordgo.InteractionCreate, optMap map[string]string) {
	var err error
	// 1. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}
	// 3. Parsing Inputs
	var limit int
	if _, ok := optMap["limit"]; ok {
		limit, err = strconv.Atoi(optMap["limit"])
		if err != nil {
			sendDeferredError("⚠️ Limit must be a number")
			return
		}
	} else {
		limit = 1000 // Arbitrary large number to fetch all records if no limit specified
	}
	areaInput := strings.ToLower(optMap["area"])
	if areaInput == "" {
		areaInput = "all"
	}
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}
	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	finalArea := optMap["area"]

	finalRound := optMap["round"]
	if finalRound == "" {
		finalRound = "all"
	}
	records, err := h.SegaClient.GetListOBRanking(finalRound, finalArea)

	if records == nil || len(records.Records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(records.Records) < limit {
		limit = len(records.Records)
	}
	// 4. Get Ranking cfgs
	obRankingCfgMap, err := h.OBRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		sendDeferredError("⚠️ Error fetching ranking configuration.")
		return
	}
	// 4. Build Pages (Slice of Strings)
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0

	// Pre-calculate Header
	header := fmt.Sprintf("# Initial D Rankings (Online Battle)\n"+
		"### 🌎 : %s | Round: %s\n"+
		"### Calculated at: %s (JST time, local time coming soon.)\n\n", areaName, finalRound, records.CalcDate)

	// Initialize first page with header
	currentMessage.WriteString(header)

	for j := 0; j < limit; j++ {
		r := records.Records[j]
		displayRank := obRankingCfgMap[r.OnlineBattleRankId].Name
		displayPoint := r.Point
		var entry string
		// if rank was found in cfg -> display rank name, star count, and points
		if displayRank != "" {
			entry = fmt.Sprintf("%s. **%s** — %s — %s —`%s`\n", r.Rank, r.Name, displayRank, r.GetDisplayStarCount(), displayPoint)
		}
		// if displayRank not found -> check whether player is a Pride player
		if displayRank == "" {
			displayRank = obRankingCfgMap[r.PrideId].Name
			displayPoint = strconv.Itoa(r.PridePoint)
			entry = fmt.Sprintf("%s. **%s** — %s —`%s`\n", r.Rank, r.Name, displayRank, displayPoint)
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

	// Append final page
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	// 5. Hand over to Pagination Helper
	h.SendPagination(i, pages)
}

func (h *Handler) HandleOBMeta(i *discordgo.InteractionCreate, optMap map[string]string) {
	var err error

	// 1. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}

	// 2. Parsing Inputs
	var limit int
	if val, ok := optMap["limit"]; ok {
		limit, err = strconv.Atoi(val)
		if err != nil {
			sendDeferredError("⚠️ Limit must be a number")
			return
		}
	} else {
		limit = 1000 // Sample size: Analyze top 1000 players by default
	}

	areaInput := strings.ToLower(optMap["area"])
	if areaInput == "" {
		areaInput = "all"
	}
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}
	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	finalArea := optMap["area"]

	currentRound, err := h.SegaClient.GetCurrentRound()
	if err != nil {
		sendDeferredError("⚠️ Error fetching current round.")
		return
	}

	// 3. Fetch Data
	records, err := h.SegaClient.GetListOBRanking(strconv.Itoa(currentRound), finalArea)
	if err != nil {
		sendDeferredError("⚠️ Error fetching list ob ranking.")
		return
	}
	if records == nil || len(records.Records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Online Battle)\n🌎 : %s\n\nNo records found.", areaName)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	// Adjust limit to not exceed actual record count
	if len(records.Records) < limit {
		limit = len(records.Records)
	}

	// ---------------------------------------------------------
	// 4. AGGREGATION LOGIC (New)
	// ---------------------------------------------------------

	// Map to store CarName -> Count
	metaMap := make(map[string]int)
	totalSampled := 0

	// Iterate only up to the limit (Sample Size)
	for j := 0; j < limit; j++ {
		r := records.Records[j]
		car := strings.TrimSpace(r.CarName) // Clean up potential whitespace
		if car == "" {
			car = "Unknown"
		}
		metaMap[car]++
		totalSampled++
	}

	// Struct for sorting
	type CarStat struct {
		Name  string
		Count int
	}
	var stats []CarStat
	for name, count := range metaMap {
		stats = append(stats, CarStat{Name: name, Count: count})
	}

	// Sort Descending by Count
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// ---------------------------------------------------------
	// 5. Build Pages (Output)
	// ---------------------------------------------------------

	var pages []string
	var currentMessage strings.Builder

	// Pre-calculate Header
	header := fmt.Sprintf("# Initial DAC Car Meta (Online Battle)\n### 🌎 : %s | Round: %d\n**Sample Size:** Top %d PRIDE Players\n\n", areaName, currentRound, totalSampled)
	currentMessage.WriteString(header)

	for idx, stat := range stats {
		// Calculate usage percentage
		percentage := (float64(stat.Count) / float64(totalSampled)) * 100

		// Format: 1. **FL5[DH]** — 150 Uses (15.0%)
		entry := fmt.Sprintf("%d. **%s** — %d Uses (`%.1f%%`)\n", idx+1, stat.Name, stat.Count, percentage)

		// Pagination Check (1900 chars safety limit)
		if currentMessage.Len()+len(entry) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
		}

		currentMessage.WriteString(entry)
	}

	// Append final page
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	// 6. Hand over to Pagination Helper
	h.SendPagination(i, pages)
}
