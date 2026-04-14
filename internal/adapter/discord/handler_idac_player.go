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

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain/entity"
)

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

func (h *Handler) HandlePlayerTournamentInfo(i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	sendDeferredError := func(msg string) {
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}

	// 1. Validate & Parse Inputs
	areaInput, isInputArea := optMap["area"]
	if !isInputArea {
		sendDeferredError("❌ Area input is required.")
		return
	}
	areaCode := entity.AreaAliases[areaInput]
	// 2. Get ranking / grade data from DB
	listGradeCfg, err := h.RankingCfgRepo.GetListPlayerGradeCfg(context.TODO())
	if err != nil {
		sendDeferredError("❌ Error getting ranking configs from DB.")
		return
	}
	// key: sega_id
	gradeCfgMap := make(map[string]*entity.PlayerGradeCfg)
	for _, cfg := range listGradeCfg {
		gradeCfgMap[cfg.SegaID] = cfg
	}
	// 3. Get Sega data: OB Ranking & Account Grade
	obRankingRes, err := h.SegaClient.GetListOBRanking("all", areaCode)
	if err != nil {
		errStr := "❌ Error getting list OB ranking from Sega"
		_, _ = h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	listSegaPlayerGrade, err := h.SegaClient.GetListPlayerGrade(areaCode)
	if err != nil {
		errStr := "❌ Error getting list player grade from Sega"
		_, _ = h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	// key: segaID
	obRankingCfgMap, err := h.OBRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		sendDeferredError("⚠️ Error fetching ranking configuration.")
		return
	}
	// 4.a. Map player data - By OB Ranking
	// key: player name
	mapFoundPlayers := make(map[string]entity.PlayerTournamentInfo)
	for _, obRanking := range obRankingRes.Records {
		var foundPlayer entity.PlayerTournamentInfo
		var ok bool
		if foundPlayer, ok = mapFoundPlayers[obRanking.Name]; !ok {
			foundPlayer = entity.PlayerTournamentInfo{
				Name:     obRanking.Name,
				GradeExp: 0,
			}
			mapFoundPlayers[obRanking.Name] = foundPlayer
		}
		if foundOBRankCfg, ok := obRankingCfgMap[obRanking.OnlineBattleRankId]; ok {
			foundPlayer.OBRank = foundOBRankCfg.Name
			foundPlayer.OBRankNum = obRanking.GetDisplayStarCount()
			exp, _ := strconv.Atoi(obRanking.Point)
			foundPlayer.OBRankExp = exp
		}
		mapFoundPlayers[obRanking.Name] = foundPlayer
	}
	// 4.b. Map player data - By Player Grade
	for _, segaPlayerGrade := range listSegaPlayerGrade.Records {
		var foundPlayer entity.PlayerTournamentInfo
		var ok bool
		if foundPlayer, ok = mapFoundPlayers[segaPlayerGrade.Name]; !ok {
			foundPlayer = entity.PlayerTournamentInfo{
				Name: segaPlayerGrade.Name,
			}
		}
		foundPlayer.GradeExp = segaPlayerGrade.GradeExp
		// get grade text
		if foundGrade, ok := gradeCfgMap[segaPlayerGrade.GradeID]; ok {
			foundPlayer.Grade = foundGrade.Name
		}
		// get grade number
		if foundGradeNum, ok := gradeCfgMap[segaPlayerGrade.NumberIcon]; ok {
			foundPlayer.GradeNum = foundGradeNum.Name
		}
		mapFoundPlayers[segaPlayerGrade.Name] = foundPlayer
	}
	// 6. Build Response
	var playerCount int
	var pages []string
	var currentMessage strings.Builder
	itemsInChunk := 0

	areaName := entity.AreaDisplayNameByCode[areaInput]
	if areaName == "" {
		areaName = areaInput
	}
	// header
	header := fmt.Sprintf("## Player Info (Tournament Mode)\n"+
		"### 🌎 : %s | Round: %s\n"+"### Calculated at: %s.\n\n", areaName, "all", obRankingRes.CalcDate)
	currentMessage.WriteString(header)
	// body
	if len(mapFoundPlayers) == 0 {
		errStr := "❌ Player map not found"
		_, _ = h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	listFoundPlayers := make([]*entity.PlayerTournamentInfo, 0)
	for _, player := range mapFoundPlayers {
		listFoundPlayers = append(listFoundPlayers, &player)
	}
	sort.Slice(listFoundPlayers, func(i, j int) bool {
		return listFoundPlayers[i].OBRankExp > listFoundPlayers[j].OBRankExp
	})

	for _, player := range listFoundPlayers {
		playerCount++
		if player.Grade == "" {
			player.Grade = "N/A"
		}
		if player.OBRank == "" {
			player.OBRank = "N/A"
		}
		var playerEntry = fmt.Sprintf("%d. **%s** — **%s%s** `(%d pts)` — **%s %s** `(%d pts)`\n",
			playerCount, player.Name, player.Grade, player.GradeNum, player.GradeExp, player.OBRank, player.OBRankNum, player.OBRankExp)
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
	// Append final page
	if currentMessage.Len() > 0 {
		pages = append(pages, currentMessage.String())
	}
	h.SendPagination(i, pages)
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
