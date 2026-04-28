package discord

import (
	"context"
	"fmt"
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
	playerInfo, hasPlayerInfo := optMap["user"]
	if !hasPlayerInfo {
		errStr := "⚠️ Missing player info"
		h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	// 2. Get Player Grade Info
	playerName, playerArea, _, err := h.ResolvePlayerCredentialDB(playerInfo, "")
	if err != nil {
		h.SendDeferredError(i, "Cannot find player info.")
		return
	}
	playerGrade, err := h.SegaClient.GetPlayerGradeByIGN(playerName, playerArea)
	if err != nil {
		h.SendDeferredError(i, "Error finding player grade.")
		return
	}
	if playerGrade == nil {
		h.SendDeferredError(i, "Cannot find player grade.")
		return
	}
	playerGradeCfgs, err := h.RankingCfgRepo.GetPlayerGradeBySegaIDs(context.TODO(), playerGrade.GradeID, playerGrade.NumberIcon)
	if err != nil {
		h.SendDeferredError(i, "Error getting player grade in DB."+err.Error())
		return
	}
	var gradeName, gradeNum string
	for _, playerGradeCfg := range playerGradeCfgs {
		switch playerGradeCfg.Type {
		case "RANK_NUMBER":
			if playerGradeCfg.Emoji != nil {
				gradeNum = *playerGradeCfg.Emoji
			} else {
				gradeNum = playerGradeCfg.Name
			}
			continue
		case "GRADE":
			if playerGradeCfg.Emoji != nil {
				gradeName = *playerGradeCfg.Emoji
			} else {
				gradeName = playerGradeCfg.Name
			}
		}
	}
	// 3. Get OB Info
	obRankingRes, err := h.SegaClient.GetOBRankingByIGN(playerName, "all", playerArea)
	if err != nil {
		h.SendDeferredError(i, "❌ Error getting list OB ranking from Sega")
		return
	}
	// key: segaID
	obRankingCfgMap, err := h.OBRankingCfgRepo.GetRankingCfgMap()
	if err != nil {
		h.SendDeferredError(i, "⚠️ Error fetching OB ranking configuration from DB.")
		return
	}

	// 4. Build Response
	var sb strings.Builder
	areaName := entity.AreaDisplayNameByCode[playerArea]
	if areaName == "" {
		areaName = playerArea
	}
	sb.WriteString(fmt.Sprintf(
		"### **IGN:** %s | Area: %s\n"+
			"### **Account Grade**:\n"+
			"# %s***%s***\n"+
			"### **Online Battle Rank:**\n"+
			"# %s\n", playerName, "VN", gradeName, gradeNum, obRankingRes.OnlineBattleRankId))
	// Send Response
	finalContent := sb.String()
	embed := &discordgo.MessageEmbed{
		Title:       "__**Player Profile**__",
		Description: finalContent,
	}
	h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
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
