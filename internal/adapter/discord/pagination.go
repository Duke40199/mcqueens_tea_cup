package discord

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"McQueens_Tea_Cup/pkg/logger"
	"McQueens_Tea_Cup/pkg/tracer"
)

// paginationTTL is how long a paginated message stays interactive before its
// buttons are removed and its state is dropped.
const paginationTTL = 2 * time.Minute

// Component CustomIDs for the pagination buttons.
const (
	paginationPrevID   = "pagination_prev"
	paginationNextID   = "pagination_next"
	paginationStatusID = "pagination_status"
	paginationPrefix   = "pagination_"
)

// paginationSession is the server-side state for one paginated message. When
// thumbnailURL is set, pages render as an embed (description = page text) with that
// thumbnail; otherwise they render as plain message content.
type paginationSession struct {
	pages        []string
	index        int
	ownerID      string
	thumbnailURL string
	timer        *time.Timer
}

// paginationResult is the outcome of a button click lookup.
type paginationResult int

const (
	paginationOK paginationResult = iota
	paginationNotFound
	paginationNotOwner
)

// paginationStore keeps per-message pagination state, keyed by message ID, and is
// safe for concurrent access. This replaces the previous approach of registering a
// new global AddHandler closure (with a captured, unsynchronized pageIndex) per
// invocation, which both raced on the index and fanned every component event out
// to every live pagination.
type paginationStore struct {
	mu       sync.Mutex
	sessions map[string]*paginationSession
	ttl      time.Duration
}

func newPaginationStore() *paginationStore {
	return &paginationStore{
		sessions: make(map[string]*paginationSession),
		ttl:      paginationTTL,
	}
}

// register stores a new session for messageID. onExpire runs once, after the TTL,
// after the session has been removed (used to clear the buttons).
func (ps *paginationStore) register(messageID, ownerID string, pages []string, thumbnailURL string, onExpire func()) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if existing, ok := ps.sessions[messageID]; ok && existing.timer != nil {
		existing.timer.Stop()
	}

	sess := &paginationSession{pages: pages, ownerID: ownerID, thumbnailURL: thumbnailURL}
	sess.timer = time.AfterFunc(ps.ttl, func() {
		ps.remove(messageID)
		if onExpire != nil {
			onExpire()
		}
	})
	ps.sessions[messageID] = sess
}

func (ps *paginationStore) remove(messageID string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if sess, ok := ps.sessions[messageID]; ok {
		if sess.timer != nil {
			sess.timer.Stop()
		}
		delete(ps.sessions, messageID)
	}
}

// advance applies a button click (prev/next) to the session for messageID and
// returns the content + components to render. The whole read-modify-read happens
// under the lock, so concurrent clicks can't race on the index.
func (ps *paginationStore) advance(messageID string, user *discordgo.User, customID string) (text, thumbnailURL string, components []discordgo.MessageComponent, result paginationResult) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	sess, ok := ps.sessions[messageID]
	if !ok {
		return "", "", nil, paginationNotFound
	}
	if user == nil || user.ID != sess.ownerID {
		return "", "", nil, paginationNotOwner
	}

	switch customID {
	case paginationPrevID:
		if sess.index > 0 {
			sess.index--
		}
	case paginationNextID:
		if sess.index < len(sess.pages)-1 {
			sess.index++
		}
	}
	return sess.pages[sess.index], sess.thumbnailURL, paginationComponents(sess.index, len(sess.pages)), paginationOK
}

// pageWebhookEdit builds the WebhookEdit for a page: an embed (with thumbnail) when
// thumbnailURL is set, otherwise plain content. components may be nil to leave them
// unchanged.
func pageWebhookEdit(text, thumbnailURL string, components *[]discordgo.MessageComponent) *discordgo.WebhookEdit {
	edit := &discordgo.WebhookEdit{Components: components}
	if thumbnailURL != "" {
		edit.Embeds = &[]*discordgo.MessageEmbed{{
			Description: text,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: thumbnailURL},
		}}
	} else {
		edit.Content = &text
	}
	return edit
}

// pageResponseData builds the UpdateMessage payload for a page (embed when a
// thumbnail is set, otherwise plain content).
func pageResponseData(text, thumbnailURL string, components []discordgo.MessageComponent) *discordgo.InteractionResponseData {
	data := &discordgo.InteractionResponseData{Components: components}
	if thumbnailURL != "" {
		data.Embeds = []*discordgo.MessageEmbed{{
			Description: text,
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: thumbnailURL},
		}}
	} else {
		data.Content = text
	}
	return data
}

