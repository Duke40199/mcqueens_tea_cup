package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"McQueens_Tea_Cup/internal/adapter/database"
	discord_handler "McQueens_Tea_Cup/internal/adapter/discord"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/infra/db"
	"McQueens_Tea_Cup/internal/infra/sega"
	"McQueens_Tea_Cup/internal/usecase"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	// 2. Connect to DB
	dbConn, err := database.NewPostgresDBConn(cfg.DatabaseCfg)
	if err != nil {
		log.Fatal("Postgres database connection error: ", err)
	}

	// 3. Init Repositories
	aliasRepo := db.NewAliasRepo(dbConn)
	obRankingCfgRepo := db.NewOBRankingCfgRepository(dbConn)
	areaRepo := db.NewAreaRepository(dbConn)
	taTimeMetadataRepo := db.NewTATimeMetadataRepository(dbConn)
	rankingCfgRepo := db.NewRankingCfgRepo(dbConn)
	cfsStateRepo := db.NewCfsStateRepository(dbConn)
	carRepo := db.NewCarRepository(dbConn)

	// 4. Init Discord Session
	discordSession, err := discord_handler.NewDiscordSession(&cfg.DiscordCfg)
	if err != nil {
		log.Fatal("Discord session creation error:", err)
	}
	if err = discordSession.OpenSession(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer discordSession.CloseSession()

	// 5. Init clients
	segaClient := sega.NewSegaIDACClient(&cfg.SegaClientCfg)
	metaLogic := usecase.NewMetaLogicService(segaClient, carRepo)
	cmdHandler := discord_handler.NewHandler(
		discordSession.Session,
		aliasRepo,
		obRankingCfgRepo,
		rankingCfgRepo,
		carRepo,
		taTimeMetadataRepo,
		cfsStateRepo,
		metaLogic,
		segaClient,
	)
	// 6. Register Commands & Event Handlers
	if err := cmdHandler.RegisterCommands(); err != nil {
		log.Printf("⚠️ Failed to register commands: %v", err)
	}

	// 7.a. Cron Job: Online Battle Car Meta
	metaSync := usecase.NewMetaSyncService(discordSession.Session, metaLogic, cfg.MetaSyncCfg)
	go func() {
		ctx := context.Background()
		for {
			log.Println("📊 Starting OBMeta Sync...")
			metaLogic.SleepUntilNextSync(ctx, cfg.MetaSyncCfg.DowntimeStart, cfg.MetaSyncCfg.DowntimeEnd, cfg.MetaSyncCfg.DowntimeTZ)
			detectTime, err := metaSync.Sync(ctx)
			if err != nil {
				log.Printf("❌ Scheduled OBMeta Sync Failed: %v", err)
			} else {
				log.Printf("✅ OBMeta Sync detected data at: %s", detectTime)
			}
		}
	}()

	// 7.b. Cron Job: Online Battle Active Players
	activePlayersSync := usecase.NewActivePlayerSyncService(discordSession.Session, segaClient, areaRepo, obRankingCfgRepo, metaLogic, cfg.ActivePlayersSyncCfg)
	go func() {
		ctx := context.Background()
		for {
			log.Println("🔍 Starting Active Players SEA Sync...")
			metaLogic.SleepUntilNextSync(ctx, cfg.ActivePlayersSyncCfg.DowntimeStart, cfg.ActivePlayersSyncCfg.DowntimeEnd, cfg.ActivePlayersSyncCfg.DowntimeTZ)
			detectTime, err := activePlayersSync.Sync(ctx)
			if err != nil {
				log.Printf("❌ Scheduled Active Players Sync Failed: %v", err)
			} else {
				log.Printf("✅ Active Players Sync detected data at: %s", detectTime)
			}
		}
	}()

	// 8. Init graceful shutdown
	log.Println("✅ Bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Gracefully shutting down...")
}

// ---------------------------------------------------------
// FEATURE C: CAR DATA SYNC (Cron)
// ---------------------------------------------------------
// TODO: check for both model_code & aliases
// carSyncService := usecase.NewCarSyncService(segaClient, carRepo)

// Run Sync in background (every 24h)
// go func() {
// 	// Run once on startup
// 	log.Println("⏳ Initializing Car Data Sync...")
// 	if err := carSyncService.SyncData(context.Background()); err != nil {
// 		log.Printf("❌ Initial Car Sync Failed: %v", err)
// 	}

// 	ticker := time.NewTicker(24 * time.Hour)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		log.Println("⏰ Starting Scheduled Car Sync...")
// 		if err := carSyncService.SyncData(context.Background()); err != nil {
// 			log.Printf("❌ Scheduled Car Sync Failed: %v", err)
// 		}
// 	}
// }()

// ---------------------------------------------------------
// FEATURE B: RSS FEED CHECKER (Existing Logic)
// ---------------------------------------------------------

// // B1. Init Adapters
// rssFetcher := infra.NewGoFeedFetcher()
// stateStore := infra.NewFileStore("feed_state.json")
// translator := infra.NewGoogleTranslator()

// // B2. Init Notifier (Updated)
// // NOTE: You must update NewDiscordNotifier to accept the existing 'dg' session
// // instead of creating a new one internally.
// rssNotifier := discord_infra.NewDiscordNotifier(discordSession.Session, "")

// // B3. Init Use Case
// feedLogic := usecase.NewFeedChecker(rssFetcher, rssNotifier, stateStore, translator)

// ---------------------------------------------------------
// STARTUP & RUN LOOP
// ---------------------------------------------------------

// 4. Run RSS Ticker in a Goroutine (Background)
// rssTicker := time.NewTicker(time.Duration(cfg.RSSCfg.Interval) * time.Minute)
// defer rssTicker.Stop()

// Run immediately once on startup
// go feedLogic.Check(cfg.RSSCfg.Feeds)

// ---------------------------------------------------------
// FEATURE D: OBMETA SYNC (Scheduled)
// ---------------------------------------------------------

// Loop
// go func() {
// 	for range rssTicker.C {
// 		// feedLogic.Check(cfg.RSSCfg.Feeds)
// 	}
// }()
