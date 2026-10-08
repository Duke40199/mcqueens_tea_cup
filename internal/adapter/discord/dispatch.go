package discord

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/pkg/logger"
	"McQueens_Tea_Cup/pkg/tracer"
)

// commandTimeout bounds how long a single slash-command handler may run before its
// context is cancelled, so a slow upstream (Sega/AllNet/DB) can't keep the handler
// goroutine and its connections alive indefinitely. It stays well under Discord's
// 15-minute deferred-response window.
const commandTimeout = 30 * time.Second

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
	// Start a trace for this interaction and tag it with the invoking user so all
	// downstream logs carry the same trace_id / user_id.
	base := tracer.NewContext(context.Background())
	if u := interactionUser(i); u != nil {
		base = tracer.WithUserID(base, u.ID)
	}

	if err := h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		logger.Error(base, "failed to defer interaction", err)
		return
	}

	ctx, cancel := context.WithTimeout(base, commandTimeout)
	defer cancel()

	cc := &CommandContext{
		Ctx:         ctx,
		Session:     h.Session,
		Interaction: i,
		OptMap:      optMap,
		SpecInput:   specInput,
		h:           h,
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "recovered from panic in command handler", fmt.Errorf("%v", r), slog.String("panic_stack", string(debug.Stack())))
			cc.replyError(NewUserError("Something went wrong while processing that command. Please try again later."))
		}
	}()

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
	cc.h.SendPagination(cc.Ctx, cc.Interaction, pages)
}

// SendPagesWithThumbnail paginates like SendPages but renders each page as an embed
// carrying the given thumbnail image.
func (cc *CommandContext) SendPagesWithThumbnail(pages []string, thumbnailURL string) {
	cc.h.SendPaginationWithThumbnail(cc.Ctx, cc.Interaction, pages, thumbnailURL)
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
		logger.Error(cc.Ctx, "command error", err)
	}
	if _, e := cc.Session.InteractionResponseEdit(cc.Interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	}); e != nil {
		logger.Error(cc.Ctx, "failed to send error reply", e)
	}
}
