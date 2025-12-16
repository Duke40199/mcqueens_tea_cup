package discord

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain"
	idac_domain "McQueens_Tea_Cup/internal/domain/idac"
)

//go:embed resource/nuhuh.gif
var nuhuhGif []byte

//go:embed resource/desuwa.gif
var desuwaGif []byte

//go:embed resource/miemebell.json
var miemebellJson []byte

// StartCommands registers the commands with Discord and sets up the listener
func (d *DiscordNotifier) StartCommands() error {
	// 1. Define Commands
	// Constants for limits
	minLimit := 1.0
	maxLimit := 1000.0 // Cap at 25 to prevent exceeding Discord's 2000 char limit
	// Helper to build Track Choices (Pre-defined list)
	// Discord limits choices to 25. You have exactly 24 tracks in the map above. Perfect.
	trackChoices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, name := range idac_domain.GetTrackNames() {
		trackChoices = append(trackChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: name,
		})
	}
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show current bot status",
		},
		{
			Name:        "miemebell",
			Description: "A special-flavored tea.",
		},
		{
			Name:        "nuhuh",
			Description: "Reply with the nuh uh gif",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to tag",
					Required:    false,
				},
			},
		},
		{
			Name:        "mckween",
			Description: "desuwa",
		},
		{
			Name:        "idac",
			Description: "Get Initial D Arcade Rankings",
			Options: []*discordgo.ApplicationCommandOption{
				// Subcommand: Time Attack
				{
					Name:        "time-attack",
					Description: "Get Time Attack Rankings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "track", // Changed from 'course'
							Description: "Select the track",
							Required:    true,
							Choices:     trackChoices, // User picks from list
						},
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "variant", // New field
							Description:  "Select direction/condition (Downhill, Uphill, etc)",
							Required:     true,
							Autocomplete: true, // Triggers dynamic behavior
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area",
							Description: "Area ID or Alias (e.g. 'vn', 'area-57')",
							Required:    true, // Strictly required for TA
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "car",
							Description: "Car Option (default: car-all)",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "spec",
							Description: "Car spec variant (AR, HC, etc)",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "limit",
							Description: "Number of results to show (1-25, default: 10)",
							Required:    false,
							MinValue:    &minLimit,
							MaxValue:    maxLimit,
						},
					},
				},
				// Subcommand: Team
				{
					Name:        "team",
					Description: "Get Team Rankings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "rank",
							Description: "Team Rank Class",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "All", Value: "all"}, // New "All" option
								{Name: "Master", Value: "6"},
								{Name: "Platinum", Value: "5"},
								{Name: "Gold", Value: "4"},
								{Name: "Silver", Value: "3"},
								//{Name: "Bronze", Value: "2"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "country",
							Description: "Filter by Country/Area (e.g. 'vn', 'jp')",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "limit",
							Description: "Number of results to show (1-25, default: 10)",
							Required:    false,
							MinValue:    &minLimit,
							MaxValue:    maxLimit,
						},
					},
				},
				// Subcommand: Player Info
				{
					Name:        "player-info",
					Description: "Find a player's rank on a specific course",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ign",
							Description: "Player Name (Case Insensitive)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "course",
							Description: "Course ID or Alias",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area",
							Description: "Area ID or Alias",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "compare",
							Description: "Compare scope",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Local", Value: "local"},
								{Name: "World", Value: "world"},
							},
						},
					},
				},
				// Subcommand: Player Compare Subcommand
				{
					Name:        "player-compare",
					Description: "Compare times between two players on a specific course",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "track",
							Description: "Select the track",
							Required:    true,
							Choices:     trackChoices, // Reuses the list from time-attack
						},
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "variant",
							Description:  "Select direction/condition",
							Required:     true,
							Autocomplete: true, // Reuses the autocomplete handler
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ign1",
							Description: "First Player Name",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area1",
							Description: "Area for Player 1 (e.g. 'vn')",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "ign2",
							Description: "Second Player Name",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area2",
							Description: "Area for Player 2 (Optional, defaults to Area 1)",
							Required:    false,
						},
					},
				},
			},
		},
	}

	// 2. Define Handlers
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// --- IDAC Handler ---
		"idac": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// The top-level options array contains the Subcommand as the first item
			options := i.ApplicationCommandData().Options
			if len(options) == 0 {
				return
			}
			subcommand := options[0]
			// Parse Nested Options inside the subcommand
			optMap := make(map[string]string)
			specInput := ""
			// Robust Option Parsing based on Type
			for _, opt := range subcommand.Options {
				switch opt.Type {
				case discordgo.ApplicationCommandOptionString:
					if opt.Name == "spec" {
						specInput = opt.StringValue()
					} else {
						optMap[opt.Name] = opt.StringValue()
					}
				case discordgo.ApplicationCommandOptionInteger:
					// Convert integers to string for map storage
					val := strconv.FormatInt(opt.IntValue(), 10)
					if opt.Name == "spec" {
						// Handle cached/legacy spec as integer
						specInput = val
					} else {
						// Handles "limit" and any other future ints
						optMap[opt.Name] = val
					}
				}
			}
			// Dispatch based on Subcommand Name
			switch subcommand.Name {
			case "time-attack":
				// Pass the parsed map and extra spec input
				handleTimeAttack(s, i, optMap, specInput)
			case "team":
				handleTeamRanking(s, i, optMap)
			case "player-info":
				handlePlayerInfo(s, i, optMap)
			case "player-compare":
				handlePlayerCompare(s, i, optMap)
			}
		},
		"status": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "I am running and monitoring your feeds! 📡",
				},
			})
		},
		"nuhuh": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// 1. DEFER: Buy 15 minutes of processing time
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Brewing tea...",
				},
			})
			// 2. LOGIC: Parse Options
			options := i.ApplicationCommandData().Options
			var content string
			for _, opt := range options {
				if opt.Name == "user" {
					targetID := opt.Value.(string)
					if targetID == s.State.User.ID {
						content = "Nice try, but ***nuh-uh***"
					} else {
						content = fmt.Sprintf("<@%s>", targetID)
					}
				}
			}

			// 4. EDIT: Attach File and Content to the original deferred message
			// Note: If content is empty, pass nil to avoid overwriting existing (though here it's fresh)
			var contentPtr *string
			if content != "" {
				contentPtr = &content
			}

			// 3. EDIT: Send embedded file
			// Using bytes.NewReader(nuhuhGif) is extremely fast as it's just memory reading
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: contentPtr,
				Files: []*discordgo.File{
					{Name: "nuhuh.gif", Reader: bytes.NewReader(nuhuhGif)},
				},
			})
		},
		"mckween": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Desuwa~",
					Files: []*discordgo.File{
						{
							Name:   "desuwa.gif",
							Reader: bytes.NewReader(desuwaGif),
						},
					},
				},
			})
		},
		"miemebell": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// 3. Unmarshal the embedded data into the struct
			var data domain.MieMeBell
			err := json.Unmarshal(miemebellJson, &data)
			if err != nil {
				log.Fatal("Error parsing JSON:", err)
			}
			randomIndex := rand.IntN(len(data.Blocks))
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: data.Blocks[randomIndex],
				},
			})
		},
	}

	// 3. Define Autocomplete Handler
	// This function figures out which variants to show based on the selected track
	autocompleteHandler := func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		data := i.ApplicationCommandData()
		if data.Name != "idac" {
			return
		}

		// Find the subcommand options
		// structure: idac -> [time-attack] -> [track, variant, area...]
		if len(data.Options) == 0 {
			return
		}
		subCmd := data.Options[0]

		var selectedTrack string

		// 1. Find what the user has currently selected for "track"
		for _, opt := range subCmd.Options {
			if opt.Name == "track" {
				selectedTrack = opt.StringValue()
			}
		}

		// 2. Generate choices for "variant"
		var choices []*discordgo.ApplicationCommandOptionChoice

		if variants, ok := idac_domain.TrackRegistry[selectedTrack]; ok {
			for _, v := range variants {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  v.Name, // Display: "Downhill"
					Value: v.ID,   // Value passed to handler: "course-12"
				})
			}
		} else {
			// Default hint if no track selected yet
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  "Select a track first",
				Value: "none",
			})
		}

		// 3. Send choices back to Discord
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionApplicationCommandAutocompleteResult,
			Data: &discordgo.InteractionResponseData{
				Choices: choices,
			},
		})
	}

	// 4. Register Router
	d.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			autocompleteHandler(s, i)
		}
	})

	// 4. Register Commands with Discord API (Bulk Overwrite)
	// Note: We use empty string "" for GuildID to register globally.
	// Global commands take up to 1 hour to update. For instant testing, put your Guild ID there.
	log.Println("Registering commands...")
	createdCommands, err := d.session.ApplicationCommandBulkOverwrite(d.session.State.User.ID, "", commands)
	if err == nil {
		for _, createdCommand := range createdCommands {
			fmt.Println("Created command:", createdCommand.Name)
		}
	}
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func handleTimeAttack(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string, specInput string) {
	var err error
	// 1. DEFER
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	sendDeferredError := func(msg string) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
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
	baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
	//filename := fmt.Sprintf("ta_%s_%s_%s.json", optMap["course"], optMap["area"], finalCarID)
	filename := fmt.Sprintf("ta_%s_%s_%s.json", finalCourseID, optMap["area"], finalCarID)
	fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)

	resp, err := http.Get(fullURL)
	if err != nil {
		sendDeferredError("❌ Error contacting SEGA server.")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		sendDeferredError(fmt.Sprintf("❌ Failed to fetch data (Status: %d).", resp.StatusCode))
		return
	}

	body, _ := io.ReadAll(resp.Body)
	var data domain.IdacResponse
	if err := json.Unmarshal(body, &data); err != nil {
		sendDeferredError("❌ Error parsing JSON.")
		return
	}

	if len(data.Records) == 0 {
		msg := fmt.Sprintf("# Initial D Rankings (Time Trial)\n🗾 : %s | 🌎 : %s | 🚗 : %s\n\nNo records found.", courseName, areaName, carDisplayName)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(data.Records) < limit {
		limit = len(data.Records)
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
		r := data.Records[j]
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
	sendPagination(s, i, pages)
}

func handleTeamRanking(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER INTERACTION
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Error Helper
	sendDeferredError := func(msg string) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
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
		records, err := fetchTeamRankings(roundNum, code)
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
	sendPagination(s, i, pages)
}

// Helper to fetch records for a specific round/rank
func fetchTeamRankings(roundNum int, rankCode string) ([]domain.TeamRecord, error) {
	rankURL := fmt.Sprintf("https://initiald.sega.jp/inidac/json/ranking/v1/leaguePoint/lp-round-%d_rank-%s.json", roundNum, rankCode)
	fmt.Printf("Fetching ranking data from %s\n", rankURL)
	resp, err := http.Get(rankURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var data domain.TeamRankingResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	foundTeams := make([]domain.TeamRecord, 0)
	// set league emoji values for each team
	for _, foundTeam := range data.Records {
		foundTeam.LeagueEmoji = domain.TeamLeagueEmojis[rankCode]
		foundTeams = append(foundTeams, foundTeam)
	}
	return foundTeams, nil
}

func handlePlayerInfo(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string) {
	// 0. DEFER
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// 1. Validate & Parse Inputs
	ign, hasIgn := optMap["ign"]
	if !hasIgn || optMap["course"] == "" || optMap["area"] == "" {
		errStr := "⚠️ Missing required arguments (ign, course, or area)."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	// --- NORMALIZE INPUT ---
	// 1. Normalize width (ＳＨＩＲＯ -> SHIRO)
	// 2. Lowercase (SHIRO -> shiro) for case-insensitive comparison
	targetName := strings.ToLower(domain.NormalizeTextWidth(ign))
	// Resolve Aliases
	courseInput := strings.ToLower(optMap["course"])
	if val, ok := domain.CourseAliases[courseInput]; ok {
		courseInput = val
	}
	areaInput := strings.ToLower(optMap["area"])
	if val, ok := domain.AreaAliases[areaInput]; ok {
		areaInput = val
	}

	// 2. Construct URL
	// We strictly use "car-all" to find the player regardless of what car they drove
	baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
	filename := fmt.Sprintf("ta_%s_%s_%s.json", courseInput, areaInput, "car-all")
	fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)

	// 3. Fetch Data
	resp, err := http.Get(fullURL)
	if err != nil {
		errStr := "❌ Network error contacting SEGA."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errStr := fmt.Sprintf("❌ Failed to fetch ranking data (Status: %d). Check your course/area codes.", resp.StatusCode)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}

	// 4. Parse JSON
	var data domain.IdacResponse
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		errStr := "❌ Error parsing SEGA data."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &errStr})
		return
	}

	// 5. Search for Player
	var foundRecord *domain.TimeAttackRecord
	var leaderRecord *domain.TimeAttackRecord

	for idx, r := range data.Records {
		// Capture Leader
		if idx == 0 {
			val := r
			leaderRecord = &val
		}

		// --- NORMALIZE RECORD NAME ---
		// We normalize the record name from SEGA too, just in case
		recordNameNormalized := strings.ToLower(domain.NormalizeTextWidth(r.Name))

		if recordNameNormalized == targetName {
			val := r
			foundRecord = &val
			break
		}
	}

	// 6. Build Response
	var sb strings.Builder

	// Display Names
	courseName := domain.CourseDisplayNameByCode[courseInput]
	if courseName == "" {
		courseName = courseInput
	}

	areaName := domain.AreaDisplayNameByCode[areaInput]
	if areaName == "" {
		areaName = areaInput
	}

	sb.WriteString(fmt.Sprintf("**Player Info**\n"))
	sb.WriteString(fmt.Sprintf("**In-game Name:** `%s` | **Course:** %s | **Area:** %s\n\n", ign, courseName, areaName))

	if foundRecord != nil {
		// --- Calculate Delta ---
		deltaStr := ""
		if leaderRecord != nil && foundRecord.Name != leaderRecord.Name {
			playerMs, err1 := domain.ParseIdacTime(foundRecord.Record)
			leaderMs, err2 := domain.ParseIdacTime(leaderRecord.Record)

			// Only calculate if parsing succeeded
			if err1 == nil && err2 == nil {
				diff := playerMs - leaderMs
				deltaStr = fmt.Sprintf(" (%s)", domain.FormatIdacTimeDelta(diff))
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
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &finalContent,
	})
}

