# McQueen's Tea Cup
A multi-purpose bot for Vietnam's Initial D The Arcades server. Main features include:
- Spamming random things.
- Initial D The Arcade features:
  - Get Time Attack (TA) by track with selective variants / players / countries / cars.
  - Compare TA results between players.
  - Cron-job to crawl a list of most used cars / active players in Online Battles (OB).

## Project Structure
This project is strictly following the Clean Architecture rules. This might change later when needed.
```rss-bot/
├── cmd/
│   └── bot/
│       └── main.go           # Entry point (Wiring & Dependency Injection)
├── internal/
│   ├── domain/               # Pure data structs & Interface definitions
│   │   ├── models.go
│   │   └── interfaces.go
│   ├── usecase/              # Business logic (The "checkFeeds" logic)
│   │   └── feed_checker.go
│   └── infrastructure/       # Implementations of interfaces
│       ├── config/           # Config loading
│       ├── discord/          # DiscordGo wrapper
│       ├── rss/              # GoFeed wrapper
│       ├── storage/          # JSON state file handling
│       └── translator/       # Google Translate HTTP calls
├── config.json
└── go.mod
```
