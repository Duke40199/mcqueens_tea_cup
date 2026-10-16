-- bot_command_guides: authored help articles surfaced by /idac help.
--
-- command_key is the stable lookup key used by the command and autocomplete
-- (matches how the command is invoked, e.g. "idac time-attack").
-- content is Discord-flavoured markdown and is rendered live, so editing a row
-- takes effect without restarting the bot. New rows appear in the /idac help
-- autocomplete after the next restart (the list is cached at startup).
-- type groups a guide by command family (e.g. 'IDAC', 'COMMON').

CREATE TABLE IF NOT EXISTS bot_command_guides (
    id          SERIAL PRIMARY KEY,
    command_key TEXT        NOT NULL UNIQUE,
    title       TEXT        NOT NULL,
    content     TEXT        NOT NULL,
    type        TEXT        NOT NULL DEFAULT 'IDAC',
    sort_order  INT         NOT NULL DEFAULT 0,
    enabled     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- For tables created before the type column existed.
ALTER TABLE bot_command_guides
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'IDAC';

-- Keep updated_at fresh on edits.
CREATE OR REPLACE FUNCTION set_bot_command_guides_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_bot_command_guides_updated_at ON bot_command_guides;
CREATE TRIGGER trg_bot_command_guides_updated_at
    BEFORE UPDATE ON bot_command_guides
    FOR EACH ROW
    EXECUTE FUNCTION set_bot_command_guides_updated_at();

-- ---------------------------------------------------------------------------
-- Seed: all /idac command guides (type = 'IDAC')
-- ---------------------------------------------------------------------------

-- /idac time-attack
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac time-attack',
    '/idac time-attack',
    $guide$**Get the Time Attack rankings for a course**, optionally narrowed to a specific car and tune.

__How to use__
` /idac time-attack track:<track> area:<country> [car:<car>] [spec:<spec>] `

__Options__
- **track** *(required)* — the course to look up. Start typing to search (e.g. `akina`, `myogi`) and pick from the list.
- **area** *(required)* — the country/region leaderboard. Type to search by name or ISO code (e.g. `vn`, `japan`) and pick a match.
- **car** *(optional)* — filter to one car. Type the maker or model to search (e.g. `ae86`, `evo`).
- **spec** *(optional)* — a tune/spec of the selected car. **Pick a car first**, then choose its spec.

__Tips__
- Leave **car** and **spec** empty to see the overall fastest times on the course.
- The **area** list is sorted World first, then countries A–Z, then prefectures A–Z.
- Need the grade time thresholds instead? Use ` /idac track-details `.$guide$,
    'IDAC', 10, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac track-details
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac track-details',
    '/idac track-details',
    $guide$**Show a course's rank time thresholds** — the exact times you need for each grade, fastest first.

__How to use__
` /idac track-details track:<track> `

__Options__
- **track** *(required)* — the course to inspect. Type to search (e.g. `akina`, `irohazaka`) and pick from the list.

__What you get__
- The **Top 3 most used cars** on the course (based on the top 1000 global results).
- A paginated list of **grade → required time** thresholds, with the course image as a thumbnail.

__Tips__
- Use the ◀️ / ▶️ buttons to page through the thresholds.
- Want the actual leaderboard times instead? Use ` /idac time-attack `.$guide$,
    'IDAC', 20, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac time-attack-compare
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac time-attack-compare',
    '/idac time-attack-compare',
    $guide$**Compare two players' Time Attack times** on the same course, side by side.

__How to use__
` /idac time-attack-compare track:<track> player1:<ign|@user> player2:<ign|@user> [area1:<country>] [area2:<country>] `

__Options__
- **track** *(required)* — the course to compare on. Type to search and pick from the list.
- **player1** *(required)* — first player, by in-game name or Discord @mention.
- **player2** *(required)* — second player, by in-game name or Discord @mention.
- **area1** *(optional)* — Player 1's country. Type to search.
- **area2** *(optional)* — Player 2's country. Defaults to Player 1's area if left empty.

__Tips__
- Linking a Discord user to an IGN (via ` /idac player-alias `) lets you @mention them here instead of typing their IGN.$guide$,
    'IDAC', 30, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac team
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac team',
    '/idac team',
    $guide$**Get Team class & rankings**, optionally filtered by class and country.

__How to use__
` /idac team rank:<class> [country:<country>] [limit:<1-25>] `

__Options__
- **rank** *(required)* — team class to show: **All**, **Master**, **Platinum**, **Gold**, or **Silver**.
- **country** *(optional)* — filter to one country. Type to search.
- **limit** *(optional)* — number of results to show (1–25, default 10).$guide$,
    'IDAC', 40, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac ob-ranking
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac ob-ranking',
    '/idac ob-ranking',
    $guide$**Get the Online Battle (OB) rankings**, optionally filtered by rank, country, and round.

__How to use__
` /idac ob-ranking [rank:<rank>] [area:<country>] [round:<number|all>] [limit:<1-25>] `

__Options__
- **rank** *(optional)* — filter by rank: **Pride**, **Ruby**, **Emerald**, **Red**, **Blue**, or **Green**.
- **area** *(optional)* — filter by country. Type to search.
- **round** *(optional)* — a round number, or `all`. Defaults to the **current round** when empty.
- **limit** *(optional)* — number of results to show (1–25, default 10).

__Tips__
- Curious which cars dominate OB right now? See ` /idac ob-meta `.$guide$,
    'IDAC', 50, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac ob-meta
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac ob-meta',
    '/idac ob-meta',
    $guide$**Show the current Online Battle meta cars** — the most-used cars among top OB players.

__How to use__
` /idac ob-meta `

__Options__
- None. Just run the command.

__Tips__
- Pair this with ` /idac ob-ranking ` to see who's at the top and what they drive.$guide$,
    'IDAC', 60, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac profile
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac profile',
    '/idac profile',
    $guide$**Find a player's rank** on a specific course and variant.

__How to use__
` /idac profile user:<ign|@user> [track:<track>] [variant:<variant>] `

__Options__
- **user** *(required)* — the player, by in-game name or Discord @mention.
- **track** *(optional)* — the course to look up.
- **variant** *(optional)* — direction/condition (Downhill, Uphill, etc). **Pick a track first**, then choose its variant.

__Tips__
- Link a Discord user to an IGN with ` /idac player-alias ` so you can @mention them here.$guide$,
    'IDAC', 70, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac player-info-tournament
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac player-info-tournament',
    '/idac player-info-tournament',
    $guide$**Get player info for tournaments**, filtered by country and rank.

__How to use__
` /idac player-info-tournament [area:<country>] [rank:<rank>] [limit:<1-25>] `

__Options__
- **area** *(optional)* — filter by country. Type to search.
- **rank** *(optional)* — filter by rank: **Pride**, **Ruby**, **Emerald**, **Red**, **Blue**, or **Green**.
- **limit** *(optional)* — number of results to show (1–25, default 10).$guide$,
    'IDAC', 80, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac player-alias (owner only)
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac player-alias',
    '/idac player-alias',
    $guide$**Manage player aliases** — link a Discord user (or a custom tag) to an in-game name and area so you can reference them by @mention or tag elsewhere.

> ⚠️ This command is **owner-only**.

__How to use__
` /idac player-alias set user:<@user> ign:<ign> area:<area> `
` /idac player-alias set-custom tag:<tag> ign:<ign> area:<area> `

__Subcommands__
- **set** — link a Discord **user** to an IGN and area.
- **set-custom** — create a custom **tag** (e.g. `rival`, `alt`) for a non-Discord player, linked to an IGN and area.

__Tips__
- Once set, aliases work anywhere a player input is accepted, e.g. ` /idac profile ` and ` /idac time-attack-compare `.$guide$,
    'IDAC', 90, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac list-store-location
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac list-store-location',
    '/idac list-store-location',
    $guide$**List arcade stores** from All.Net for a given country/area.

__How to use__
` /idac list-store-location area-select:<area> `

__Options__
- **area-select** *(required)* — the country/area to list stores for. Type to search (e.g. `vn`, `jp`) and pick a match.$guide$,
    'IDAC', 100, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;

-- /idac help
INSERT INTO bot_command_guides (command_key, title, content, type, sort_order, enabled)
VALUES (
    'idac help',
    '/idac help',
    $guide$**How to use the bot's commands.** Shows a guide for a command, or lists all available guides.

__How to use__
` /idac help [command:<command>] `

__Options__
- **command** *(optional)* — pick a command to read its guide. Leave it empty to list every available guide.$guide$,
    'IDAC', 110, TRUE
)
ON CONFLICT (command_key) DO UPDATE
SET title = EXCLUDED.title, content = EXCLUDED.content, type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order, enabled = EXCLUDED.enabled;
