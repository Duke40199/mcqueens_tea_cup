package discord

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"McQueens_Tea_Cup/internal/domain"

	"github.com/bwmarrin/discordgo"
)

// Handler holds dependencies.
// It doesn't know about HTTP, it only knows "IdacRepository".
type Handler struct {
	Session   *discordgo.Session
	SegaRepo  domain.IdacRepository
	AliasRepo domain.AliasRepository
	OwnerID   string
}

// NewHandler creates our controller
func NewHandler(s *discordgo.Session, sega domain.IdacRepository, alias domain.AliasRepository) *Handler {
	return &Handler{
		Session:   s,
		SegaRepo:  sega,
		AliasRepo: alias,
		OwnerID:   "384015507302383616",
	}
}

func (h *Handler) HandleTimeAttack(i *discordgo.InteractionCreate, optMap map[string]string, specInput string) {
	var err error
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
	if val, ok := domain.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}

	// Resolve Car
	finalCarID := domain.ResolveCarID(optMap["car"], specInput)

	// Display Names (Optional: Lookup ID back to Name for pretty printing)
	courseName := domain.CourseDisplayNameByCode[finalCourseID]
	if courseName == "" {
		// Fallback
		courseName = optMap["track"]
	}

	areaName := optMap["area"]
	if val, ok := domain.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}
	carDisplayName := finalCarID
	baseCar := domain.ResolveCarID(optMap["car"], "")
	if optMap["car"] == "car-all" {
		carDisplayName = "All"
	} else if val, ok := domain.CarDisplayNameByCode[baseCar]; ok {
		if emoji, ok := domain.SpecEmojis[strings.ToLower(specInput)]; ok {
			carDisplayName = fmt.Sprintf("%s %s", emoji, val)
		} else {
			carDisplayName = val
		}
	}

	// Check Limit
	var limit int
	if _, ok := optMap["limit"]; ok {
		limit, err = strconv.Atoi(optMap["limit"])
		if err != nil {
			sendDeferredError("⚠️ Limit must be a number")
			return
		}
	} else {
		limit = 10
	}

	// 3. Fetch Data
	finalCar := optMap["car"]
	finalArea := optMap["area"]
	records, err := h.SegaRepo.GetTimeAttack(finalCourseID, finalArea, finalCar, "")

	if len(records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\nNo records found.", courseName, areaName, carDisplayName)
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(records) < limit {
		limit = len(records)
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
		entry := fmt.Sprintf("%s. **%s** — %s — `%s`\n", r.Rank, r.Name, r.CarName, r.Record)

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

func (h *Handler) HandlePlayerAliasSet(i *discordgo.InteractionCreate, key string, optMap map[string]string) {
	ign := optMap["ign"]
	area := optMap["area"]

	// Normalize Area
	if val, ok := domain.AreaAliases[strings.ToLower(area)]; ok {
		area = val
	}

	// Normalize Key (custom tags)
	isNumeric := regexp.MustCompile(`^\d+$`).MatchString(key)
	if !isNumeric {
		key = strings.ToLower(key)
	}

	// Save
	err := domain.Aliases.Set(key, ign, area)

	msg := ""
	if err != nil {
		msg = "❌ Failed to save alias: " + err.Error()
	} else {
		areaName := domain.AreaDisplayNameByCode[area]
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

	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			// Extra safety: explicit empty AllowedMentions ensures no one gets pinged even if you use <@ID>
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

	// 1. Fetch Current Round Info
	roundURL := "https://initiald.sega.jp/inidac/json/ranking/v1/currentRoundInfo.json"
	respRound, err := http.Get(roundURL)
	if err != nil {
		sendDeferredError("⚠️ Error fetching round info.")
		return
	}
	defer respRound.Body.Close()

	bodyBytes, err := io.ReadAll(respRound.Body)
	if err != nil {
		sendDeferredError("⚠️ Error reading round info.")
		return
	}

	roundStr := strings.TrimSpace(string(bodyBytes))
	roundNum, err := strconv.Atoi(roundStr)
	if err != nil {
		sendDeferredError(fmt.Sprintf("⚠️ Error parsing round number: %s", roundStr))
		return
	}

	// 2. Resolve Country Filter
	filterCountryID := -1
	filterCountryName := "All"
	if countryInput, ok := optMap["country"]; ok && countryInput != "" {
		areaCode := countryInput
		normalizedInput := strings.ToLower(countryInput)
		if val, ok := domain.AreaAliases[normalizedInput]; ok {
			areaCode = val
		}
		if strings.HasPrefix(areaCode, "area-") {
			idStr := strings.TrimPrefix(areaCode, "area-")
			if id, err := strconv.Atoi(idStr); err == nil {
				filterCountryID = id
				if name, ok := domain.AreaDisplayNameByCode[areaCode]; ok {
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
	var allRecords []domain.TeamRecord
	for _, code := range targetRanks {
		records, err := h.SegaRepo.GetTeamRanking(roundNum, code)
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
		countryFlag := domain.GetCountryFlag(r.Country)

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
