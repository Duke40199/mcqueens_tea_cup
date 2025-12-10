package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

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
		// --- IDAC Handler ---
		"idac": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			// 1. Parse Options
			options := i.ApplicationCommandData().Options
			optMap := make(map[string]interface{})
			for _, opt := range options {
				optMap[opt.Name] = opt.Value
			}

			// Defaults
			if _, ok := optMap["car"]; !ok {
				optMap["car"] = "car-all"
			}

			// --- 1.5 RESOLVE ALIASES ---
			// Check Course Alias
			courseInput := strings.ToLower(optMap["course"].(string))
			if val, ok := domain.CourseAliases[courseInput]; ok {
				optMap["course"] = val
			}

			// Check Area Alias
			areaInput := strings.ToLower(optMap["area"].(string))
			if val, ok := domain.AreaAliases[areaInput]; ok {
				optMap["area"] = val
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
			filename := fmt.Sprintf("%s_%s_%s_%s.json", optMap["mode"], optMap["course"], optMap["area"], optMap["car"])
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

			// 5. Generate Spacer Image (300x1 Transparent)
			// This forces the embed to be wider on mobile without needing external URLs.
			spacerImg := image.NewRGBA(image.Rect(0, 0, 300, 1))
			// (Optional) Fill with transparent color, though NewRGBA defaults to zero-value (transparent black)
			for x := 0; x < 300; x++ {
				spacerImg.Set(x, 0, color.RGBA{0, 0, 0, 0})
			}
			var imgBuf bytes.Buffer
			if err := png.Encode(&imgBuf, spacerImg); err != nil {
				log.Printf("Failed to encode spacer image: %v", err)
			}

			// 6. Build Embed Fields
			embed := &discordgo.MessageEmbed{
				Title: fmt.Sprintf("Initial D Rankings (%s)", optMap["mode"]),
				Description: fmt.Sprintf("Course: %s | Area: %s",
					domain.CourseDisplayNameByCode[optMap["course"].(string)], domain.AreaDisplayNameByCode[optMap["area"].(string)]),
				Color:     0xEE0000,
				URL:       fullURL,
				Timestamp: time.Now().Format(time.RFC3339),
				// Link the image to the attachment
				Image: &discordgo.MessageEmbedImage{
					URL: "attachment://spacer.png",
				},
			}

			var files []*discordgo.File
			// Add the spacer file to the response payload
			if imgBuf.Len() > 0 {
				files = append(files, &discordgo.File{
					Name:   "spacer.png",
					Reader: &imgBuf,
				})
			}

			if len(data.Records) == 0 {
				embed.Description += "\n\nNo records found."
			} else {
				if len(data.Records) < limit {
					limit = len(data.Records)
				}

				var rankCol, infoCol, timeCol string

				for j := 0; j < limit; j++ {
					r := data.Records[j]

					name := r.Name
					if len(name) > 20 {
						name = name[:12] + ".."
					}

					car := r.CarName
					if len(car) > 20 {
						car = car[:12] + ".."
					}

					rankCol += fmt.Sprintf("**%s**\n\n", r.Rank)
					infoCol += fmt.Sprintf("%s\n%s\n", name, car)
					timeCol += fmt.Sprintf("%s\n\n", r.Record)
				}

				embed.Fields = []*discordgo.MessageEmbedField{
					{Name: "Rank", Value: rankCol, Inline: true},
					{Name: "Driver / Car", Value: infoCol, Inline: true},
					{Name: "Time", Value: timeCol, Inline: true},
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{embed},
					Files:  files,
				},
			})
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