func handlePlayerCompare(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string) {
	// 1. DEFER
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	sendDeferredError := func(msg string) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
	}

	// 2. Validate Inputs
	if optMap["ign1"] == "" || optMap["ign2"] == "" || optMap["variant"] == "" || optMap["area1"] == "" {
		sendDeferredError("⚠️ **Missing Arguments**\nPlease provide both IGNs, Track, Variant, and at least Area 1.")
		return
	}

	// 3. Resolve Areas
	resolveArea := func(input string) string {
		normalized := strings.ToLower(input)
		// Explicitly check for 'all' keyword if missing from map, just in case
		if normalized == "world" || normalized == "global" {
			return "all"
		}

		if val, ok := domain.AreaAliases[normalized]; ok {
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

	// 4. Helper: Fetch Data & Find Player
	fetchAndFind := func(areaCode, targetIgn string) (*domain.TimeAttackRecord, string, error) {
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

		var data domain.IdacResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, "Parse Error", nil
		}

		// Find Player
		normalizedTarget := strings.ToLower(domain.NormalizeTextWidth(targetIgn))
		for _, r := range data.Records {
			if strings.ToLower(domain.NormalizeTextWidth(r.Name)) == normalizedTarget {
				val := r
				return &val, "", nil
			}
		}
		return nil, "", nil
	}

	// 5. Execution
	var p1, p2 *domain.TimeAttackRecord
	var errStr1, errStr2 string

	// OPTIMIZATION: If areas are identical (e.g. both 'all'), fetch once
	if area1 == area2 {
		baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
		filename := fmt.Sprintf("ta_%s_%s_%s.json", courseID, area1, "car-all")
		fullURL := fmt.Sprintf("%s/timeTrial/%s", baseURL, filename)
		resp, err := http.Get(fullURL)
		if err == nil && resp.StatusCode == 200 {
			var data domain.IdacResponse
			_ = json.NewDecoder(resp.Body).Decode(&data)
			resp.Body.Close()

			target1 := strings.ToLower(domain.NormalizeTextWidth(optMap["ign1"]))
			target2 := strings.ToLower(domain.NormalizeTextWidth(optMap["ign2"]))

			for _, r := range data.Records {
				if p1 != nil && p2 != nil {
					break
				}
				normName := strings.ToLower(domain.NormalizeTextWidth(r.Name))
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

	courseName := domain.CourseDisplayNameByCode[courseID]
	if courseName == "" {
		courseName = optMap["track"]
	}

	areaName1 := domain.AreaDisplayNameByCode[area1]
	if areaName1 == "" {
		areaName1 = area1
	}

	areaName2 := domain.AreaDisplayNameByCode[area2]
	if areaName2 == "" {
		areaName2 = area2
	}

	sb.WriteString(fmt.Sprintf("# Player Comparison\n"))
	sb.WriteString(fmt.Sprintf("**Course:** %s\n", courseName))

	// Helper to print
	printPlayer := func(label, areaName, inputName, errStr string, p *domain.TimeAttackRecord) {
		sb.WriteString(fmt.Sprintf("### %s (%s): ", label, areaName))
		if p != nil {
			sb.WriteString(fmt.Sprintf("**%s**\n", p.Name))
			sb.WriteString(fmt.Sprintf("- **Local Rank:** #%s\n", p.Rank))
			sb.WriteString(fmt.Sprintf("- **Time:** `%s`\n", p.Record))
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
		sb.WriteString("\n---\n")
		ms1, err1 := domain.ParseIdacTime(p1.Record)
		ms2, err2 := domain.ParseIdacTime(p2.Record)

		if err1 == nil && err2 == nil {
			diff := ms1 - ms2
			if diff < 0 {
				gap := domain.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p1.Name, strings.TrimPrefix(gap, "-")))
			} else if diff > 0 {
				gap := domain.FormatIdacTimeDelta(diff)
				sb.WriteString(fmt.Sprintf("🏆 **%s** is faster by **%s**!", p2.Name, strings.TrimPrefix(gap, "+")))
			} else {
				sb.WriteString("🤝 **It's a tie!** Exact same time.")
			}
		}
	}

	finalContent := sb.String()
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &finalContent,
	})
}

