package discord

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/domain/entity"

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
	// 4. FETCH BASE SPEC MAP FROM DB
	// ---------------------------------------------------------
	ctx := context.Background()
	baseSpecMap, err := h.CarRepo.GetBaseSpecMap(ctx)
	if err != nil {
		sendDeferredError("⚠️ Error fetching car spec data from database.")
		return
	}
	// ---------------------------------------------------------
	// 5. AGGREGATION LOGIC: Group by BaseSpec -> Style -> Car
	// ---------------------------------------------------------

	// CarStat tracks count per car within a style group
	type CarStat struct {
		Name  string
		Count int
	}

	// specStyleMap: map[baseSpec]map[style]map[modelCode]*CarStat
	specStyleMap := make(map[string]map[string]map[string]*CarStat)
	totalSampled := 0

	for j := 0; j < limit; j++ {
		r := records.Records[j]
		carName := strings.TrimSpace(r.CarName)
		if carName == "" {
			continue
		}
		// Parse "FL5[DH]" -> modelCode="FL5", style="DH"
		bracketStart := strings.Index(carName, "[")
		bracketEnd := strings.Index(carName, "]")

		var modelCode, style string
		if bracketStart != -1 && bracketEnd != -1 && bracketEnd > bracketStart {
			modelCode = strings.TrimSpace(carName[:bracketStart])
			style = strings.ToUpper(strings.TrimSpace(carName[bracketStart+1 : bracketEnd]))
		} else {
			modelCode = carName
			style = "Unknown"
		}

		// Look up spec info from DB (resolves aliases to canonical model code)
		specInfo, ok := baseSpecMap[modelCode]
		if ok {
			modelCode = specInfo.ModelCode // Normalize: "CZ4Aエボ10" -> "CZ4A"
		}

		baseSpec := "Unknown"
		if ok {
			baseSpec = specInfo.BaseSpec
		} else {
			fmt.Println("Unknown base spec: " + modelCode)
		}

		// Initialize nested maps as needed
		if specStyleMap[baseSpec] == nil {
			specStyleMap[baseSpec] = make(map[string]map[string]*CarStat)
		}
		if specStyleMap[baseSpec][style] == nil {
			specStyleMap[baseSpec][style] = make(map[string]*CarStat)
		}
		if specStyleMap[baseSpec][style][modelCode] == nil {
			specStyleMap[baseSpec][style][modelCode] = &CarStat{Name: getCarStatName(&specInfo, modelCode), Count: 0}
		}
		specStyleMap[baseSpec][style][modelCode].Count++
		totalSampled++
	}
	// ---------------------------------------------------------
	// 6. Build Pages (Output)
	// ---------------------------------------------------------
	var pages []string
	var currentMessage strings.Builder

	header := fmt.Sprintf(
		"# Initial DAC Most Used Cars (Online Battle)\n"+
			"### **Round: %d | Sample Size:** Top %d PRIDE Players\n"+
			"### Calculated at: %s (JST)\n"+
			"### ***(p.s. results are refreshed every 15 minutes.)***\n\n",
		currentRound, totalSampled, records.CalcDate)
	currentMessage.WriteString(header)

	// Sort base specs for consistent output
	specKeys := make([]string, 0, len(specStyleMap))
	for k := range specStyleMap {
		specKeys = append(specKeys, k)
	}
	sort.Strings(specKeys)

	for _, baseSpec := range specKeys {
		styleMap := specStyleMap[baseSpec]
		specEmoji := entity.SpecEmojis[baseSpec]
		if specEmoji == "" {
			specEmoji = "📦"
		}
		var section string
		// Sort styles for consistent output (AR, DH, HC)
		styleKeys := make([]string, 0, len(styleMap))
		for k := range styleMap {
			styleKeys = append(styleKeys, k)
		}
		sort.Strings(styleKeys)

		for _, style := range styleKeys {
			styleEmoji := entity.SpecEmojis[strings.ToLower(style)]
			carMap := styleMap[style]
			// Collect and sort cars by count descending
			cars := make([]CarStat, 0, len(carMap))
			for _, cs := range carMap {
				cars = append(cars, *cs)
			}
			sort.Slice(cars, func(i, j int) bool {
				return cars[i].Count > cars[j].Count
			})

			// Build car list (Limit to top 5)
			carParts := make([]string, 0, 5)
			for idx, c := range cars {
				if idx >= 5 {
					break
				}
				carParts = append(carParts, fmt.Sprintf("- %s (%d)", c.Name, c.Count))
			}
			section += fmt.Sprintf("%s %s:\n%s\n", specEmoji, styleEmoji, strings.Join(carParts, "\n"))
		}
		section += "\n"

		// Pagination check
		if currentMessage.Len()+len(section) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
		}
		currentMessage.WriteString(section)
	}

	// Append final page
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	// 7. Hand over to Pagination Helper
	h.SendPagination(i, pages)
}

func getCarStatName(carStat *entity.CarSpecInfo, defaultCode string) string {
	if carStat.Maker == "" || carStat.CarName == "" {
		return defaultCode
	}
	return fmt.Sprintf("%s %s [%s]", strings.Title(carStat.Maker), strings.Title(carStat.CarName), strings.Title(carStat.ModelCode))
}
