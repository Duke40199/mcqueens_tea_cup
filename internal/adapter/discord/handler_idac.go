package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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

	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}
	// 2. Validate Inputs
	// We look for 'variant' because that holds the actual Course ID now
	if optMap["variant"] == "" || optMap["variant"] == "none" || optMap["area"] == "" {
		sendDeferredError("⚠️ **Missing Arguments**\nPlease select both a **Track** and a **Variant**, plus an Area.")
		return
	}
	// Defaults
	if _, ok := optMap["car"]; !ok {
		optMap["car"] = "car-all"
	}
	// The User selected: Track="Akina", Variant="course-12" (Downhill)
	// We only care about the Variant ID now.
	finalCourseID := optMap["variant"]

	// Resolve Area Alias
	areaInput := strings.ToLower(optMap["area"])
	if val, ok := entity.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}

	// Resolve Car
	finalCarID := entity.ResolveCarID(optMap["car"], specInput)

	// Display Names (Optional: Lookup ID back to Name for pretty printing)
	courseName := entity.CourseDisplayNameByCode[finalCourseID]
	if courseName == "" {
		// Fallback
		courseName = optMap["track"]
	}

	areaName := optMap["area"]
	if val, ok := entity.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	carDisplayName := finalCarID
	baseCar := entity.ResolveCarID(optMap["car"], "")
	if optMap["car"] == "car-all" {
		carDisplayName = "All"
	} else if val, ok := entity.CarDisplayNameByCode[baseCar]; ok {
		if emoji, ok := entity.SpecEmojis[strings.ToLower(specInput)]; ok {
			carDisplayName = fmt.Sprintf("%s %s", emoji, val)
		} else {
			carDisplayName = val
		}
	}

	// Check Limit
	var limit = 1000

	// 3. Fetch Data
	finalArea := optMap["area"]
	records, err := h.SegaClient.GetTimeAttack(finalCourseID, finalArea, finalCarID, specInput)
	if err != nil {
		sendDeferredError("⚠️ Failed to fetch data from Sega API: " + err.Error())
		return
	}
	if len(records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\nNo records found.", courseName, areaName, carDisplayName)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(records) < limit {
		limit = len(records)
	}
	taTimeMetadata, err := h.TATimeMetadataRepo.GetByCourseID(context.TODO(), finalCourseID)
	if taTimeMetadata == nil {
		sendDeferredError("⚠️ Failed to metadata TA time.")
		return
	}
	// 4. Build Pages (Slice of Strings)
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0

	// Pre-calculate Header
	header := fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\n", courseName, areaName, carDisplayName)

	// Initialize first page with header
	currentMessage.WriteString(header)

	for j := 0; j < limit; j++ {
		r := records[j]
		timeRecord, _ := entity.ParseRaceTime(r.Record)
		recordRank, _ := h.GetPlayerTimeAttackRank(timeRecord, finalCourseID, taTimeMetadata)
		entry := fmt.Sprintf("%s. **%s**  — `%s` — **%s** — `%s`\n", r.Rank, recordRank, r.Record, r.Name, r.CarName)

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

// GetPlayerTimeAttackRank assumes thresholds is sorted ascending by RequiredTime
func (h *Handler) GetPlayerTimeAttackRank(playerTime time.Time, courseID string, thresholds []*entity.TimeAttackRankingMetadata) (string, error) {
	if thresholds == nil {
		taTimeMetadata, err := h.TATimeMetadataRepo.GetByCourseID(context.TODO(), courseID)
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
	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}
	// 2. Resolve Player 1
	p1Name, p1Area, _, err := h.ResolvePlayerCredentialDB(optMap["player1"], optMap["area1"])
	if err != nil {
		sendDeferredError("⚠️ **Player 1 Error:** " + err.Error())
		return
	}
	if p1Area == "" {
		// This happens if "player1" wasn't a known alias AND "area1" was empty.
		sendDeferredError(fmt.Sprintf("⚠️ **Unknown Player 1:** `%s`\nTag not found and no Area provided.", optMap["player1"]))
		return
	}

	// 3. Resolve Player 2
	p2Name, p2Area, _, err := h.ResolvePlayerCredentialDB(optMap["player2"], optMap["area2"])
	if err != nil {
		sendDeferredError("⚠️ **Player 2 Error:** " + err.Error())
		return
	}

	// 4. Smart Inheritance for Player 2
	// If P2 has no area (meaning it wasn't a complete alias, and user didn't type area2),
	// we default to Player 1's area.
	if p2Area == "" {
		p2Area = p1Area
	}

	// 5. Update Map & Validate Track
	optMap["ign1"] = p1Name
	optMap["area1"] = p1Area
	optMap["ign2"] = p2Name
	optMap["area2"] = p2Area

	if optMap["variant"] == "" {
		sendDeferredError("⚠️ **Missing Track**\nPlease select a track variant.")
		return
	}
	// 3. Resolve Areas
	resolveArea := func(input string) string {
		normalized := strings.ToLower(input)
		// Explicitly check for 'all' keyword if missing from map, just in case
		if normalized == "world" || normalized == "global" {
			return "all"
		}

		if val, ok := entity.AreaAliases[normalized]; ok {
			return val
		}
		// Fallback: Return input as-is (allows 'all' to pass through if input is 'all')
		return normalized
	}

	area1 := resolveArea(optMap["area1"])
	area2 := ""

	// Handle Area 2
	if val, ok := optMap["area2"]; ok && val != "" {
		area2 = resolveArea(val)
	} else {
		area2 = area1 // Default to Area 1
	}

	courseID := optMap["variant"]
	taTimeMetadata, err := h.TATimeMetadataRepo.GetByCourseID(context.Background(), courseID)
	if err != nil {
		sendDeferredError("error getting taMetadata:" + err.Error())
		return
	}
	// 4. Helper: Fetch Data & Find Player
	fetchAndFind := func(areaCode, targetIgn string) (*entity.TimeAttackRecord, string, error) {
		baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
		// This URL structure supports "ta_course-12_all_car-all.json" perfectly
		filename := fmt.Sprintf("ta_%s_%s_%s.json", courseID, areaCode, "car-all")
		fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)
		//fmt.Println("fullURL", fullURL)
		resp, err := http.Get(fullURL)
		if err != nil {
			return nil, "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Sprintf("Status %d", resp.StatusCode), nil
		}

		var data entity.IdacTimeAttackRecordResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, "Parse Error", nil
		}

		// Find Player
		normalizedTarget := strings.ToLower(entity.NormalizeTextWidth(targetIgn))
		for _, r := range data.Records {
			if strings.ToLower(entity.NormalizeTextWidth(r.Name)) == normalizedTarget {
				val := r
				return &val, "", nil
			}
		}
		return nil, "", nil
	}

	// 5. Execution
	var p1, p2 *entity.TimeAttackRecord
	var errStr1, errStr2 string

	// OPTIMIZATION: If areas are identical (e.g. both 'all'), fetch once
	if area1 == area2 {
		baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
		filename := fmt.Sprintf("ta_%s_%s_%s.json", courseID, area1, "car-all")
		fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)
		resp, err := http.Get(fullURL)
		if err == nil && resp.StatusCode == 200 {
			var data entity.IdacTimeAttackRecordResponse
			_ = json.NewDecoder(resp.Body).Decode(&data)
			resp.Body.Close()

			target1 := strings.ToLower(entity.NormalizeTextWidth(optMap["ign1"]))
			target2 := strings.ToLower(entity.NormalizeTextWidth(optMap["ign2"]))

			for _, r := range data.Records {
				if p1 != nil && p2 != nil {
					break
				}
				normName := strings.ToLower(entity.NormalizeTextWidth(r.Name))
				if p1 == nil && normName == target1 {
					val := r
					p1 = &val
				}
				if p2 == nil && normName == target2 {
					val := r
					p2 = &val
				}
			}
		} else {
			errStr1 = "Failed to fetch data"
			errStr2 = "Failed to fetch data"
		}
	} else {
		// Different Areas: Fetch Separate
		var err error
		p1, errStr1, err = fetchAndFind(area1, optMap["ign1"])
		if err != nil {
			errStr1 = "Network Error"
		}

		p2, errStr2, err = fetchAndFind(area2, optMap["ign2"])
		if err != nil {
			errStr2 = "Network Error"
		}
	}

	// 6. Build Response
	var sb strings.Builder

	courseName := entity.CourseDisplayNameByCode[courseID]
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
			taRank, _ = h.GetPlayerTimeAttackRank(timeRecord, courseID, taTimeMetadata)
			sb.WriteString(fmt.Sprintf("**%s**\n", p.Name))
			sb.WriteString(fmt.Sprintf("- **Local Rank:** #%s\n", p.Rank))
			sb.WriteString(fmt.Sprintf("- **Time:** `%s`\n", p.Record))
			sb.WriteString(fmt.Sprintf("- **Rank:** `%s`\n", taRank))
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

	printPlayer("Player 1", areaName1, optMap["ign1"], errStr1, p1)
	printPlayer("Player 2", areaName2, optMap["ign2"], errStr2, p2)

	// 7. Calculate Delta
	if p1 != nil && p2 != nil {
		sb.WriteString("---\n")
		ms1, err1 := entity.ParseIdacTime(p1.Record)
		ms2, err2 := entity.ParseIdacTime(p2.Record)

		if err1 == nil && err2 == nil {
			diff := ms1 - ms2
			if diff < 0 {
				gap := entity.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p1.Name, strings.TrimPrefix(gap, "-")))
			} else if diff > 0 {
				gap := entity.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p2.Name, strings.TrimPrefix(gap, "+")))
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