// paginationComponents builds the Prev / "Page x/y" / Next button row for the
// given page position.
func paginationComponents(current, total int) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "◀️ Previous",
					Style:    discordgo.PrimaryButton,
					CustomID: paginationPrevID,
					Disabled: current == 0,
				},
				discordgo.Button{
					Label:    fmt.Sprintf("Page %d/%d", current+1, total),
					Style:    discordgo.SecondaryButton,
					CustomID: paginationStatusID,
					Disabled: true, // Just a label
				},
				discordgo.Button{
					Label:    "Next ▶️",
					Style:    discordgo.PrimaryButton,
					CustomID: paginationNextID,
					Disabled: current == total-1,
				},
			},
		},
	}
}

// SendPagination edits the deferred response with the first page as plain message
// content and, when there is more than one page, wires up navigation buttons.
func (h *Handler) SendPagination(ctx context.Context, i *discordgo.InteractionCreate, pages []string) {
	h.sendPagination(ctx, i, pages, "")
}

// SendPaginationWithThumbnail is like SendPagination but renders each page as an
// embed carrying the given thumbnail image (top-right).
func (h *Handler) SendPaginationWithThumbnail(ctx context.Context, i *discordgo.InteractionCreate, pages []string, thumbnailURL string) {
	h.sendPagination(ctx, i, pages, thumbnailURL)
}

func (h *Handler) sendPagination(ctx context.Context, i *discordgo.InteractionCreate, pages []string, thumbnailURL string) {
	if len(pages) == 0 {
		return
	}
	// Single page: no buttons needed.
	if len(pages) == 1 {
		if _, err := h.Session.InteractionResponseEdit(i.Interaction, pageWebhookEdit(pages[0], thumbnailURL, nil)); err != nil {
			logger.Error(ctx, "failed to send single page", err)
		}
		return
	}

	components := paginationComponents(0, len(pages))
	msg, err := h.Session.InteractionResponseEdit(i.Interaction, pageWebhookEdit(pages[0], thumbnailURL, &components))
	if err != nil {
		logger.Error(ctx, "failed to send first page", err)
		return
	}

	ownerID := ""
	if u := interactionUser(i); u != nil {
		ownerID = u.ID
	}

	// Capture the interaction so the expiry callback can clear the buttons.
	interaction := i.Interaction
	h.pagination.register(msg.ID, ownerID, pages, thumbnailURL, func() {
		empty := []discordgo.MessageComponent{}
		if _, err := h.Session.InteractionResponseEdit(interaction, &discordgo.WebhookEdit{
			Components: &empty,
		}); err != nil {
			logger.Error(ctx, "failed to clear buttons on expiry", err)
		}
	})
}

// handlePaginationComponent handles button clicks for all paginated messages. It
// is registered once (from the router), not per invocation.
func (h *Handler) handlePaginationComponent(s *discordgo.Session, ic *discordgo.InteractionCreate) {
	defer recoverInteraction("pagination handler")

	ctx := tracer.NewContext(context.Background())
	if u := interactionUser(ic); u != nil {
		ctx = tracer.WithUserID(ctx, u.ID)
	}

	if ic.Type != discordgo.InteractionMessageComponent {
		return
	}
	customID := ic.MessageComponentData().CustomID
	// Only react to our pagination buttons; the status button is a disabled label.
	if customID != paginationPrevID && customID != paginationNextID {
		return
	}
	if ic.Message == nil {
		return
	}

	content, thumbnailURL, components, res := h.pagination.advance(ic.Message.ID, interactionUser(ic), customID)
	switch res {
	case paginationOK:
		if err := s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: pageResponseData(content, thumbnailURL, components),
		}); err != nil {
			logger.Error(ctx, "failed to update message", err)
		}
	case paginationNotOwner:
		h.respondEphemeral(ctx, s, ic, "Only the person who ran the command can use these buttons.")
	case paginationNotFound:
		h.respondEphemeral(ctx, s, ic, "This menu has expired. Please run the command again.")
	}
}

// respondEphemeral sends a private, self-dismissing reply to a component click so
// the user doesn't see Discord's generic "interaction failed".
func (h *Handler) respondEphemeral(ctx context.Context, s *discordgo.Session, ic *discordgo.InteractionCreate, content string) {
	if err := s.InteractionRespond(ic.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		logger.Error(ctx, "failed to send ephemeral reply", err)
	}
}
