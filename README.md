# McQueen's Tea Cup
A multi-purpose bot for Vietnam's Initial D The Arcade server. Main features include:
- Spamming random things.
- Initial D The Arcade features:
  - Get Time Attack (TA) by track with selective variants / players / countries / cars.
  - Compare TA results between players.
  - Cron-job to crawl a list of most used cars / active players in Online Battles (OB).

## Technical Info:
- Programming languages / Frameworks: Go, Discord Go SDK, PostgreSQL
- Database hosting: Supabase
- CI / CD: Github Actions / Railway
## Project Structure
- Design Pattern: Clean / Hexagonal Architecture. Dependencies point inward:
  `adapter` → `domain/port` ← `domain/service`. The domain never imports an
  adapter; `cmd/mckween/main.go` is the composition root that wires everything.
- File structure:
```
mcqueens_tea_cup/
├── .env.example                 # sample env (real .env is gitignored)
├── go.mod / go.sum
├── README.md
├── cmd/
│   └── mckween/
│       └── main.go              # entry point / dependency wiring
├── internal/
│   ├── adapter/                 # outward edges (DB, HTTP clients, Discord)
│   │   ├── client/
│   │   │   ├── all_net/client.go
│   │   │   └── sega_idac/client.go
│   │   ├── database/
│   │   │   └── postgres.go      # connection pool (NewPostgresDBConn)
│   │   ├── discord/
│   │   │   ├── discord_session.go
│   │   │   ├── dispatch.go      # slash-command lifecycle (defer → run → reply)
│   │   │   ├── handler*.go      # command handlers (idac, cfs, shitpost, …)
│   │   │   ├── notifier.go      # port.Notifier impl (channel publishing)
│   │   │   ├── pagination.go    # button-paged messages
│   │   │   ├── router.go / router_definition.go
│   │   │   └── resource/        # embedded gifs / json
│   │   └── repository/          # port.*Repository impls (Postgres)
│   │       ├── alias_repo.go  car_repo.go  cfs_state_repo.go
│   │       ├── idac_area_metadata.go  idac_ob_sync_area_cfg_repo.go
│   │       ├── idac_store_repo.go  ob_ranking_cfg_repo.go
│   │       ├── ranking_cfg_repo.go  ta_time_metadata_repo.go
│   │       └── batch.go         # bulk-insert chunking helper
│   ├── config/
│   │   └── config.go            # env-driven config (incl. table names)
│   └── domain/                  # core — no adapter imports
│       ├── entity/              # domain types (idac, cars, views, …)
│       ├── port/                # interfaces: client, repository, notifier, service
│       └── service/             # business logic (sync, meta, player, team, …)
└── pkg/                         # shared, project-agnostic helpers
    ├── logger/                  # slog wrapper (JSON / colored dev console)
    ├── tracer/                  # context-propagated trace IDs
    └── utils/                   # generic helpers (ChunkSlice, …)
```
