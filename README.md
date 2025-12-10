# Desuwa
This bot is created for IDAC VN's server. Currently it's simply a RSS bot for Initial D's official news.

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