package discord

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain"
)

//go:embed resource/nuhuh.gif
var nuhuhGif []byte

//go:embed resource/desuwa.gif
var desuwaGif []byte

// StartCommands registers the commands with Discord and sets up the listener
func (d *DiscordNotifier) StartCommands() error {
	// 1. Define Commands
	// Constants for limits
	minLimit := 1.0
	maxLimit := 1000.0 // Cap at 25 to prevent exceeding Discord's 2000 char limit

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show current bot status",
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
							Name:        "course",
							Description: "Course ID or Alias (e.g. 'iro', 'course-16')",
							Required:    true, // Strictly required for TA
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
							Description: "Spec Variant (AR, HC, etc)",
							Required:    false,
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
			})

			// 2. WORK: Download the GIF
			resp, err := http.Get("https://c.tenor.com/SCfWfZvA8_0AAAAd/tenor.gif")
			if err != nil {
				errStr := "❌ Failed to retrieve GIF (Network Error)."
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errStr := fmt.Sprintf("❌ Failed to retrieve GIF (Status: %d).", resp.StatusCode)
				s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errStr,
				})
				return
			}

			// 3. LOGIC: Parse Options
			options := i.ApplicationCommandData().Options
			var content string
			for _, opt := range options {
				if opt.Name == "user" {
					targetID := opt.Value.(string)
					if targetID == s.State.User.ID {
						content = "Nice try, but **nuh-uh**"
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
	}

	// 3. Register Handler to Router
	d.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// 4. Register Commands with Discord API (Bulk Overwrite)
	// Note: We use empty string "" for GuildID to register globally.
	// Global commands take up to 1 hour to update. For instant testing, put your Guild ID there.
	log.Println("Registering commands...")
	_, err := d.session.ApplicationCommandBulkOverwrite(d.session.State.User.ID, "", commands)
	return err
}

func handleTimeAttack(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string, specInput string) {
	var err error
	// 0. Validate Inputs (Since they are now optional at command level)
	if optMap["course"] == "" || optMap["area"] == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "⚠️ **Missing Arguments**\nFor Time Attack (`ta`) mode, you must provide a `course` and an `area`.",
			},
		})
		return
	}
	// Defaults
	if _, ok := optMap["car"]; !ok {
		optMap["car"] = "car-all"
	}

	// --- 1.5 RESOLVE ALIASES (Using domain package) ---
	courseInput := strings.ToLower(optMap["course"])
	if val, ok := domain.CourseAliases[courseInput]; ok {
		optMap["course"] = val
	}

	areaInput := strings.ToLower(optMap["area"])
	if val, ok := domain.AreaAliases[areaInput]; ok {
		optMap["area"] = val
	}

	// Resolve Car with Spec (New Logic)
	finalCarID := domain.ResolveCarID(optMap["car"], specInput)

	// Resolve Display Names
	courseName := optMap["course"]
	if val, ok := domain.CourseDisplayNameByCode[courseName]; ok {
		courseName = val
	}

	areaName := optMap["area"]
	if val, ok := domain.AreaDisplayNameByCode[areaName]; ok {
		areaName = val
	}

	carDisplayName := finalCarID
	baseCar := domain.ResolveCarID(optMap["car"], "")
	if val, ok := domain.CarDisplayNameByCode[baseCar]; ok {
		if emoji, ok := domain.SpecEmojis[strings.ToLower(specInput)]; ok {
			carDisplayName = fmt.Sprintf("%s %s", emoji, val)
		}
	}
	if optMap["car"] == "car-all" {
		carDisplayName = "All"
	}
	// Check limit Alias
	var limit int
	if _, ok := optMap["limit"]; ok {
		limit, err = strconv.Atoi(optMap["limit"])
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "⚠️ Limit must be a number",
				},
			})
			return
		}
	} else {
		limit = 10
	}

	// 2. Construct URL
	modeFolder := "timeTrial"
	//if optMap["mode"] != "ta" {
	//	modeFolder = optMap["mode"]
	//}

	baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
	filename := fmt.Sprintf("%s_%s_%s_%s.json", "ta", optMap["course"], optMap["area"], finalCarID)
	fullURL := fmt.Sprintf("%s/%s/%s", baseURL, modeFolder, filename)

	// 3. Fetch Data
	resp, err := http.Get(fullURL)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "Error contacting SEGA server."},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Failed to fetch data. Status: %d\nURL tried: `%s`", resp.StatusCode, fullURL),
			},
		})
		return
	}

	// 4. Parse JSON
	body, _ := io.ReadAll(resp.Body)
	var data domain.IdacResponse
	if err := json.Unmarshal(body, &data); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: "Error parsing JSON data."},
		})
		return
	}
	fmt.Println("full url:", fullURL)
	// 5. Build LIST 1 (Clean & Mobile Friendly)
	var sb strings.Builder
	// Header Information
	sb.WriteString(fmt.Sprintf("# Initial D Rankings (%s)\n", "Time Trial"))
	sb.WriteString(fmt.Sprintf("🗾 : %s |  🌎 : %s |  🚗 : %s\n\n",
		domain.CourseDisplayNameByCode[optMap["course"]],
		domain.AreaDisplayNameByCode[optMap["area"]],
		carDisplayName,
	))
	if len(data.Records) == 0 {
		sb.WriteString("No records found.")
	} else {
		if len(data.Records) < limit {
			limit = len(data.Records)
		}

		for j := 0; j < limit; j++ {
			r := data.Records[j]

			// Compact Single-Line Format:
			// **1. DriverName** (CarName) — `Time`
			sb.WriteString(fmt.Sprintf("%s. **%s** — %s — `%s`\n", r.Rank, r.Name, r.CarName, r.Record))

		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: sb.String(),
			},
		})
	}
}