// Helper to send ephemeral error messages
func sendError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "⚠️ " + msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// sendPagination sends a paginated message with Next/Prev buttons
func sendPagination(s *discordgo.Session, i *discordgo.InteractionCreate, pages []string) {
	// If only 1 page, just send it without buttons
	if len(pages) == 1 {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &pages[0],
		})
		return
	}

	// Current Page Index
	pageIndex := 0

	// Helper to create buttons based on current page
	getComponents := func(current int) []discordgo.MessageComponent {
		prevDisabled := current == 0
		nextDisabled := current == len(pages)-1

		return []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "◀️ Previous",
						Style:    discordgo.PrimaryButton,
						CustomID: "pagination_prev",
						Disabled: prevDisabled,
					},
					discordgo.Button{
						Label:    fmt.Sprintf("Page %d/%d", current+1, len(pages)),
						Style:    discordgo.SecondaryButton,
						CustomID: "pagination_status",
						Disabled: true, // Just a label
					},
					discordgo.Button{
						Label:    "Next ▶️",
						Style:    discordgo.PrimaryButton,
						CustomID: "pagination_next",
						Disabled: nextDisabled,
					},
				},
			},
		}
	}

	// 1. Send the FIRST page
	components := getComponents(pageIndex)
	msg, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &pages[0],
		Components: &components,
	})
	if err != nil {
		return
	}

	// 2. Register a Handler for Button Clicks
	// We use a closure so we can access 'pageIndex' and 'pages' safely
	// We also need a "stop" channel to kill the listener after timeout
	stop := make(chan struct{})

	// Create the handler function
	cleanup := s.AddHandler(func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		// Filter: Must be Button click, match Message ID, and match User ID (optional security)
		if ic.Type != discordgo.InteractionMessageComponent ||
			ic.Message.ID != msg.ID ||
			ic.Member.User.ID != i.Member.User.ID {
			return
		}

		// Handle Buttons
		switch ic.MessageComponentData().CustomID {
		case "pagination_prev":
			if pageIndex > 0 {
				pageIndex--
			}
		case "pagination_next":
			if pageIndex < len(pages)-1 {
				pageIndex++
			}
		}

		// Update the Message
		newComps := getComponents(pageIndex)
		s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    pages[pageIndex],
				Components: newComps,
			},
		})
	})

	// 3. Cleanup Routine (Timeout after 2 minutes)
	// We run this in a goroutine so we don't block
	go func() {
		select {
		case <-time.After(2 * time.Minute):
			// Remove buttons after timeout
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Components: &[]discordgo.MessageComponent{}, // Empty components clears them
			})
			cleanup() // Remove the event handler
		case <-stop:
			cleanup()
		}
	}()
}