func (h *Handler) HandlePlayerInfo(i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// 1. Validate & Parse Inputs
	_, hasIgn := optMap["user"]
	if !hasIgn || optMap["course"] == "" {
		errStr := "⚠️ Missing required arguments (ign, course, or area)."
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	ign, areaInput, _, err := h.ResolvePlayerCredentialDB(optMap["user"], "")
	// --- NORMALIZE INPUT ---
	// 1. Normalize width (ＳＨＩＲＯ -> SHIRO)
	// 2. Lowercase (SHIRO -> shiro) for case-insensitive comparison
	targetName := strings.ToLower(entity.NormalizeTextWidth(ign))
	// Resolve Aliases
	courseInput := strings.ToLower(optMap["course"])
	if val, ok := entity.CourseAliases[courseInput]; ok {
		courseInput = val
	}
	//areaInput := strings.ToLower(optMap["area"])
	//if val, ok := domain.AreaAliases[areaInput]; ok {
	//	areaInput = val
	//}

	// 2. Construct URL
	// We strictly use "car-all" to find the player regardless of what car they drove
	baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
	filename := fmt.Sprintf("ta_%s_%s_%s.json", courseInput, areaInput, "car-all")
	fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)

	// 3. Fetch Data
	resp, err := http.Get(fullURL)
	if err != nil {
		errStr := "❌ Network error contacting SEGA."
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("❌ Failed to fetch ranking data (Status: %d). Check your course/area codes.", resp.StatusCode)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}

	// 4. Parse JSON
	var data entity.IdacTimeAttackRecordResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		errStr := "❌ Error parsing SEGA data."
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}

	// 5. Search for Player
	var foundRecord *entity.TimeAttackRecord
	var leaderRecord *entity.TimeAttackRecord

	for idx, r := range data.Records {
		// Capture Leader
		if idx == 0 {
			val := r
			leaderRecord = &val
		}

		// --- NORMALIZE RECORD NAME ---
		// We normalize the record name from SEGA too, just in case
		recordNameNormalized := strings.ToLower(entity.NormalizeTextWidth(r.Name))

		if recordNameNormalized == targetName {
			val := r
			foundRecord = &val
			break
		}
	}

	// 6. Build Response
	var sb strings.Builder

	// Display Names
	courseName := entity.CourseDisplayNameByCode[courseInput]
	if courseName == "" {
		courseName = courseInput
	}

	areaName := entity.AreaDisplayNameByCode[areaInput]
	if areaName == "" {
		areaName = areaInput
	}

	sb.WriteString(fmt.Sprintf("**Player Info**\n"))
	sb.WriteString(fmt.Sprintf("**In-game Name:** `%s` | **Course:** %s | **Area:** %s\n\n", ign, courseName, areaName))

	if foundRecord != nil {
		// --- Calculate Delta ---
		deltaStr := ""
		if leaderRecord != nil && foundRecord.Name != leaderRecord.Name {
			playerMs, err1 := entity.ParseIdacTime(foundRecord.Record)
			leaderMs, err2 := entity.ParseIdacTime(leaderRecord.Record)

			// Only calculate if parsing succeeded
			if err1 == nil && err2 == nil {
				diff := playerMs - leaderMs
				deltaStr = fmt.Sprintf(" (%s)", entity.FormatIdacTimeDelta(diff))
			}
		}
		// Player Found
		sb.WriteString(fmt.Sprintf("## 📊 Rank #%s\n", foundRecord.Rank)) // Using string rank from JSON usually safest
		sb.WriteString(fmt.Sprintf("** ⏱️Time:** `%s`\n", foundRecord.Record))
		sb.WriteString(fmt.Sprintf("** 🚗Car:** %s\n", foundRecord.CarName))

		// Comparison Logic
		if leaderRecord != nil {
			sb.WriteString("\n__**Leader Comparison:**__\n")
			if foundRecord.Name == leaderRecord.Name {
				sb.WriteString(fmt.Sprintf("🏆 **#1 %s**: `%s` (You are the leader! 👑)\n", leaderRecord.Name, leaderRecord.Record))
			} else {
				// Format: #1 LEADERNAME: 3'14"727 (+0'03"123)
				sb.WriteString(fmt.Sprintf("🏆 **#1 %s**: `%s`%s\n", leaderRecord.Name, leaderRecord.Record, deltaStr))
			}
		}
	} else {
		// Player Not Found
		sb.WriteString(fmt.Sprintf("❌ Player **%s** not found in the top %d records for this course/area.\n", ign, len(data.Records)))
		sb.WriteString("> *Note: Ensure the IGN is exact (though case is ignored) and the Area is correct.*")
	}

	// Send Response
	finalContent := sb.String()
	h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &finalContent,
	})
}