func handleTeamRanking(s *discordgo.Session, i *discordgo.InteractionCreate, optMap map[string]string) {
	//0. DEFER INTERACTION: Acknowledge immediately to avoid 3s timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return
	}
	// Error Helper for Deferred Response
	//fail := func(msg string) {
	//	str := "⚠️ " + msg
	//	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
	//		Content: &str,
	//	})
	//}

	// 1. Fetch Current Round Info
	roundURL := "https://initiald.sega.jp/inidac/json/ranking/v1/currentRoundInfo.json"
	respRound, err := http.Get(roundURL)
	if err != nil {
		sendError(s, i, "Error fetching round info.")
		return
	}
	defer respRound.Body.Close()

	bodyBytes, err := io.ReadAll(respRound.Body)
	if err != nil {
		sendError(s, i, "Error reading round info.")
		return
	}

	roundStr := strings.TrimSpace(string(bodyBytes))
	roundNum, err := strconv.Atoi(roundStr)
	if err != nil {
		sendError(s, i, fmt.Sprintf("Error parsing round number: %s", roundStr))
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

	// 4. Determine Target Ranks (Single or All)
	rankCodeInput := optMap["rank"]
	targetRanks := []string{}

	if rankCodeInput == "all" {
		// All ranks: Master(6) down to Bronze(2)
		targetRanks = []string{"6", "5", "4", "3"}
	} else {
		targetRanks = []string{rankCodeInput}
	}

	// 5. Fetch and Aggregate Records
	var allRecords []domain.TeamRecord
	for _, code := range targetRanks {
		records, err := fetchTeamRankings(roundNum, code)
		if err != nil {
			if len(targetRanks) == 1 {
				sendError(s, i, fmt.Sprintf("Error fetching data for rank %s: %v", code, err))
				return
			}
			continue // Skip this rank if one fails in "all" mode
		}
		allRecords = append(allRecords, records...)
	}

	// 6. Sort Combined Records by Points (Descending)
	sort.Slice(allRecords, func(i, j int) bool {
		p1, _ := strconv.Atoi(allRecords[i].Point)
		p2, _ := strconv.Atoi(allRecords[j].Point)
		return p1 > p2
	})

	// 7. Filter and Build Message
	var sb strings.Builder
	// Filter and Build Messages (Splitting)
	var messages []string
	var currentMessage strings.Builder

	rankDisplayName := "Unknown"
	if rankCodeInput == "all" {
		rankDisplayName = "All Classes"
	} else {
		switch rankCodeInput {
		case "6":
			rankDisplayName = "Master"
		case "5":
			rankDisplayName = "Platinum"
		case "4":
			rankDisplayName = "Gold"
		case "3":
			rankDisplayName = "Silver"
			//case "2":
			//	rankDisplayName = "Bronze"
		}
	}

	sb.WriteString(fmt.Sprintf("# Initial D Team Rankings (Round %d)\n", roundNum))
	sb.WriteString(fmt.Sprintf("**Class:** %s | **Region:** %s\n", rankDisplayName, filterCountryName))

	foundCount := 0
	for _, r := range allRecords {
		// Apply Filter
		if filterCountryID != -1 && r.Country != filterCountryID {
			continue
		}
		if foundCount >= limit {
			break
		}
		// Calculate display rank. If "All" mode, r.Rank is just the rank within its class.
		// We can use (foundCount + 1) as the global rank in this sorted view.
		globalRank := foundCount + 1
		countryFlag := domain.GetCountryFlag(r.Country)

		entry := fmt.Sprintf("%d. %s **%s** \n", globalRank, r.LeagueEmoji, r.TeamName)
		entry += fmt.Sprintf("+ **Country:** %s\n", countryFlag)
		entry += fmt.Sprintf("+ **Points:** %s\n", r.Point)
		entry += fmt.Sprintf("+ **Ace:** %s | **Leader:** %s\n\n", r.AceUserName, r.LeaderUserName)
		if currentMessage.Len()+len(entry) > 1900 {
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
		}
		currentMessage.WriteString(entry)
		foundCount++
	}
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}
	// Send Results
	if len(messages) == 0 {
		noRes := "No teams found matching criteria."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &noRes,
		})
		return
	}
	// 1. Edit the deferred message with the FIRST chunk
	firstChunk := messages[0]
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &firstChunk,
	})

	// 2. Send remaining chunks as Followups
	for _, msg := range messages[1:] {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: msg,
		})
	}
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
