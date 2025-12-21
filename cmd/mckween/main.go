package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"McQueens_Tea_Cup/internal/delivery"
	"McQueens_Tea_Cup/internal/infra"
	"McQueens_Tea_Cup/internal/infra/config"
	persistence "McQueens_Tea_Cup/internal/infra/presistence"
	"McQueens_Tea_Cup/internal/infra/sega"
	"McQueens_Tea_Cup/internal/usecase"

	"github.com/bwmarrin/discordgo"
)

func main() {
	// 1. Config
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Config error:", err)
	}

	// 2. SHARED INFRASTRUCTURE: Create Discord Session ONCE
	// We do not Open() it yet. We just create the struct.
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		log.Fatal("Discord creation error:", err)
	}

	// ---------------------------------------------------------
	// FEATURE A: IDAC COMMANDS
	// ---------------------------------------------------------

	// A1. Init Repositories (Data Access)
	segaClient := sega.NewClient()                              // Handles HTTP to SEGA
	aliasStore := persistence.NewJSONAliasStore("aliases.json") // Handles JSON file

	// A2. Init Delivery (The Command Controller)
	// We inject the shared 'dg' session here
	cmdHandler := delivery.NewHandler(dg, segaClient, aliasStore)

	// A3. Register Commands & Event Handlers
	if err := cmdHandler.RegisterCommands(); err != nil {
		log.Printf("⚠️ Failed to register commands: %v", err)
	}

	// ---------------------------------------------------------
	// FEATURE B: RSS FEED CHECKER (Existing Logic)
	// ---------------------------------------------------------

	// B1. Init Adapters
	rssFetcher := infra.NewGoFeedFetcher()
	stateStore := infra.NewFileStore("feed_state.json")
	translator := infra.NewGoogleTranslator()

	// B2. Init Notifier (Updated)
	// NOTE: You must update NewDiscordNotifier to accept the existing 'dg' session
	// instead of creating a new one internally.
	rssNotifier := infra.NewDiscordNotifier(dg, cfg.ChannelID)

	// B3. Init Use Case
	feedLogic := usecase.NewFeedChecker(rssFetcher, rssNotifier, stateStore, translator)

	// ---------------------------------------------------------
	// STARTUP & RUN LOOP
	// ---------------------------------------------------------

	// 3. Open Connection
	// This starts the WebSocket listener for Commands AND enables sending for RSS
	if err := dg.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	log.Println("✅ Bot is running. Press CTRL-C to exit.")

	// 4. Run RSS Ticker in a Goroutine (Background)
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately once on startup
	go feedLogic.Check(cfg.Feeds)

	// Handle OS Signals (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Loop
	go func() {
		for range ticker.C {
			feedLogic.Check(cfg.Feeds)
		}
	}()

	// Block Main Thread until signal received
	<-stop
	log.Println("Gracefully shutting down...")
}
