package discord

import (
	"log"

	"McQueens_Tea_Cup/internal/config"

	"github.com/bwmarrin/discordgo"
)

type DiscordSession struct {
	Session *discordgo.Session
}

func NewDiscordSession(cfg *config.DiscordConfig) (s *DiscordSession, err error) {
	discordConn, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("Discord creation error:", err)
		return nil, err
	}
	return &DiscordSession{Session: discordConn}, nil
}

func (ds *DiscordSession) OpenSession() error {
	return ds.Session.Open()
}

func (ds *DiscordSession) CloseSession() error {
	return ds.Session.Close()
}
