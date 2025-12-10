package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/internal/domain"
)

// StartCommands registers the commands with Discord and sets up the listener
func (d *DiscordNotifier) StartCommands() error {
	// 1. Define Commands
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "status",
			Description: "Show current bot status",
		},
		{
			Name:        "nuhuh",
			Description: "Your opinion is in the trash",
		},
		{
			Name:        "mckween",
			Description: "desuwa",
		},
		{
			Name:        "idac",
			Description: "Get Initial D Arcade Rankings",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "mode",
					Description: "Game Mode (e.g. 'ta')",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "course",
					Description: "Course ID (e.g. 'course-16')",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "area",
					Description: "Area ID (e.g. 'area-57')",
					Required:    true,
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
					Description: "Spec Variant (AR, HC, DH) - Ignored if car is empty",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionNumber,
					Name:        "limit",
					Description: "Result limit (default: 10)",
					Required:    false,
				},
			},
		},
	}

	// 2. Define Handlers
	handlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// --- IDAC Handler ---
		"idac": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// 1. Parse Options
			options := i.ApplicationCommandData().Options
			optMap := make(map[string]interface{})
			specInput := ""

			for _, opt := range options {
				if opt.Name == "spec" {
					specInput = opt.StringValue()
				} else {
					optMap[opt.Name] = opt.StringValue()
				}
			}

			// Defaults
			if _, ok := optMap["car"]; !ok {
				optMap["car"] = "car-all"
			}

			// --- 1.5 RESOLVE ALIASES (Using domain package) ---
			courseInput := strings.ToLower(optMap["course"].(string))
			if val, ok := domain.CourseAliases[courseInput]; ok {
				optMap["course"] = val
			}

			areaInput := strings.ToLower(optMap["area"].(string))
			if val, ok := domain.AreaAliases[areaInput]; ok {
				optMap["area"] = val
			}

			// Resolve Car with Spec (New Logic)
			finalCarID := domain.ResolveCarID(optMap["car"].(string), specInput)

			// Resolve Display Names
			courseName := optMap["course"].(string)
			if val, ok := domain.CourseDisplayNameByCode[courseName]; ok {
				courseName = val
			}

			areaName := optMap["area"].(string)
			if val, ok := domain.AreaDisplayNameByCode[areaName]; ok {
				areaName = val
			}

			carDisplayName := finalCarID
			baseCar := domain.ResolveCarID(optMap["car"].(string), "")
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
				limit = int(optMap["limit"].(float64))
			} else {
				limit = 10
			}

			// 2. Construct URL
			modeFolder := "timeTrial"
			if optMap["mode"] != "ta" {
				modeFolder = optMap["mode"].(string)
			}

			baseURL := "https://initiald.sega.jp/inidac/json/ranking/v1"
			filename := fmt.Sprintf("%s_%s_%s_%s.json", optMap["mode"], optMap["course"], optMap["area"], finalCarID)
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
			// 5. Build LIST Message (Clean & Mobile Friendly)
			var sb strings.Builder

			// Header Information
			sb.WriteString(fmt.Sprintf("# Initial D Rankings (%s)\n", optMap["mode"]))
			sb.WriteString(fmt.Sprintf("🗾 : %s |  🌎 : %s |  🚗 : %s\n\n",
				domain.CourseDisplayNameByCode[optMap["course"].(string)],
				domain.AreaDisplayNameByCode[optMap["area"].(string)],
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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "https://c.tenor.com/SCfWfZvA8_0AAAAd/tenor.gif",
				},
			})
		},

		"mckween": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			resp, err := http.Get("https://c.tenor.com/9HXPPljXLxUAAAAC/tenor.gif")
			if err != nil {
				log.Printf("Error fetching wink gif: %v", err)
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Desuwa~",
					Files: []*discordgo.File{
						{
							Name:   "desuwa.gif",
							Reader: resp.Body,
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
