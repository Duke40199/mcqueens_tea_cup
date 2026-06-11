package discord

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleTimeAttack(i *discordgo.InteractionCreate, optMap map[string]string, specInput string) {
	// 1. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	// 2. Validate input
	inputTrack := optMap["track"]
	//inputTrackVariant := optMap["variant"]
	inputArea := strings.ToLower(optMap["area"])
	//if inputTrackVariant == "" {
	//	h.SendDeferredError(i, "⚠️ **Please select a track variant.**")
	//	return
	//}
	finalCourseID := inputTrack
	courseName := entity.CourseDisplayNameByCode[finalCourseID]
	if courseName == "" {
		h.SendDeferredError(i, "⚠️ **Track not found based on input.**")
		return
	}
	// 2.b. Car Input
	var finalCarID string
	var foundCar *entity.CarMetadata
	isInputCar := optMap["car"] != "" && optMap["car"] != "all"
	if !isInputCar {
		finalCarID = "car-all"
	}
	if isInputCar {
		if specInput == "" {
			h.SendDeferredError(i, "⚠️ **Please select a car spec.**")
			return
		}
		for _, carChoice := range h.CarChoices {
			for _, specID := range carChoice.SpecIDs {
				if specInput == specID {
					foundCar = carChoice
					finalCarID = "car-" + specInput
					break
				}
			}
		}
		if foundCar == nil {
			h.SendDeferredError(i, "⚠️ Car not found with input.")
			return
		}
		if inputArea != "" && inputArea != "all" {
			h.SendDeferredError(i, "ℹ️ Currently search with car only supported with __**all**__ area input.")
			return
		}
	}
	// Get Car Display Name from Input
	var carDisplayName string
	if !isInputCar {
		carDisplayName = "All"
	} else {
		carDisplayName = foundCar.Name
		// if emoji, ok := entity.SpecEmojis[strings.ToLower(specInput)]; ok {
		// 	carDisplayName = fmt.Sprintf("%s %s", emoji, val)
		// } else {
		// 	carDisplayName = val
		// }
	}
	// 2.d Resolve Player Area
	if val, ok := entity.AreaAliases[inputArea]; ok {
		optMap["area"] = val
	}
	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}

	var resultLimit = 1000
	// 3. Fetch Data
	finalArea := optMap["area"]
	records, err := h.SegaClient.GetListTimeTrail(finalCourseID, finalArea, finalCarID, specInput)
	if err != nil {
		h.SendDeferredError(i, "⚠️ Failed to fetch data from Sega API: "+err.Error())
		return
	}
	if len(records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\nNo records found.", courseName, areaName, carDisplayName)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(records) < resultLimit {
		resultLimit = len(records)
	}
	taTimeMetadata, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(context.Background(), finalCourseID)
	if err != nil {
		h.SendDeferredError(i, "⚠️ Failed to metadata TA time, error log:"+err.Error())
		return
	}
	// 4. Build Pages (Slice of Strings)
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0

	// Pre-calculate Header
	header := fmt.Sprintf("# Initial D Rankings (Time Trial)\n 🗾 : %s | 🌎 : %s | 🚗 : %s\n", courseName, areaName, carDisplayName)

	// Initialize first page with header
	currentMessage.WriteString(header)
	listCarPercentages, err := h.GetListTACarsPercentage(i, finalCourseID, optMap)
	if err != nil {
		h.SendDeferredError(i, "⚠️ Failed to get top three TA cars: "+err.Error())
		return
	}
	// Initialize car percentage header
	headerCarPercentage := "## Top 3 Most Used Cars (based on top 1000 **Global** results)\n"
	// car name sega format: FD3S[DH]
	listCarNameSegaFormat := make([]string, 0)
	for i := 0; i < 3; i++ {
		listCarNameSegaFormat = append(listCarNameSegaFormat, listCarPercentages[i].SegaCarName)
	}
	carListFullInfo, err := h.GetListCarDetailByTAFormat(context.TODO(), listCarNameSegaFormat)
	if err != nil {
		h.SendDeferredError(i, "⚠️ Failed to GetListCarDetailByTAFormat: "+err.Error())
		return
	}
	var carCount = 0
	for z := 0; z < 3; z++ {
		splitSegaCarName := strings.Split(listCarNameSegaFormat[z], "[")
		// if not found by chassis code -> find by aliases
		var foundCarFullInfo entity.CarSpecInfo
		for _, carFullInfo := range carListFullInfo {
			if slices.Contains(carFullInfo.Aliases, splitSegaCarName[0]) || splitSegaCarName[0] == carFullInfo.ModelCode {
				foundCarFullInfo = carFullInfo
				continue
			}
		}

		entry := fmt.Sprintf("%d. %s %s **%s %s (%s)** - `%.1f%%`\n", carCount+1,
			entity.SpecEmojis[strings.ToLower(foundCarFullInfo.BaseSpec)],
			entity.SpecEmojis[strings.ToLower(foundCarFullInfo.SpecStyleName)],
			strings.ToTitle(foundCarFullInfo.Maker),
			foundCarFullInfo.CarName,
			foundCarFullInfo.ModelCode,
			listCarPercentages[z].Percentage)
		headerCarPercentage += entry
		carCount++
	}
	headerCarPercentage += "\n"
	currentMessage.WriteString(headerCarPercentage)

	for j := 0; j < resultLimit; j++ {
		r := records[j]
		timeRecord, _ := entity.ParseRaceTime(r.Record)
		recordRank := "NO DATA"
		if taTimeMetadata != nil {
			recordRank, _ = h.GetPlayerTimeAttackRank(timeRecord, finalCourseID, taTimeMetadata)
		}
		entry := fmt.Sprintf("%s. **%s**  — `%s` — **%s** — `%s`\n", r.Rank, recordRank, r.Record, r.Name, r.CarName)

		// Split if 10 items OR length > 1900
		if itemsInChunk >= 10 || currentMessage.Len()+len(entry) > 2000 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
			currentMessage.WriteString(headerCarPercentage)
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

func (h *Handler) GetListTACarsPercentage(i *discordgo.InteractionCreate, segaCourseID string, optMap map[string]string) ([]entity.IDACCarUsagePercentage, error) {
	// 1. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}
	// 2. Validate Inputs
	// We look for 'variant' because that holds the actual Course ID now
	if segaCourseID == "" {
		sendDeferredError("⚠️ **GetListTACarsPercentage: Invalid course input.")
		return nil, nil
	}
	// Defaults
	if _, ok := optMap["car"]; !ok {
		optMap["car"] = "car-all"
	}
	// The User selected: Track="Akina", Variant="course-12" (Downhill)
	// We only care about the Variant ID now.

	// Resolve Area Alias
	areaInput := strings.ToLower(optMap["area"])
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}
	// Display Names (Optional: Lookup ID back to Name for pretty printing)
	courseName := entity.CourseDisplayNameByCode[segaCourseID]
	if courseName == "" {
		// Fallback
		courseName = optMap["track"]
	}
	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	// Check Limit
	var limit = 1000

	// 3. Fetch Data
	listMostUsedCars, err := h.IDACCarService.GetListTopTACarsWithPercentage(context.Background(), segaCourseID, 4)
	if err != nil {
		sendDeferredError("⚠️ Failed to GetListTopTACarsWithPercentage: " + err.Error())
		return nil, err
	}
	if len(listMostUsedCars) == 0 {
		msg := fmt.Sprintf("# Not found Sega Data")
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return nil, err
	}
	if len(listMostUsedCars) < limit {
		limit = len(listMostUsedCars)
	}
	// 4. Calculate percentage
	return listMostUsedCars, nil
}

