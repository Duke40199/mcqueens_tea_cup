package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"McQueens_Tea_Cup/internal/adapter/client/all_net"
	"McQueens_Tea_Cup/internal/adapter/client/sega_idac"
	"McQueens_Tea_Cup/internal/adapter/database"
	discord_handler "McQueens_Tea_Cup/internal/adapter/discord"
	"McQueens_Tea_Cup/internal/adapter/repository"
	"McQueens_Tea_Cup/internal/config"
	"McQueens_Tea_Cup/internal/domain/service"
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
	aliasRepo := repository.NewAliasRepo(dbConn)
	obRankingCfgRepo := repository.NewOBRankingCfgRepository(dbConn)
	areaRepo := repository.NewAreaRepository(dbConn)
	taTimeMetadataRepo := repository.NewTATimeMetadataRepository(dbConn)
	rankingCfgRepo := repository.NewRankingCfgRepo(dbConn)
	cfsStateRepo := repository.NewCfsStateRepository(dbConn)
	carRepo := repository.NewCarRepository(dbConn)
	areaMetadataRepo := repository.NewIDACAreaMetadataRepository(dbConn)
	storeLocationRepo := repository.NewAllNetStoreLocationsRepository(dbConn)

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
	segaClient := sega_idac.NewSegaIDACClient(cfg)
	allNetClient := all_net.NewAllNetClient(cfg)
	// 5. Init services
	idacCarService := service.NewIDACCarService(cfg, carRepo, segaClient)
	idacTimeAttackMetadata := service.NewIDACTimeAttackService(cfg, taTimeMetadataRepo, segaClient)
	storeLocationService := service.NewIDACStoreLocationService(cfg, allNetClient, segaClient, storeLocationRepo)
	idacAreaService := service.NewIDACAreaService(cfg, areaMetadataRepo, segaClient)
	obRankingService := service.NewIDACOBRankingService(segaClient, obRankingCfgRepo)
	teamService := service.NewIDACTeamService(segaClient)
	playerService := service.NewIDACPlayerService(segaClient, aliasRepo, rankingCfgRepo, obRankingCfgRepo)
	obMetaService := service.NewIDACOBMetaService(segaClient, carRepo)
	metaLogic := service.NewMetaLogicService(obMetaService)

	cmdHandler := discord_handler.NewHandler(
		discordSession.Session,
		aliasRepo,
		obRankingCfgRepo,
		rankingCfgRepo,
		carRepo,
		taTimeMetadataRepo,
		cfsStateRepo,
		metaLogic,
		storeLocationService,
		idacCarService,
		idacTimeAttackMetadata,
		idacAreaService,
		obRankingService,
		teamService,
		playerService,
		obMetaService,
		segaClient,
		allNetClient,
	)
	// 6. Register Commands & Event Handlers
	if err := cmdHandler.RegisterCommands(); err != nil {
		log.Printf("⚠️ Failed to register commands: %v", err)
	}

	// 7.a. Cron Job: Online Battle Car Meta
	metaSync := service.NewMetaSyncService(discordSession.Session, metaLogic, cfg.MetaSyncCfg)
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
	activePlayersSync := service.NewActivePlayerSyncService(discordSession.Session, segaClient, areaRepo, obRankingCfgRepo, metaLogic, cfg.ActivePlayersSyncCfg)
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

	// 7.c. Health server
	// The bot is outbound-only (Discord gateway, Sega/AllNet HTTP, Postgres) and
	// never accepts inbound traffic. Web-service hosts like Render port-scan the
	// container and route health checks to an open port, so we expose a tiny HTTP
	// endpoint purely to satisfy that. Bind 0.0.0.0 (not localhost) or the scan
	// won't detect it. Deployed as a background worker, this is simply unused.
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000" // Render's default
	}
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	healthSrv := &http.Server{Addr: "0.0.0.0:" + port, Handler: healthMux}
	go func() {
		log.Printf("🩺 Health server listening on :%s", port)
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ Health server error: %v", err)
		}
	}()

	// 8. Init graceful shutdown
	log.Println("✅ Bot is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Health server shutdown error: %v", err)
	}
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
