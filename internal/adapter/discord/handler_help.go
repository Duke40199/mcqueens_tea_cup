package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandleHelp renders a command guide read live from the DB as an embed. With no
// "command" option it lists the available guides; with one it shows that guide's
// content.
func (h *Handler) HandleHelp(cc *CommandContext) error {
	key := strings.TrimSpace(cc.OptMap["command"])
	if key == "" {
		return h.handleHelpList(cc)
	}

	guide, err := h.CommandGuideRepo.GetByCommandKey(cc.Ctx, key)
	if err != nil {
		return fmt.Errorf("fetching command guide %q: %w", key, err)
	}
	if guide == nil {
		return cc.Edit(fmt.Sprintf("⚠️ No help guide found for **%s** yet.", key))
	}

	return cc.editEmbed(&discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📖 %s", guide.Title),
		Description: guide.Content,
	})
}

// handleHelpList shows the directory of available guides, read live so newly added
// rows appear without a restart.
func (h *Handler) handleHelpList(cc *CommandContext) error {
	guides, err := h.CommandGuideRepo.ListEnabled(cc.Ctx)
	if err != nil {
		return fmt.Errorf("listing command guides: %w", err)
	}
	if len(guides) == 0 {
		return cc.Edit("ℹ️ No command guides are available yet.")
	}

	var body strings.Builder
	body.WriteString("Use `/idac help command:<name>` to read a specific guide.\n\n")
	for _, g := range guides {
		body.WriteString(fmt.Sprintf("• **%s**\n", g.Title))
	}
	return cc.editEmbed(&discordgo.MessageEmbed{
		Title:       "📖 Command Guides",
		Description: body.String(),
	})
}

// editEmbed replaces the deferred response with a single embed, mirroring the
// embed-based reply style used by HandlePlayerInfo.
func (cc *CommandContext) editEmbed(embed *discordgo.MessageEmbed) error {
	_, err := cc.Session.InteractionResponseEdit(cc.Interaction.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	})
	return err
}

// SearchCommandGuideChoiceQuery powers the "command" autocomplete for /idac help,
// matching the typed query against guide titles and keys from the in-memory cache.
func (h *Handler) SearchCommandGuideChoiceQuery(query string) []*discordgo.ApplicationCommandOptionChoice {
	query = strings.ToLower(strings.TrimSpace(query))
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, g := range h.CommandGuides {
		if query != "" &&
			!strings.Contains(strings.ToLower(g.Title), query) &&
			!strings.Contains(strings.ToLower(g.CommandKey), query) {
			continue
		}
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  g.Title,
			Value: g.CommandKey,
		})
		if len(choices) >= maxAutocompleteChoices {
			break
		}
	}
	return choices
}