// GetPlayerTimeAttackRank assumes thresholds is sorted ascending by RequiredTime
func (h *Handler) GetPlayerTimeAttackRank(playerTime time.Time, courseID string, thresholds []*entity.TimeAttackRankingMetadata) (string, error) {
	if thresholds == nil {
		taTimeMetadata, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(context.Background(), courseID)
		if err != nil {
			return "", err
		}
		thresholds = taTimeMetadata
	}
	for _, t := range thresholds {
		if playerTime.Before(t.RequiredTime) || playerTime.Equal(t.RequiredTime) {
			return t.RankName, nil
		}
	}
	return "NO DATA", nil
}

func (h *Handler) HandleSetPlayerAlias(i *discordgo.InteractionCreate, key string, optMap map[string]string) {
	ign := optMap["ign"]
	area := optMap["area"]

	// Normalize Area
	if val, ok := entity.AreaAliases[strings.ToLower(area)]; ok {
		area = val
	}

	// Normalize Key (custom tags)
	isNumeric := regexp.MustCompile(`^\d+$`).MatchString(key)
	if !isNumeric {
		key = strings.ToLower(key)
	}

	// Save
	err := h.AliasRepo.SetPlayerAlias(key, ign, area)
	msg := ""
	if err != nil {
		msg = "❌ Failed to save alias: " + err.Error()
	} else {
		areaName := entity.AreaDisplayNameByCode[area]
		if areaName == "" {
			areaName = area
		}

		if isNumeric && len(key) > 15 {
			// --- NEW: Fetch Username to avoid tagging ---
			displayName := key // Fallback to ID if fetch fails

			// Try fetching user from Discord
			if u, err := h.Session.User(key); err == nil {
				displayName = u.Username
				// Optional: Use u.GlobalName if you prefer the display name
				if u.GlobalName != "" {
					displayName = u.GlobalName
				}
			}
			msg = fmt.Sprintf("✅ **User Linked:** %s → **%s** (%s)", displayName, ign, areaName)
		} else {
			msg = fmt.Sprintf("✅ **Tag Registered!**\nTag `%s` → **%s** (%s)\nUse it like: `/idac player-info player:%s`", key, ign, areaName, key)
		}
	}

	_ = h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content:         msg,
			AllowedMentions: &discordgo.MessageAllowedMentions{},
		},
	})
}