func (h *Handler) ResolvePlayerCredentialDB(input, manualArea string) (string, string, bool, error) {
	// 1. Sanitize Inputs
	cleanInput := strings.TrimSpace(input)

	// 2. Regex for Discord Mention/ID
	reMention := regexp.MustCompile(`^<@!?(\d+)>$`)
	// reID := regexp.MustCompile(`^\d{17,20}$`)

	var lookupKey string
	if match := reMention.FindStringSubmatch(cleanInput); len(match) > 1 {
		lookupKey = match[1] // Extract ID from <@123>
	} else {
		lookupKey = cleanInput // Raw ID
	}

	// 3. Check Alias Store (The "Check First" Logic)
	var aliasIgn, aliasArea string
	var foundAlias bool

	if lookupKey != "" {
		// Case A: Discord ID -> Strict Lookup
		playerAlias, isFound, err := h.AliasRepo.GetByAliasKey(lookupKey)
		if err != nil {
			fmt.Println(err)
			return "", "", false, err
		}
		if !isFound {
			return "", "", false, fmt.Errorf("couldn't find a matching alias")
		}
		foundAlias = true
		aliasIgn = playerAlias.Ign
		aliasArea = playerAlias.Area
	} else {
		areaCode := entity.AreaAliases[manualArea]
		// Case B: Text Input -> Try Custom Tag Lookup
		// Use lowercase key for case-insensitive matching
		val, ok, err := h.AliasRepo.GetByIgnAndAreaCode(strings.ToLower(cleanInput), areaCode)
		if ok && err == nil {
			aliasIgn = val.Ign
			aliasArea = val.Area
			foundAlias = true
		} else {
			return "", "", false, fmt.Errorf("couldn't find a matching alias")
		}
	}

	// 4. Merge Logic (Alias vs Manual)
	finalIgn := cleanInput
	finalArea := aliasArea

	if foundAlias {
		finalIgn = aliasIgn
		// If Manual Area is provided, it OVERRIDES the alias area.
		// If not, we fall back to the alias area.
		if finalArea == "" {
			finalArea = aliasArea
		}
	}

	//// 5. Normalize Final Area (Handle "world", "vn", etc.)
	//if finalArea != "" {
	//	norm := strings.ToLower(finalArea)
	//	if norm == "world" || norm == "global" {
	//		norm = "all"
	//	}
	//	//if val, ok := AreaAliases[norm]; ok {
	//	//	finalArea = val
	//	//}
	//}

	// 6. Final check: Do we have an area?
	// If no alias was found AND no manual area provided, we can't search.
	if finalArea == "" {
		return "", "", false, nil // Return empty to signal "Missing Area"
	}

	return finalIgn, finalArea, foundAlias, nil
}
