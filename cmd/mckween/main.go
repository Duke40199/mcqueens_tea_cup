package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	discord_handler "McQueens_Tea_Cup/internal/adapter/discord"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/infra"
	"McQueens_Tea_Cup/internal/infra/db"
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
	obRankingCfgRepo := db.NewOBRankingCfgRepository(dbConn)
	areaRepo := db.NewAreaRepository(dbConn)

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
	carRepo := db.NewCarRepository(dbConn)

	// A2. Init Logic Services
	metaLogic := usecase.NewMetaLogicService(segaClient, carRepo)

	// A3. Init Delivery (The Command Controller)
	// We inject the shared 'dg' session here
	cmdHandler := discord_handler.NewHandler(dg, segaClient, aliasRepo, obRankingCfgRepo, carRepo, metaLogic)

	// A4. Register Commands & Event Handlers
	if err := cmdHandler.RegisterCommands(); err != nil {
		log.Printf("⚠️ Failed to register commands: %v", err)
	}

	// ---------------------------------------------------------
	// FEATURE C: CAR DATA SYNC (Cron)
	// ---------------------------------------------------------
	carSyncService := usecase.NewCarSyncService(segaClient, carRepo)

	// Run Sync in background (every 24h)
	go func() {
		// Run once on startup
		log.Println("⏳ Initializing Car Data Sync...")
		if err := carSyncService.SyncData(context.Background()); err != nil {
			log.Printf("❌ Initial Car Sync Failed: %v", err)
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			log.Println("⏰ Starting Scheduled Car Sync...")
			if err := carSyncService.SyncData(context.Background()); err != nil {
				log.Printf("❌ Scheduled Car Sync Failed: %v", err)
			}
		}
	}()

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
	rssTicker := time.NewTicker(time.Duration(cfg.RSSCfg.Interval) * time.Minute)
	defer rssTicker.Stop()

	// Run immediately once on startup
	go feedLogic.Check(cfg.RSSCfg.Feeds)

	// Handle OS Signals (Graceful Shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// ---------------------------------------------------------
	// FEATURE D: OBMETA SYNC (Scheduled)
	// ---------------------------------------------------------
	metaSync := usecase.NewMetaSyncService(dg, metaLogic, cfg.MetaSyncCfg)

	// Run Sync in background
	go func() {
		// Interval from config
		interval := time.Duration(cfg.MetaSyncCfg.Interval) * time.Minute
		metaTicker := time.NewTicker(interval)
		defer metaTicker.Stop()

		// Run once on startup (after Discord is open)
		time.Sleep(5 * time.Second) // Small delay to ensure bot is fully ready
		log.Println("📊 Initializing OBMeta Sync...")
		if err := metaSync.Sync(context.Background()); err != nil {
			log.Printf("❌ Initial OBMeta Sync Failed: %v", err)
		}

		for range metaTicker.C {
			if err := metaSync.Sync(context.Background()); err != nil {
				log.Printf("❌ Scheduled OBMeta Sync Failed: %v", err)
			}
		}
	}()

	// ---------------------------------------------------------
	// FEATURE E: ACTIVE PLAYERS SYNC (Scheduled)
	// ---------------------------------------------------------
	activePlayersSync := usecase.NewActivePlayerSyncService(dg, segaClient, areaRepo, obRankingCfgRepo, cfg.ActivePlayersSyncCfg)

	// Run Sync in background
	go func() {
		interval := time.Duration(cfg.ActivePlayersSyncCfg.Interval) * time.Minute
		activeTicker := time.NewTicker(interval)
		defer activeTicker.Stop()

		// Initial run on startup (with delay)
		time.Sleep(10 * time.Second)
		log.Println("🔍 Initializing Active Players SEA Sync...")
		if err := activePlayersSync.Sync(context.Background()); err != nil {
			log.Printf("❌ Initial Active Players Sync Failed: %v", err)
		}

		for range activeTicker.C {
			if err := activePlayersSync.Sync(context.Background()); err != nil {
				log.Printf("❌ Scheduled Active Players Sync Failed: %v", err)
			}
		}
	}()

	// Loop
	go func() {
		for range rssTicker.C {
			feedLogic.Check(cfg.RSSCfg.Feeds)
		}
	}()

	// Block Main Thread until signal received
	<-stop
	log.Println("Gracefully shutting down...")
}