func (h *Handler) HandleTeamRanking(i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER INTERACTION
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Error Helper
	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}
	roundNum, err := h.SegaClient.GetCurrentRound()
	if err != nil {
		sendDeferredError("⚠️ Error fetching round info.")
		return
	}
	// 2. Resolve Country Filter
	filterCountryID := -1
	filterCountryName := "All"
	if countryInput, ok := optMap["country"]; ok && countryInput != "" {
		areaCode := countryInput
		normalizedInput := strings.ToLower(countryInput)
		if val, ok := entity.AreaAliases[normalizedInput]; ok {
			areaCode = val
		}
		if strings.HasPrefix(areaCode, "area-") {
			idStr := strings.TrimPrefix(areaCode, "area-")
			if id, err := strconv.Atoi(idStr); err == nil {
				filterCountryID = id
				if name, ok := entity.AreaDisplayNameByCode[areaCode]; ok {
					filterCountryName = name
				}
			}
		}
	}

	// 3. Parse Limit
	limit := 10
	if limitStr, ok := optMap["limit"]; ok {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 4. Determine Target Ranks
	rankCodeInput := optMap["rank"]
	targetRanks := []string{}
	rankDisplayName := "Unknown"

	if rankCodeInput == "all" {
		targetRanks = []string{"6", "5", "4", "3"}
		rankDisplayName = "All Classes"
	} else {
		targetRanks = []string{rankCodeInput}
		switch rankCodeInput {
		case "6":
			rankDisplayName = "Master"
		case "5":
			rankDisplayName = "Platinum"
		case "4":
			rankDisplayName = "Gold"
		case "3":
			rankDisplayName = "Silver"
		}
	}

	// 5. Fetch and Aggregate Records
	var allRecords []entity.TeamRecord
	for _, code := range targetRanks {
		records, err := h.SegaClient.GetTeamRanking(roundNum, code)
		if err != nil {
			if len(targetRanks) == 1 {
				sendDeferredError(fmt.Sprintf("⚠️ Error fetching data for rank %s", code))
				return
			}
			continue
		}
		allRecords = append(allRecords, records...)
	}

	if len(allRecords) == 0 {
		sendDeferredError("No records found.")
		return
	}

	// 6. Sort Combined Records by Points (Descending)
	sort.Slice(allRecords, func(i, j int) bool {
		p1, _ := strconv.Atoi(allRecords[i].Point)
		p2, _ := strconv.Atoi(allRecords[j].Point)
		return p1 > p2
	})

	// 7. Filter and Build Pages
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	foundCount := 0

	// Pre-calculate Header (will be added to every page)
	header := fmt.Sprintf("# Initial D Team Rankings (Round %d)\n", roundNum)
	header += fmt.Sprintf("**Class:** %s | **Region:** %s\n\n", rankDisplayName, filterCountryName)

	// Add header to the first page buffer
	currentMessage.WriteString(header)

	for _, r := range allRecords {
		// Apply Country Filter
		if filterCountryID != -1 && r.Country != filterCountryID {
			continue
		}
		// Stop if overall limit reached
		if foundCount >= limit {
			break
		}

		// Calculate display rank
		globalRank := foundCount + 1
		countryFlag := entity.GetCountryFlag(r.Country)

		// Format Entry
		entry := fmt.Sprintf("%d. %s **%s** \n", globalRank, r.LeagueEmoji, r.TeamName)
		entry += fmt.Sprintf("+ **Country:** %s\n", countryFlag)
		entry += fmt.Sprintf("+ **Points:** %s\n", r.Point)
		entry += fmt.Sprintf("+ **Ace:** %s | **Leader:** %s\n\n", r.AceUserName, r.LeaderUserName)

		// Check Splitting Condition (Every 5 teams OR 1900 chars)
		// Teams take up more vertical space, so 5 items per page is usually safer/cleaner than 10.
		// You can change `itemsInChunk >= 5` to 10 if you prefer.
		if itemsInChunk >= 5 || currentMessage.Len()+len(entry) > 1900 {
			pages = append(pages, currentMessage.String())

			currentMessage.Reset()
			currentMessage.WriteString(header) // Add header to next page
			itemsInChunk = 0
		}

		currentMessage.WriteString(entry)
		itemsInChunk++
		foundCount++
	}

	// Append whatever is left
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}

	if len(pages) == 0 {
		sendDeferredError("No teams found matching criteria.")
		return
	}
	// 8. Send via Pagination Helper
	h.SendPagination(i, pages)
}

