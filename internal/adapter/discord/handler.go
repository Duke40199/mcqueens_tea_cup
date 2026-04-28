package discord

import (
	"bytes"
	_ "embed"

	"McQueens_Tea_Cup/internal/adapter/postgres"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/usecase"

	"github.com/bwmarrin/discordgo"
)

//go:embed resource/unauthorized.gif
var unauthorizedGif []byte

type Handler struct {
	OwnerID string
	Session *discordgo.Session
	// db repositories
	AliasRepo          postgres.AliasRepository
	OBRankingCfgRepo   postgres.OBRankingCfgRepository
	RankingCfgRepo     postgres.RankingCfgRepository
	CarRepo            postgres.CarRepository
	TATimeMetadataRepo postgres.TATimeMetadataRepository
	CfsStateRepo       postgres.CfsStateRepository
	// internal services
	MetaLogic *usecase.MetaLogicService
	// clients
	SegaClient entity.SegaClient
}

// NewHandler creates our controller
func NewHandler(
	s *discordgo.Session,
	// db repositories
	aliasRepo postgres.AliasRepository,
	obRankingCfgRepo postgres.OBRankingCfgRepository,
	rankingCfgRepo postgres.RankingCfgRepository,
	carRepo postgres.CarRepository,
	taTimeMetadataRepo postgres.TATimeMetadataRepository,
	cfsStateRepo postgres.CfsStateRepository,
	// internal services
	metaLogic *usecase.MetaLogicService,
	// clients
	segaClient entity.SegaClient,
) *Handler {
	return &Handler{
		Session: s,
		// db repositories
		AliasRepo:          aliasRepo,
		OBRankingCfgRepo:   obRankingCfgRepo,
		RankingCfgRepo:     rankingCfgRepo,
		CarRepo:            carRepo,
		TATimeMetadataRepo: taTimeMetadataRepo,
		CfsStateRepo:       cfsStateRepo,
		// internal services
		MetaLogic: metaLogic,
		// clients
		SegaClient: segaClient,
		OwnerID:    "384015507302383616",
	}
}

func (h *Handler) HandleStatus(i *discordgo.InteractionCreate) {
	h.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Up and running, desuwa~",
		},
	})
}

// IsRequestFromOwner checks if the user is the owner.
// If NOT, it sends the unauthorized GIF/Message and returns false.
// If YES, it returns true.
func (h *Handler) IsRequestFromOwner(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	// Replace with your actual Owner ID
	const OwnerID = "384015507302383616"

	// Check ID
	if i.Member.User.ID == OwnerID {
		return true
	}
	// 1. Respond immediately with just the GIF
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Are you the tea brewer? Let's check that real quick.",
			Files: []*discordgo.File{
				{Name: "unauthorized.gif", Reader: bytes.NewReader(unauthorizedGif)},
			},
		},
	})

	// 2. Send the text as a Follow-up (appears underneath)
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Guess not. You might have used a command which is only for the bot's creator." +
			" If you need help, feel free to contact @spacepotato4109.",
	})
	return false
}

func (h *Handler) SendDeferredError(i *discordgo.InteractionCreate, msg string) {
	h.Session.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
}
