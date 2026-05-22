package discord

import (
	"bytes"
	"context"
	_ "embed"
	"log"
	"strings"

	"McQueens_Tea_Cup/internal/adapter/database"
	"McQueens_Tea_Cup/internal/domain/entity"
	"McQueens_Tea_Cup/internal/domain/port"
	"McQueens_Tea_Cup/internal/domain/service"

	"github.com/bwmarrin/discordgo"
)

//go:embed resource/unauthorized.gif
var unauthorizedGif []byte

type Handler struct {
	OwnerID string
	Session *discordgo.Session
	// db repositories
	AliasRepo          database.AliasRepository
	OBRankingCfgRepo   database.OBRankingCfgRepository
	RankingCfgRepo     database.RankingCfgRepository
	CarRepo            database.CarRepository
	TATimeMetadataRepo database.TATimeMetadataRepository
	CfsStateRepo       database.CfsStateRepository
	// internal services
	MetaLogic            *service.MetaLogicService
	StoreLocationService port.StoreLocationService
	// clients
	SegaClient   port.SegaIDACClient
	AllNetClient port.AllNetClient
	// local memory
	CarChoices   []*entity.CarMetadata
	TrackChoices map[string]string
}

// NewHandler creates our controller
func NewHandler(
	s *discordgo.Session,
	// db repositories
	aliasRepo database.AliasRepository,
	obRankingCfgRepo database.OBRankingCfgRepository,
	rankingCfgRepo database.RankingCfgRepository,
	carRepo database.CarRepository,
	taTimeMetadataRepo database.TATimeMetadataRepository,
	cfsStateRepo database.CfsStateRepository,
	// internal services
	metaLogic *service.MetaLogicService,
	storeLocationService port.StoreLocationService,
	// clients
	segaClient port.SegaIDACClient,
	allNetClient port.AllNetClient,

) *Handler {
	discordHandler := &Handler{
		Session: s,
		// db repositories
		AliasRepo:          aliasRepo,
		OBRankingCfgRepo:   obRankingCfgRepo,
		RankingCfgRepo:     rankingCfgRepo,
		CarRepo:            carRepo,
		TATimeMetadataRepo: taTimeMetadataRepo,
		CfsStateRepo:       cfsStateRepo,
		// internal services
		MetaLogic:            metaLogic,
		StoreLocationService: storeLocationService,
		// clients
		SegaClient:   segaClient,
		AllNetClient: allNetClient,
		OwnerID:      "384015507302383616",
	}
	discordHandler.CarChoices = discordHandler.PreloadListCarSelection()
	discordHandler.TrackChoices = entity.MergedTrackRegistry
	return discordHandler
}

func (h *Handler) PreloadListCarSelection() []*entity.CarMetadata {
	var carChoices []*entity.CarMetadata
	listCarWithSpecIDs, err := h.CarRepo.GetListCarWithAggregatedSpecs(context.Background())
	if err != nil {
		log.Println("error getting list car selection:", err)
		return carChoices
	}
	for _, result := range listCarWithSpecIDs {
		result.SearchBlob = strings.ToLower(result.Maker + result.Name + result.ModelCode)
	}
	return listCarWithSpecIDs
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
