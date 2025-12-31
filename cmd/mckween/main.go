package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"McQueens_Tea_Cup/internal/config"
	discord_handler "McQueens_Tea_Cup/internal/delivery/discord"
	"McQueens_Tea_Cup/internal/infra"
	db "McQueens_Tea_Cup/internal/infra/db"
	discord_infra "McQueens_Tea_Cup/internal/infra/discord"
	"McQueens_Tea_Cup/internal/infra/sega"
	"McQueens_Tea_Cup/internal/usecase"

	"github.com/bwmarrin/discordgo"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error:", err)
	}
	// --- NEW: Connect to Postgres ---
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.DatabaseCfg.Host, cfg.DatabaseCfg.Port, cfg.DatabaseCfg.User, cfg.DatabaseCfg.Password, cfg.DatabaseCfg.Name)
	dbConn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	aliasRepo := db.NewPostgresAliasRepo(dbConn)
	// 2. SHARED INFRASTRUCTURE: Create Discord Session ONCE
	// We do not Open() it yet. We just create the struct.
	dg, err := discordgo.New("Bot " + cfg.DiscordCfg.Token)
	if err != nil {
		log.Fatal("Discord creation error:", err)
	}

	// ---------------------------------------------------------
	// FEATURE A: IDAC COMMANDS
	// ---------------------------------------------------------

	// A1. Init Repositories (Data Access)
	segaClient := sega.NewClient() // Handles HTTP to SEGA

	// A2. Init Delivery (The Command Controller)
	// We inject the shared 'dg' session here
	cmdHandler := discord_handler.NewHandler(dg, segaClient, aliasRepo)

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
	rssNotifier := discord_infra.NewDiscordNotifier(dg, "")

	// B3. Init Use Case
	feedLogic := usecase.NewFeedChecker(rssFetcher, rssNotifier, stateStore, translator)

	// ---------------------------------------------------------
	// STARTUP & RUN LOOP
	// ---------------------------------------------------------

	// 3. Open Connection
	// This starts the WebSocket listener for Commands AND enables sending for RSS
	if err = dg.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	log.Println("✅ Bot is running. Press CTRL-C to exit.")

	// 4. Run RSS Ticker in a Goroutine (Background)
	ticker := time.NewTicker(time.Duration(cfg.RSSCfg.Interval) * time.Minute)
	defer ticker.Stop()

	// Run immediately once on startup
	go feedLogic.Check(cfg.RSSCfg.Feeds)

	// Handle OS Signals (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Loop
	go func() {
		for range ticker.C {
			feedLogic.Check(cfg.RSSCfg.Feeds)
		}
	}()

	// Block Main Thread until signal received
	<-stop
	log.Println("Gracefully shutting down...")
}
