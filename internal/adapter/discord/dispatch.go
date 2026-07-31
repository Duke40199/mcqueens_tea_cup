package discord

import (
	"context"
	"errors"
	"log"

	"github.com/bwmarrin/discordgo"
)

// CommandContext carries everything a slash-command handler needs for a single
// interaction and owns the reply channel back to Discord. By the time a handler
// receives it, dispatch has already sent the deferred acknowledgement, so
// handlers only produce final content (via Edit / SendPages) or return an error.
type CommandContext struct {
	Ctx         context.Context
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	OptMap      map[string]string
	SpecInput   string

	h *Handler
}

// UserError wraps a message that is safe to show to the user verbatim. Any other
// error returned by a handler is logged and replaced with a generic message so
// we never leak internals into a channel.
type UserError struct{ Msg string }

func (e *UserError) Error() string { return e.Msg }

// NewUserError builds a UserError with the given user-facing message.
func NewUserError(msg string) *UserError { return &UserError{Msg: msg} }

// CommandFunc is a lifecycle-managed slash-command handler. It never defers or
// hand-rolls error replies itself — dispatch owns that.
type CommandFunc func(*CommandContext) error

// dispatch owns the interaction lifecycle: it sends the deferred ack inside
// Discord's 3-second window, runs the handler, and turns a returned error into a
// single user-facing edit of the deferred response.
func (h *Handler) dispatch(i *discordgo.InteractionCreate, optMap map[string]string, specInput string, fn CommandFunc) {
	if err := h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("discord: failed to defer interaction: %v", err)
		return
	}

	cc := &CommandContext{
		Ctx:         context.Background(),
		Session:     h.Session,
		Interaction: i,
		OptMap:      optMap,
		SpecInput:   specInput,
		h:           h,
	}

	if err := fn(cc); err != nil {
		cc.replyError(err)
	}
}

// Edit replaces the deferred response with plain text content.
func (cc *CommandContext) Edit(content string) error {
	_, err := cc.Session.InteractionResponseEdit(cc.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &content,
	})
	return err
}

// SendPages hands the rendered pages to the pagination helper, which edits the
// deferred response and wires up navigation buttons.
func (cc *CommandContext) SendPages(pages []string) {
	cc.h.SendPagination(cc.Interaction, pages)
}

// replyError edits the deferred response with a user-facing error message.
// UserError messages are shown as-is; anything else is logged and replaced with
// a generic message.
func (cc *CommandContext) replyError(err error) {
	var ue *UserError
	msg := "⚠️ Something went wrong while processing that command. Please try again later."
	if errors.As(err, &ue) {
		msg = "⚠️ " + ue.Msg
	} else {
		log.Printf("discord: command error: %v", err)
	}
	if _, e := cc.Session.InteractionResponseEdit(cc.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	}); e != nil {
		log.Printf("discord: failed to send error reply: %v", e)
	}
}
