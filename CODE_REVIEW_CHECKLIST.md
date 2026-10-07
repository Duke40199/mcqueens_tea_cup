# Code Review Checklist — mcqueens_tea_cup

Tracking the findings from the 2026-10-07 review. Work top-down: Critical → High → Medium → Low → Architecture.

Legend: `[ ]` todo · `[x]` done · `[~]` partial

---

## ✅ Done

- [x] **SQL injection** in `GetCarWithSpecsByAliases` — parameterized the WHERE clause, parenthesized groups, guarded empty map. `internal/adapter/repository/car_repo.go`
- [x] **Broken `ob-ranking-cfg` table name** (High #5) — fixed as a side effect of the table-name migration (default now uses underscores). `internal/adapter/repository/ob_ranking_cfg_repo.go`
- [x] **Table names → config/.env** — all 10 tables across 9 repos moved to `DB_TABLE_*` env vars via `config.DatabaseTablesConfig`; removed the domain-leaking `StoreLocation.TableName()`.
- [x] **#2 Panic isolation + DM nil-deref guards** — see the Critical section; `recover` at all three goroutine entry points, all listed nil-derefs fixed.
- [x] **#3 Slice-bounds panic** in `GetListCarDetailByTAFormat` — `continue` added to the bracket-less guard.
- [x] **Owner ID → env** (`DISCORD_BOT_OWNER_ID`) — removed the hardcoded magic string; `IsRequestFromOwner` uses `h.OwnerID` and fails closed when unset.
- [x] **#4 DB constructor** — no more `log.Fatal`; errors returned + `PingContext` validates the connection at startup.

### 🎉 Critical tier complete (#1–#4)

---

## 🔴 Critical — crashes or security

- [x] **#2 No panic isolation (whole process crashes).** Added `recoverInteraction` + `interactionUser` helpers (`internal/adapter/discord/safety.go`); recover now wraps the router closure (`router.go`), `dispatch` (converts panics to a user-facing error edit), and the pagination component handler (`handler_helper.go`). All DM-reachable nil-derefs fixed:
  - [x] `IsRequestFromOwner` now uses `interactionUser(i)`. `internal/adapter/discord/handler.go:139`
  - [x] Pagination button filter nil-safe (message + user), via `interactionUser`. `internal/adapter/discord/handler_helper.go`
  - [x] `opt.Value.(string)` comma-ok + nil-checked `State.User`. `internal/adapter/discord/handler_shitpost.go`
  - [x] `HandleMiemebell` guards empty `data.Blocks` before `rand.Intn`. `internal/adapter/discord/handler_shitpost.go`
  - [x] Resolved user via `interactionUser` helper everywhere (incl. `handler_cfs.go`).
- [x] **#3 Slice-bounds panic** on bracket-less car input — added `continue` in the guard so it no longer falls through to `segaFormat[0:-1]`. `internal/domain/service/idac_car_service.go:44`
- [x] **#4 `log.Fatal` in DB constructor + no `Ping`.** Errors now wrapped+returned (caller `main.go` fails fast); added `PingContext` (5s) to validate the connection at startup, closing the pool on failure. `internal/adapter/database/postgres.go`

## 🟠 High

- [x] **#6 No context timeouts.** `dispatch` now builds a 30s `context.WithTimeout` per interaction (`commandTimeout`). `SegaIDACClient` interface takes `ctx` on all 8 methods; the client uses a shared `http.Client{Timeout: 15s}` + `NewRequestWithContext` (shared `makeHttpRequest` helper). AllNet client given a 15s timeout client too. All service/sync call sites thread `ctx`. `dispatch.go`, `port/client.go`, `sega_idac/client.go`, `all_net/client.go`, all services.
- [x] **#7 CFS double-responds + reports false success** — now defers ephemerally, does the DB write + channel post, then edits with the real outcome (success or specific error). No more double `InteractionRespond`. `internal/adapter/discord/handler_cfs.go`
- [x] **#8 Owner-alias does slow work before replying** — now defers first, then edits the deferred response after the DB write + `Session.User()` lookup. `internal/adapter/discord/handler_idac.go`
- [x] **#9 Meta-sync poll loop can spin forever** — hoisted the `maxPollingDuration` break to the top of the loop (fires during a Sega outage); fixed the header check to test the raw `strings.Index` before adding the prefix length (also removes a slice-out-of-range risk). `internal/domain/service/meta_sync_service.go`
- [x] **Pagination: global handler + data race.** Redesigned (`internal/adapter/discord/pagination.go`): one component handler registered once at the router; a mutex-guarded per-message `paginationStore` (keyed by message ID) holds pages/index/owner. No per-invocation `AddHandler`, no shared-variable race, bounded work per click; sessions auto-expire after 2 min (buttons cleared). Removed the old implementation + dead `stop` channel from `handler_helper.go`.

## 🟡 Medium

- [x] **Bulk upserts have no chunking** — added `chunkSlice` helper (`batch.go`, `pgMaxBulkRows=1000`); `UpsertCars`, `UpsertCarStyles`, `BulkUpsertStoreLocation` now batch within a transaction.
- [x] **Store upsert conflicts on `name` only** — **WON'T FIX** (decision 2026-10-08). The table enforces `UNIQUE (name)`, so `ON CONFLICT (name)` is correct; arcade store names are globally unique. A composite key would need a destructive migration (drop `UNIQUE(name)`, add `UNIQUE(name, sega_area_code)` with `NULLS NOT DISTINCT` since the column is nullable) for no real benefit. Left as-is. `internal/adapter/repository/idac_store_repo.go`
- [x] **Nil `*time.Location` panic risk** — switched to `time.FixedZone("JST", …)`; the remaining `LoadLocation` calls already check their error and fall back to it. `internal/domain/service/active_player_sync_service.go`
- [x] **Top-car match uses `continue` not `break`** — now `break`s on first match and falls back to the raw Sega name when no metadata match (no blank render). `internal/adapter/discord/handler_idac.go`
- [x] **Direct area codes silently become empty** — now falls back to the raw input when not an alias. `internal/adapter/discord/handler_idac_player.go`
- [x] **CFS message length uncapped** — `MaxLength: 1900` on the `message` option. `internal/adapter/discord/router.go`
- [x] **Context accepted but discarded** — `GetLatestCfsState` uses `QueryRowContext`; `CreateCfsState` now takes `ctx` (interface + caller updated). `internal/adapter/repository/cfs_state_repo.go`
- [x] **`SELECT *` positional scans + missing `rows.Err()`** — `rows.Err()` added everywhere; all `SELECT *` pinned to explicit columns (`idac_area_metadata`, `ob_ranking_cfg` both methods — schemas confirmed). This also fixed `GetBySegaID`'s scan-count mismatch (was erroring every call). `GetBySegaID` is still unused (dead code) — candidate for removal later.
- [x] **Weak config validation** — now validates `DISCORD_BOT_TOKEN`, `DB_{HOST,PORT,USER,NAME}`, `SEGA_IDAC_HOST` and reports all missing vars by their real names. `internal/config/config.go`
- [x] **`sslmode` now configurable** via `DB_SSLMODE` (default `require`, was hardcoded `disable`). Set `DB_SSLMODE=disable` to revert. `internal/adapter/database/postgres.go`, `internal/config/config.go`

## 🟢 Low

- [x] Missing `rows.Err()` after loops — added to `ranking_cfg_repo.go`, `ob_ranking_cfg_repo.go`, `idac_ob_sync_area_cfg_repo.go`, `idac_area_metadata.go`.
- [x] `sql.ErrNoRows` returned as an error — now `ErrNoRows` returns `(zero, false, nil)` / `(nil, nil)` so "not found" isn't a real error. Fixed a latent dead-branch in `ResolvePlayer` (the raw SQL error was leaking past the `!isFound` check, hiding the "couldn't find a matching alias" message). `alias_repo.go`, `cfs_state_repo.go`, `ob_ranking_cfg_repo.go`
- [x] Ignored `strconv.Atoi` error → now returns a wrapped parse error (fixed during #6). `internal/adapter/client/sega_idac/client.go`
- [x] Wrong `%d` verb for an `error` value → now `%w` (fixed during #6). `internal/adapter/client/sega_idac/client.go`
- [x] Hardcoded `"VN"` player area → now shows the resolved `playerArea`, mapped via `AreaDisplayNameByCode` to a friendly name. `internal/adapter/discord/handler_idac_player.go`
- [x] Dead `cfsCounter`/`cfsMutex`/`init()` reading `cfs_counter.txt` on every startup — removed during the #7 rewrite. `internal/adapter/discord/handler_cfs.go`
- [x] `fmt.Print*` error/diagnostic output → `log.*` across clients, repos, services and handlers (9 files, ~21 calls). `fmt` kept only where still used (Errorf/Sprintf); removed from `idac_car_service.go`.
- [x] **Structured logger + tracer** (2026-10-08) — added `pkg/logger` (slog JSON: `timestamp/level/msg/trace_id/user_id/source`, plus `error/stack_trace`) and `pkg/tracer` (per-unit-of-work trace IDs via context; Discord-adapted, no Gin/X-Ray). Wired `logger.Init(APP_ENV)` in main; trace IDs created at every entry point (`dispatch`, both cron loops, cfs/miemebell/player-alias/pagination/autocomplete hand-rolled handlers) with `user_id` where known.
  - **Full migration done:** all ~90 `log.*` calls → `logger.*` across clients, repos, services, handlers, main. Added `ctx` params to the 4 ctx-less repo methods (`AliasRepository.GetByAliasKey`/`GetByIgnAndAreaCode`, `OBRankingCfgRepository.GetBySegaID`/`GetRankingCfgMap`) + their interfaces + callers.
  - **Kept as `log.Fatal`:** 4 startup calls in `main.go` + 1 in `discord_session.go` (no `logger.Fatal`; they run before any trace and must exit the process).
  - **Dev console output** (`pkg/logger/pretty.go`): when `APP_ENV=dev`, logs render as colored, compact, human-readable lines (short `file:line`, empty attrs omitted, ANSI auto-disabled when piped); prod stays JSON. `APP_ENV` added to `.env`/`.env.example`.
- [x] `formatQuery` debug helper — deleted (its only callers were the bulk upserts rewritten in the chunking change). `internal/adapter/repository/car_repo.go`
- [x] ~~Save-after-unlock race in the JSON alias store~~ — moot; `alias.go` was deleted (2026-10-08).

## 🏛️ Architecture (larger refactors)

- [x] **Dead/duplicated code & broken second binary.**
  - [x] **RSS feature removed** (2026-10-08): deleted root `main.go` (broken RSS bot), `entity/interface.go` (unused `RSSFetcher`/`Notifier`/`StateStore`/`Translator`), the dead `entity.Item` struct, the `gofeed` dependency (`go mod tidy`), and `RSS_FETCH_INTERVAL` from `.env`. `cmd/mckween/main.go` is now the only `package main`.
  - [x] **Removed unused JSON `alias.go`** (2026-10-08) — the pre-Postgres local-file store (`JSONAliasStore`), confirmed unreferenced. Postgres `alias_repo.go` is now the sole `AliasRepository`.
- [x] **Dependency inversion fixed** (2026-10-08) — moved all 9 repository interfaces from `adapter/database` to `domain/port/repository.go`; deleted `adapter/database/repository.go`. `domain/` no longer imports `adapter/database`; only `main.go` does (for `NewPostgresDBConn`). Adapters now implement `port.*`, so the dependency arrow points inward.
- [x] **Transport removed from domain** (2026-10-08) — introduced a `port.Notifier` (`BotMessages` + `SyncPages`) implemented by `adapter/discord/notifier.go` (`DiscordNotifier`). `meta_sync_service` and `active_player_sync_service` no longer import `discordgo` — they depend on `port.Notifier`. The channel reconcile logic (reverse → edit/send/delete), previously duplicated in both services, now lives once in the adapter. `discordgo` is confined to the `adapter/discord` package.
- [x] **Two handler conventions** — **resolved, no full unification** (decision 2026-10-08). The concrete defects this item cited are already fixed: panic recovery (`recoverInteraction`) wraps every handler at the router; CFS defers-first (no double-response); owner-alias defers before slow work. Full `dispatch` routing was rejected because the hand-rolled handlers are instant (`status`/`mckween`/`miemebell`/`nuhuh`, several with GIF attachments) or ephemeral (`cfs`) — forcing them through the defer→edit lifecycle would degrade UX and require generalizing `dispatch` for no bug-level benefit.

## 🧹 Housekeeping (optional)

- [x] Config table names wrapped with `pq.QuoteIdentifier` via a `quoteIdent` helper (`repository/ident.go`), applied in all repo constructors. Belt-and-suspenders (also makes a hyphenated name safe).
- [x] `lib/pq` is now a direct dependency in `go.mod` (promoted by `go mod tidy`).
- [x] `README.md` file-tree updated to reflect the current structure (removed root `main.go`/RSS, renamed/deleted files; added `pkg/`, `port/`, `notifier.go`, etc.) + a note on the dependency direction.
- [x] Replaced the generic Node/Java/Python `.gitignore` with a Go-appropriate one.

---

## Suggested order

1. **This week:** #2, #3, #4 (crashes/security — tightly scoped, high-impact).
2. **Next:** #6, #7, #8, #9 + pagination race (reliability).
3. **Then:** the Medium batch + decide the fate of the root `main.go`/RSS code (Arch item 1).
4. **Ongoing:** Architecture items 2–4.