func (h *Handler) HandlePlayerCompare(i *discordgo.InteractionCreate, optMap map[string]string) {
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	finalCourseID := optMap["track"]
	// 1. Input Validation
	//if optMap["variant"] == "" {
	//	h.SendDeferredError(i, "⚠️ **Missing Track**\nPlease select a track variant.")
	//	return
	//}
	// 2. Resolve Player 1
	p1Name, foundP1Area, isFoundP1, err := h.ResolvePlayerCredentialDB(optMap["player1"], optMap["area1"])
	if err != nil {
		fmt.Printf("⚠️ **Player 1 Error:** %s\n", err.Error())
	}
	if !isFoundP1 && optMap["area1"] == "" {
		h.SendDeferredError(i, "⚠️ **Searching Player1 by IGN error:** Please input area1.")
		return
	}
	// 3. Resolve Player 2
	p2Name, foundP2Area, isFoundP2, err := h.ResolvePlayerCredentialDB(optMap["player2"], optMap["area2"])
	if err != nil {
		fmt.Printf("⚠️ **Player 2 Error:** %s\n", err.Error())
	}
	if !isFoundP2 && optMap["area2"] == "" {
		h.SendDeferredError(i, "⚠️ **Searching Player2 by IGN error:** Please input area2.")
		return
	}
	// 4. Map sega codes (area, course, car) & make requests to get list TA
	var area1, area2 string
	if !isFoundP1 {
		area1 = entity.AreaAliases[optMap["area1"]] // optMap["area1"] = tokyo -> area1 = area-12
		p1Name = optMap["player1"]
	} else {
		area1 = foundP1Area // foundP1Area = area-12
	}
	if !isFoundP2 {
		area2 = entity.AreaAliases[optMap["area2"]]
		p2Name = optMap["player2"]
	} else {
		area2 = foundP2Area
	}
	taTimeMetadata, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(context.Background(), finalCourseID)
	if err != nil {
		h.SendDeferredError(i, "⚠️ error getting taMetadata:"+err.Error())
		return
	}
	var p1Result, p2Result *entity.TimeAttackRecord
	// Get Player 1 Sega Time Trail results
	listTAResult1, err := h.SegaClient.GetListTimeTrail(finalCourseID, area1, "car-all", "")
	if err != nil {
		h.SendDeferredError(i, "⚠️ error getting player1Info:"+err.Error())
		return
	}
	normalizedTarget := strings.ToLower(strings.TrimSpace(entity.NormalizeTextWidth(p1Name)))
	for _, result := range listTAResult1 {
		normalizedResultName := strings.ToLower(strings.TrimSpace(entity.NormalizeTextWidth(result.Name)))
		if normalizedResultName == normalizedTarget {
			p1Result = &result
			break
		}
	}
	// Get Player 2 Sega Time Trail results
	var listTAResult2 []entity.TimeAttackRecord
	// optimize: if same area, use same list
	if area2 == area1 {
		listTAResult2 = listTAResult1
	} else {
		listTAResult2, err = h.SegaClient.GetListTimeTrail(finalCourseID, area2, "car-all", "")
		if err != nil {
			h.SendDeferredError(i, "error getting player2Info:"+err.Error())
			return
		}
	}
	normalizedTarget = strings.ToLower(strings.TrimSpace(entity.NormalizeTextWidth(p2Name)))
	var duplicatedResult *entity.TimeAttackRecord
	for _, result := range listTAResult2 {
		normalizedResultName := strings.ToLower(strings.TrimSpace(entity.NormalizeTextWidth(result.Name)))
		if normalizedResultName == normalizedTarget {
			// if found result with same IGN & time -> record it
			if p1Result != nil && result.Record == p1Result.Record {
				duplicatedResult = &result
				continue
			}
			// if found result with same IGN & different time -> assign it to p2Result
			p2Result = &result
			break
		}
	}
	if p2Result == nil && duplicatedResult != nil {
		p2Result = duplicatedResult
	}
	// 6. Build Response
	var sb strings.Builder
	courseName := entity.CourseDisplayNameByCode[finalCourseID]
	if courseName == "" {
		courseName = optMap["track"]
	}
	areaName1 := entity.AreaDisplayNameByCode[area1]
	if areaName1 == "" {
		areaName1 = area1
	}
	areaName2 := entity.AreaDisplayNameByCode[area2]
	if areaName2 == "" {
		areaName2 = area2
	}

	sb.WriteString(fmt.Sprintf("# Player Comparison\n"))
	sb.WriteString(fmt.Sprintf("**Course:** %s\n", courseName))

	// Helper to print
	printPlayer := func(label, areaName, inputName, errStr string, p *entity.TimeAttackRecord) {
		sb.WriteString(fmt.Sprintf("### %s (%s): ", label, areaName))
		var taRank = "NO DATA"
		if p != nil {
			timeRecord, _ := entity.ParseRaceTime(p.Record)
			taRank, _ = h.GetPlayerTimeAttackRank(timeRecord, finalCourseID, taTimeMetadata)
			sb.WriteString(fmt.Sprintf("**%s**\n", p.Name))
			sb.WriteString(fmt.Sprintf("- **Local Rank:** #%s\n", p.Rank))
			sb.WriteString(fmt.Sprintf("- **Time:** `%s`\n", p.Record))
			sb.WriteString(fmt.Sprintf("- **Rank:** **%s**\n", taRank))
			sb.WriteString(fmt.Sprintf("- **Car:** %s\n", p.CarName))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n", inputName))
			if errStr != "" {
				sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", errStr))
			} else {
				sb.WriteString("- ❌ **Not Found**\n")
			}
		}
	}

	printPlayer("Player 1", areaName1, p1Name, "", p1Result)
	printPlayer("Player 2", areaName2, p2Name, "", p2Result)

	// 7. Calculate Delta
	if p1Result != nil && p2Result != nil {
		sb.WriteString("---\n")
		ms1, err1 := entity.ParseIdacTime(p1Result.Record)
		ms2, err2 := entity.ParseIdacTime(p2Result.Record)

		if err1 == nil && err2 == nil {
			diff := ms1 - ms2
			if diff < 0 {
				gap := entity.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p1Result.Name, strings.TrimPrefix(gap, "-")))
			} else if diff > 0 {
				gap := entity.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p2Result.Name, strings.TrimPrefix(gap, "+")))
			} else {
				sb.WriteString("🤝 **It's a tie!** Exact same time.")
			}
		}
	}

	finalContent := sb.String()
	h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &finalContent,
	})
}
