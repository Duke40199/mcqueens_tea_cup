package main

import (
	"log"
	"time"

	"McQueens_Tea_Cup/internal/infra"
	"McQueens_Tea_Cup/internal/infra/config"
	"McQueens_Tea_Cup/internal/usecase"
)

func main() {
	// 1. Infrastructure: Load Config
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Infrastructure: Initialize Adapters
	rssFetcher := infra.NewGoFeedFetcher()
	stateStore := infra.NewFileStore("feed_state.json")
	translator := infra.NewGoogleTranslator()

	// Discord needs session initialization
	discordNotifier, err := infra.NewDiscordNotifier(cfg.Token)
	if err != nil {
		log.Fatal(err)
	}
	defer discordNotifier.Close()

	// 3. Use Case: Inject Dependencies
	botLogic := usecase.NewFeedChecker(rssFetcher, discordNotifier, stateStore, translator)

	// 4. Run Loop
	log.Println("Bot running...")
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Minute)

	// Initial Run
	go botLogic.Check(cfg.Feeds)

	for range ticker.C {
		botLogic.Check(cfg.Feeds)
	}
}
