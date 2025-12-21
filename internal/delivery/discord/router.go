package discord

import (
	idac_domain "McQueens_Tea_Cup/internal/domain/idac"
	_ "embed"
	"log"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

//go:embed resource/nuhuh.gif
var nuhuhGif []byte

//go:embed resource/desuwa.gif
var desuwaGif []byte

//go:embed resource/miemebell.json
var miemebellJson []byte

func (h *Handler) RegisterCommands() error {
	// 1. Define Constants & Choices
	// minLimit := 1.0
	// maxLimit := 1000.0

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
			Description: "Get Initial D Arcade Rankings",
			Options:     []*discordgo.ApplicationCommandOption{
				// [Insert your full IDAC options list here: time-attack, team, player-info, player-compare, player-alias]
				// (Omitted for brevity, paste your existing IDAC options struct here)
			},
		},
		// --- Simple Commands ---
		{Name: "status", Description: "Show current bot status"},
		{
			Name:        "nuhuh",
			Description: "Reply with the nuh uh gif",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User to tag", Required: false},
			},
		},
		{Name: "mckween", Description: "desuwa"},
		{Name: "miemebell", Description: "Get a random Mieme quote"},
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
	_, err := h.Session.ApplicationCommandBulkOverwrite(h.Session.State.User.ID, "", commands)
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
		// Owner Check
		// if !h.checkOwnerPermission(i) {
		// 	return
		// }

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
			if subCmd.Name == "set" {
				// Assumes you have moved handlePlayerAliasSet to be a method on Handler
				h.HandlePlayerAliasSet(i, optMap["user"], optMap)
			} else if subCmd.Name == "set-custom" {
				h.HandlePlayerAliasSet(i, optMap["tag"], optMap)
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
		// case "player-info":
		// 	h.HandlePlayerInfo(i, optMap)
		// case "player-compare":
		// 	h.HandlePlayerCompare(i, optMap)
	}
}

func (h *Handler) HandleStatus(i *discordgo.InteractionCreate) {
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "I am running and monitoring your feeds! 📡",
		},
	})
}
