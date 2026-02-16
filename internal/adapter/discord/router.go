package discord

import (
	_ "embed"
	"fmt"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"

	idac_domain "McQueens_Tea_Cup/internal/domain/entity"
)

//go:embed resource/nuhuh.gif
var nuhuhGif []byte

//go:embed resource/desuwa.gif
var desuwaGif []byte

//go:embed resource/miemebell.json
var miemebellJson []byte

//go:embed resource/meomeo.json
var meomeoJson []byte

var languageChoice = []*discordgo.ApplicationCommandOptionChoice{
	{
		Name:  "Vietnamese",
		Value: "vn",
	},
	{
		Name:  "English",
		Value: "en",
	},
	{
		Name:  "Desuwa",
		Value: "desuwa",
	},
}

func (h *Handler) RegisterCommands() error {
	// 1. Get Bot ID explicitly
	// We use User("@me") because h.Session.State.User is nil until Open() is called.
	u, err := h.Session.User("@me")
	if err != nil {
		return fmt.Errorf("failed to fetch bot ID: %w", err)
	}
	botID := u.ID

	minLimit := 1.0
	maxLimit := 1000.0

	trackChoices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, name := range idac_domain.GetTrackNames() {
		trackChoices = append(trackChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  name,
			Value: name,
		})
	}

	// 2. Define All Commands
	commands := []*discordgo.ApplicationCommand{
		// --- IDAC Complex Command ---
		{
			Name:        "idac",
			Description: "Commands related to the game Initial D The Arcade",
			Options: []*discordgo.ApplicationCommandOption{
				// Subcommand: Time Attack
				{
					Name:        "time-attack",
					Description: "Get Time Attack Rankings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "track",
							Description: "Select the track",
							Required:    true,
							Choices:     trackChoices,
						},
						{
							Type:         discordgo.ApplicationCommandOptionString,
							Name:         "variant",
							Description:  "Select direction/condition (Downhill, Uphill, etc)",
							Required:     true,
							Autocomplete: true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area",
							Description: "Area ID or Alias (e.g. 'vn', 'area-57')",
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
							Description: "Get Team Class & Rankings",
							Required:    true,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "All", Value: "all"},
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
				// Player Compare Subcommand
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
							Name:        "player1",
							Description: "Player 1 (IGN or @User)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "player2",
							Description: "Player 2 (IGN or @User)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area1",
							Description: "Area for Player 1 (e.g. 'vn')",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area2",
							Description: "Area for Player 2 (Optional, defaults to Area 1)",
							Required:    false,
						},
					},
				},
				// Player Alias
				{
					Name:        "player-alias",
					Description: "Manage Player Aliases",
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Options: []*discordgo.ApplicationCommandOption{
						// 1. Link Discord User
						{
							Name:        "set",
							Description: "Link a Discord User to an IGN and Area",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionUser,
									Name:        "user",
									Description: "The Discord User",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "ign",
									Description: "In-Game Name",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "area",
									Description: "Area Code",
									Required:    true,
								},
							},
						},
						// 2. Link Custom Tag (For non-Discord players)
						{
							Name:        "set-custom",
							Description: "Create a custom alias tag (e.g. 'rival', 'alt')",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
							Options: []*discordgo.ApplicationCommandOption{
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "tag",
									Description: "Unique Tag Name (e.g. 'takumi', 'rival1')",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "ign",
									Description: "In-Game Name",
									Required:    true,
								},
								{
									Type:        discordgo.ApplicationCommandOptionString,
									Name:        "area",
									Description: "Area Code",
									Required:    true,
								},
							},
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
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User info (IGN or @User)",
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
				// Subcommand: Online Battle
				{
					Name:        "ob-ranking",
					Description: "Get Battle Online Rankings",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "rank",
							Description: "Filter results by rank",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Pride", Value: "pride"},
								{Name: "Ruby", Value: "ruby"},
								{Name: "Emerald", Value: "emerald"},
								{Name: "Red", Value: "red"},
								{Name: "Blue", Value: "blue"},
								{Name: "Green", Value: "green"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area",
							Description: "Filter by Country/Area (e.g. 'vn', 'jp')",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "round",
							Description: "Filter by round (number or \"all\"). If left empty, will use current round.",
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
				{
					Name:        "ob-meta",
					Description: "Get Online Battle meta cars",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "rank",
							Description: "Filter results by rank",
							Required:    false,
							Choices: []*discordgo.ApplicationCommandOptionChoice{
								{Name: "Pride", Value: "pride"},
								{Name: "Ruby", Value: "ruby"},
								{Name: "Emerald", Value: "emerald"},
								{Name: "Red", Value: "red"},
								{Name: "Blue", Value: "blue"},
								{Name: "Green", Value: "green"},
							},
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "area",
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
		// --- Simple Commands ---
		{Name: "status", Description: "Show current bot status"},
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
		{Name: "mckween", Description: "desuwa"},
		{Name: "miemebell", Description: "Mie me bell"},
		{
			Name:        "meomeo",
			Description: "For glazing an user for their exceptional IDAC skills.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "(Optional) User to tag",
					Required:    false,
				},

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "language",
					Description: "(Optional) Choose a language",
					Required:    false,
					Choices:     languageChoice,
				},
			},
		},
	}

	// 3. Register Event Router
	h.Session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		// A. Execute Commands
		case discordgo.InteractionApplicationCommand:
			name := i.ApplicationCommandData().Name
			switch name {
			case "idac":
				h.routeIdacCommand(i)
			case "status":
				h.HandleStatus(i)
			case "nuhuh":
				h.HandleNuhuh(i)
			case "mckween":
				h.HandleMckween(i)
			case "miemebell":
				h.HandleMiemebell(i)
			case "meomeo":
				h.HandleMeoMeo(i)
			}

		// B. Handle Autocomplete
		case discordgo.InteractionApplicationCommandAutocomplete:
			if i.ApplicationCommandData().Name == "idac" {
				h.HandleAutoComplete(i)
			}
		}
	})

	// 4. Register with Discord
	log.Println("Registering commands...")
	_, err = h.Session.ApplicationCommandBulkOverwrite(botID, "", commands)
	return err
}

