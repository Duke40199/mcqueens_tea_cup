package discord

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"McQueens_Tea_Cup/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

func (h *Handler) HandleTimeAttack(cc *CommandContext) error {
	optMap := cc.OptMap
	specInput := cc.SpecInput

	// 1. Validate input
	inputTrack := optMap["track"]
	//inputArea := strings.ToLower(optMap["country-select"])
	inputArea := strings.ToLower(optMap["area"])

	finalCourseID := inputTrack
	courseName := entity.CourseDisplayNameByCode[finalCourseID]
	if courseName == "" {
		return cc.Edit("⚠️ **Track not found based on input.**")
	}
	// 1.b. Car Input
	var finalCarID string
	var foundCar *entity.CarMetadata
	isInputCar := optMap["car"] != "" && optMap["car"] != "all"
	if !isInputCar {
		finalCarID = "car-all"
	}
	if isInputCar {
		if specInput == "" {
			return cc.Edit("⚠️ **Please select a car spec.**")
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
			return cc.Edit("⚠️ Car not found with input.")
		}
		if inputArea != "" && inputArea != "all" {
			return cc.Edit("ℹ️ Currently search with car only supported with __**all**__ area input.")
		}
	}
	// Get Car Display Name from Input
	var carDisplayName string
	if !isInputCar {
		carDisplayName = "All"
	} else {
		carDisplayName = foundCar.Name
	}
	// 1.d Resolve Player Area
	if val, ok := entity.AreaAliases[inputArea]; ok {
		optMap["area"] = val
	}
	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}

	resultLimit := 1000
	// 2. Fetch Data via the service
	finalArea := optMap["area"]
	records, err := h.IDACTimeAttackMetadataService.GetTimeTrail(cc.Ctx, finalCourseID, finalArea, finalCarID, specInput)
	if err != nil {
		return fmt.Errorf("fetching time trial records: %w", err)
	}
	if len(records) == 0 {
		return cc.Edit(fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\nNo records found.", courseName, areaName, carDisplayName))
	}
	if len(records) < resultLimit {
		resultLimit = len(records)
	}
	taTimeMetadata, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(cc.Ctx, finalCourseID)
	if err != nil {
		return fmt.Errorf("fetching time attack metadata: %w", err)
	}

	// 3. Build Pages (Slice of Strings)
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0

	// Pre-calculate Header
	header := fmt.Sprintf("# Initial D Rankings (Time Trial)\n 🗾 : %s | 🌎 : %s | 🚗 : %s\n", courseName, areaName, carDisplayName)
	currentMessage.WriteString(header)

	// Top-3 most used cars section
	listCarPercentages, err := h.IDACCarService.GetListTopTACarsWithPercentage(cc.Ctx, finalCourseID, 4)
	if err != nil {
		return fmt.Errorf("fetching top TA cars: %w", err)
	}
	if len(listCarPercentages) < 3 {
		return cc.Edit("⚠️ Not enough car data available for this track yet.")
	}
	headerCarPercentage := "## Top 3 Most Used Cars (based on top 1000 **Global** results)\n"
	// car name sega format: FD3S[DH]
	listCarNameSegaFormat := make([]string, 0)
	for i := 0; i < 3; i++ {
		listCarNameSegaFormat = append(listCarNameSegaFormat, listCarPercentages[i].SegaCarName)
	}
	carListFullInfo, err := h.IDACCarService.GetListCarDetailByTAFormat(cc.Ctx, listCarNameSegaFormat)
	if err != nil {
		return fmt.Errorf("fetching car detail by TA format: %w", err)
	}
	carCount := 0
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

		// Split if 10 items OR length > 2000
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
	// 4. Hand over to Pagination Helper
	cc.SendPages(pages)
	return nil
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

func (h *Handler) HandleTeamRanking(cc *CommandContext) error {
	optMap := cc.OptMap

	// 1. Resolve Country Filter
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

	// 2. Parse Limit
	limit := 1000
	if limitStr, ok := optMap["limit"]; ok {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// 3. Determine Target Ranks
	rankCodeInput := optMap["rank"]
	var targetRanks []string
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

	// 4. Fetch aggregated + sorted records via the service
	roundNum, allRecords, err := h.TeamService.GetSortedTeamRankings(cc.Ctx, targetRanks)
	if err != nil {
		return fmt.Errorf("fetching team rankings: %w", err)
	}
	if len(allRecords) == 0 {
		return cc.Edit("No records found.")
	}

	// 5. Filter and Build Pages
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0
	foundCount := 0

	header := fmt.Sprintf("# Initial D Team Rankings (Round %d)\n", roundNum)
	header += fmt.Sprintf("**Class:** %s | **Region:** %s\n\n", rankDisplayName, filterCountryName)
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

		globalRank := foundCount + 1
		countryFlag := entity.GetCountryFlag(r.Country)

		entry := fmt.Sprintf("%d. %s **%s** \n", globalRank, r.LeagueEmoji, r.TeamName)
		entry += fmt.Sprintf("+ **Country:** %s\n", countryFlag)
		entry += fmt.Sprintf("+ **Points:** %s\n", r.Point)
		entry += fmt.Sprintf("+ **Ace:** %s | **Leader:** %s\n\n", r.AceUserName, r.LeaderUserName)

		// Split every 5 teams OR 1900 chars
		if itemsInChunk >= 5 || currentMessage.Len()+len(entry) > 1900 {
			pages = append(pages, currentMessage.String())
			currentMessage.Reset()
			currentMessage.WriteString(header)
			itemsInChunk = 0
		}

		currentMessage.WriteString(entry)
		itemsInChunk++
		foundCount++
	}

	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	if len(pages) == 0 {
		return cc.Edit("No teams found matching criteria.")
	}
	// 6. Send via Pagination Helper
	cc.SendPages(pages)
	return nil
}

func (h *Handler) HandlePlayerCompare(cc *CommandContext) error {
	optMap := cc.OptMap
	finalCourseID := optMap["track"]

	// 1. Resolve Player 1
	p1Name, foundP1Area, isFoundP1, err := h.PlayerService.ResolvePlayer(cc.Ctx, optMap["player1"], optMap["area1"])
	if err != nil {
		fmt.Printf("⚠️ **Player 1 Error:** %s\n", err.Error())
	}
	if !isFoundP1 && optMap["area1"] == "" {
		return cc.Edit("⚠️ **Searching Player1 by IGN error:** Please input area1.")
	}
	// 2. Resolve Player 2
	p2Name, foundP2Area, isFoundP2, err := h.PlayerService.ResolvePlayer(cc.Ctx, optMap["player2"], optMap["area2"])
	if err != nil {
		fmt.Printf("⚠️ **Player 2 Error:** %s\n", err.Error())
	}
	if !isFoundP2 && optMap["area2"] == "" {
		return cc.Edit("⚠️ **Searching Player2 by IGN error:** Please input area2.")
	}
	// 3. Map sega codes (area, course, car) & make requests to get list TA
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
	taTimeMetadata, err := h.IDACTimeAttackMetadataService.GetMetadataBySegaCourseID(cc.Ctx, finalCourseID)
	if err != nil {
		return fmt.Errorf("fetching time attack metadata: %w", err)
	}
	var p1Result, p2Result *entity.TimeAttackRecord
	// Get Player 1 Sega Time Trail results
	listTAResult1, err := h.IDACTimeAttackMetadataService.GetTimeTrail(cc.Ctx, finalCourseID, area1, "car-all", "")
	if err != nil {
		return fmt.Errorf("fetching player 1 time trial: %w", err)
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
		listTAResult2, err = h.IDACTimeAttackMetadataService.GetTimeTrail(cc.Ctx, finalCourseID, area2, "car-all", "")
		if err != nil {
			return fmt.Errorf("fetching player 2 time trial: %w", err)
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
	// 4. Build Response
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

	sb.WriteString("# Player Comparison\n")
	sb.WriteString(fmt.Sprintf("**Course:** %s\n", courseName))

	// Helper to print
	printPlayer := func(label, areaName, inputName, errStr string, p *entity.TimeAttackRecord) {
		sb.WriteString(fmt.Sprintf("### %s (%s): ", label, areaName))
		taRank := "NO DATA"
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

	// 5. Calculate Delta
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

	return cc.Edit(sb.String())
}
