package main

import (
	"context"
	"fmt"
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
	"McQueens_Tea_Cup/pkg/logger"
	"McQueens_Tea_Cup/pkg/tracer"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	// Init structured logger (level from APP_ENV: "dev" => debug, else info).
	logger.Init(os.Getenv("APP_ENV"))

	// 2. Connect to DB
	dbConn, err := database.NewPostgresDBConn(cfg.DatabaseCfg)
	if err != nil {
		log.Fatal("Postgres database connection error: ", err)
	}

	// 3. Init Repositories
	aliasRepo := repository.NewAliasRepo(dbConn, cfg.DatabaseTablesCfg)
	obRankingCfgRepo := repository.NewOBRankingCfgRepository(dbConn, cfg.DatabaseTablesCfg)
	areaRepo := repository.NewAreaRepository(dbConn, cfg.DatabaseTablesCfg)
	taTimeMetadataRepo := repository.NewTATimeMetadataRepository(dbConn, cfg.DatabaseTablesCfg)
	rankingCfgRepo := repository.NewRankingCfgRepo(dbConn, cfg.DatabaseTablesCfg)
	cfsStateRepo := repository.NewCfsStateRepository(dbConn, cfg.DatabaseTablesCfg)
	carRepo := repository.NewCarRepository(dbConn, cfg.DatabaseTablesCfg)
	areaMetadataRepo := repository.NewIDACAreaMetadataRepository(dbConn, cfg.DatabaseTablesCfg)
	storeLocationRepo := repository.NewAllNetStoreLocationsRepository(dbConn, cfg.DatabaseTablesCfg)

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
		cfg.DiscordCfg.BotOwnerID,
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
		logger.Error(context.Background(), "failed to register commands", err)
	}

	// Notifier abstracts Discord message publishing for the sync services.
	notifier := discord_handler.NewDiscordNotifier(discordSession.Session)

	// 7.a. Cron Job: Online Battle Car Meta
	metaSync := service.NewMetaSyncService(notifier, metaLogic, cfg.MetaSyncCfg)
	go func() {
		for {
			// Fresh trace ID per sync run so its logs can be correlated.
			ctx := tracer.NewContext(context.Background())
			logger.Info(ctx, "starting OBMeta sync")
			metaLogic.SleepUntilNextSync(ctx, cfg.MetaSyncCfg.DowntimeStart, cfg.MetaSyncCfg.DowntimeEnd, cfg.MetaSyncCfg.DowntimeTZ)
			detectTime, err := metaSync.Sync(ctx)
			if err != nil {
				logger.Error(ctx, "scheduled OBMeta sync failed", err)
			} else {
				logger.Info(ctx, fmt.Sprintf("OBMeta sync detected data at: %s", detectTime))
			}
		}
	}()

	// 7.b. Cron Job: Online Battle Active Players
	activePlayersSync := service.NewActivePlayerSyncService(notifier, segaClient, areaRepo, obRankingCfgRepo, metaLogic, cfg.ActivePlayersSyncCfg)
	go func() {
		for {
			// Fresh trace ID per sync run so its logs can be correlated.
			ctx := tracer.NewContext(context.Background())
			logger.Info(ctx, "starting active players SEA sync")
			metaLogic.SleepUntilNextSync(ctx, cfg.ActivePlayersSyncCfg.DowntimeStart, cfg.ActivePlayersSyncCfg.DowntimeEnd, cfg.ActivePlayersSyncCfg.DowntimeTZ)
			detectTime, err := activePlayersSync.Sync(ctx)
			if err != nil {
				logger.Error(ctx, "scheduled active players sync failed", err)
			} else {
				logger.Info(ctx, fmt.Sprintf("active players sync detected data at: %s", detectTime))
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
		logger.Info(context.Background(), fmt.Sprintf("health server listening on :%s", port))
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(context.Background(), "health server error", err)
		}
	}()

	// 8. Init graceful shutdown
	logger.Info(context.Background(), "bot is running, press CTRL-C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	logger.Info(context.Background(), "gracefully shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := healthSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error(context.Background(), "health server shutdown error", err)
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