// routeIdacCommand routes subcommands for /idac
func (h *Handler) routeIdacCommand(i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	if len(options) == 0 {
		return
	}
	rootOption := options[0]

	// --- 1. HANDLE SUBCOMMAND GROUPS ("player-alias") ---
	if rootOption.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		//Owner Check
		if !h.IsRequestFromOwner(h.Session, i) {
			return
		}
		subCmd := rootOption.Options[0]
		optMap := make(map[string]string)

		// Extract options for alias
		for _, opt := range subCmd.Options {
			if opt.Type == discordgo.ApplicationCommandOptionUser {
				optMap["user"] = opt.UserValue(h.Session).ID
			} else {
				optMap[opt.Name] = opt.StringValue()
			}
		}
		if rootOption.Name == "player-alias" {
			switch subCmd.Name {
			case "set":
				h.HandleSetPlayerAlias(i, optMap["user"], optMap)
			case "set-custom":
				h.HandleSetPlayerAlias(i, optMap["tag"], optMap)
			}
		}
		return
	}
	// --- 2. HANDLE REGULAR SUBCOMMANDS ---
	// (time-attack, team, player-info, player-compare)
	subcommand := rootOption
	optMap := make(map[string]string)
	specInput := ""

	for _, opt := range subcommand.Options {
		if opt.Type == discordgo.ApplicationCommandOptionUser {
			if opt.Name == "user" {
				optMap["user"] = opt.UserValue(h.Session).ID
			} else if opt.Name == "user1" {
				optMap["user1"] = opt.UserValue(h.Session).ID
			} else if opt.Name == "user2" {
				optMap["user2"] = opt.UserValue(h.Session).ID
			}
		} else if opt.Name == "spec" {
			specInput = opt.StringValue()
		} else if opt.Type == discordgo.ApplicationCommandOptionInteger {
			optMap[opt.Name] = strconv.FormatInt(opt.IntValue(), 10)
		} else {
			optMap[opt.Name] = opt.StringValue()
		}
	}

	switch subcommand.Name {
	case "time-attack":
		h.HandleTimeAttack(i, optMap, specInput)
	case "team":
		h.HandleTeamRanking(i, optMap)
	case "player-info":
		h.HandlePlayerInfo(i, optMap)
	case "player-compare":
		h.HandlePlayerCompare(i, optMap)
	case "ob-ranking":
		h.HandleOBRanking(i, optMap)
	case "ob-meta":
		h.HandleOBMeta(i, optMap)
	}
}
