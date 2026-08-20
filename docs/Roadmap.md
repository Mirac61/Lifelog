# lifelog — Build Roadmap

Personal data warehouse in Go. One binary, SQLite, collectors pull from external
sources into a unified event store.

**Rule for this project:** build the thinnest possible vertical slice first
(Phase 1–7). Abstract only after the second collector exists.

---

## Phase 0 — Decide and prepare

Time: ~30 min

- [x] Pick the final project name (currently `lifelog`)
- [x] Create GitHub repo, private for now
- [x] Add `.gitignore` (Go template + `*.db`, `*.db-wal`, `*.db-shm`, `.env`)
- [x] Add `LICENSE` — AGPL-3.0 if you plan to open source, otherwise skip for now (for noe private)
- [x] Create GitHub Personal Access Token, scope `read:user` — store in `.env`

**Done when:** empty repo cloned locally, token in `.env`, `.env` is gitignored.

---

## Phase 1 — Skeleton

Time: ~1 h · ~200 LOC

- [x] `go mod init github.com/<user>/lifelog`
- [x] `cmd/lifelog/main.go` — entry point, nothing but startup
- [x] `internal/config/config.go` — struct loaded from environment
      (`DatabasePath`, `GitHubToken`, `ListenAddress`)
- [x] Load `.env` at startup (`github.com/joho/godotenv`)
- [x] `log/slog` handler, level from environment
- [x] Open SQLite with `modernc.org/sqlite`, connection string:
      `file:data.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)`
- [x] Ping the database, log success, exit cleanly on `SIGINT`

**Done when:** `go run ./cmd/lifelog` creates `data.db` and logs a startup line.

**Watch out:** without WAL mode a running sync blocks all reads. Set it in the
connection string, not with a separate `PRAGMA` statement — the pool opens more
than one connection and a manual pragma only applies to whichever one ran it.

---

## Phase 2 — Schema and migrations

Time: ~1 h · ~150 LOC SQL

- [x] Add `github.com/pressly/goose/v3`
- [x] `internal/store/migrations/00001_init.sql`
- [x] Embed the migrations directory with `embed.FS`
- [x] Run migrations on startup, before anything else touches the database

Schema:

```sql
-- +goose Up
CREATE TABLE raw_payloads (
    id          INTEGER PRIMARY KEY,
    source      TEXT NOT NULL,
    external_id TEXT NOT NULL,
    fetched_at  TEXT NOT NULL,
    payload     TEXT NOT NULL,
    UNIQUE (source, external_id)
);

CREATE TABLE events (
    id          INTEGER PRIMARY KEY,
    source      TEXT NOT NULL,
    type        TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    local_date  TEXT NOT NULL,
    value       REAL,
    unit        TEXT,
    meta        TEXT,
    raw_id      INTEGER REFERENCES raw_payloads (id) ON DELETE CASCADE
);

CREATE INDEX idx_events_date_type ON events (local_date, type);
CREATE INDEX idx_events_raw ON events (raw_id);

CREATE TABLE sync_state (
    source       TEXT PRIMARY KEY,
    last_sync_at TEXT NOT NULL,
    last_error   TEXT
);
```

**Done when:** `sqlite3 data.db ".schema"` shows all three tables and
`goose_db_version`.

**Watch out:** `occurred_at` is UTC, `local_date` is the day the event belongs to
in your local timezone. Sleep that ends at 07:12 belongs to the wake-up day.
Decide this rule once, write it down, apply it in every collector.

---

## Phase 3 — Explore the GitHub API by hand

Time: ~30 min · 0 LOC

Do this **before** writing any Go. You want to know the response shape before
you model it.

- [ ] Open https://docs.github.com/en/graphql/overview/explorer
- [ ] Run the contributions query for the current year
- [ ] Save a sample response to `docs/samples/github-contributions.json`
- [ ] Note down: what is the natural unique key per record?

```graphql
query($from: DateTime!) {
  viewer {
    contributionsCollection(from: $from) {
      commitContributionsByRepository(maxRepositories: 100) {
        repository { nameWithOwner }
        contributions(first: 100) {
          nodes { occurredAt commitCount }
        }
      }
    }
  }
}
```

**Done when:** you can describe in one sentence what one row of this data is.

**Watch out:** this returns *daily aggregates per repository*, not individual
commits. Your `external_id` is therefore something like
`owner/repo:2026-08-19`, not a commit SHA. The REST events API would give you
individual commits but only covers 90 days — for full history you would need to
parse `git log` locally instead. Start with the aggregate.

---

## Phase 4 — Fetch into raw_payloads

Time: ~3 h · ~250 LOC

- [x] `internal/github/client.go` — HTTP client with a 30 s timeout,
      bearer token from config
- [ ] `internal/github/fetch.go` — execute the query, return
      `[]RawItem{ExternalID, OccurredAt, Payload json.RawMessage}`
- [ ] `internal/store/raw.go` — `InsertRaw` using
      `INSERT ... ON CONFLICT (source, external_id) DO UPDATE SET payload = excluded.payload, fetched_at = excluded.fetched_at`
- [ ] Wire it into `main.go` behind a temporary flag: `-sync`
- [ ] Write `sync_state` after a successful run

Do **not** parse the payload beyond extracting the ID and timestamp. Store the
JSON as it arrived.

**Done when:** `go run ./cmd/lifelog -sync` twice in a row produces the same row
count. That is your idempotency check.

**Watch out:** GitHub GraphQL returns HTTP 200 with an `errors` array on failure.
Check that array explicitly, a status code check is not enough.

---

## Phase 5 — Normalize into events

Time: ~2 h · ~200 LOC

Separate command, separate code path. This is the whole point of storing raw
payloads.

- [ ] `internal/github/normalize.go` — `RawItem` → `[]Event`
- [ ] `internal/store/events.go` — delete existing events for a `raw_id`,
      then insert the new ones (makes re-running safe)
- [ ] Flag `-normalize` that reprocesses every raw payload for a source
- [ ] Event shape: `type = "commit"`, `value = commitCount`, `unit = "count"`,
      `meta = {"repository": "owner/repo"}`

**Done when:** `-normalize` can run five times without changing the row count in
`events`, and the numbers match your GitHub profile graph.

**Watch out:** resist putting normalization inside the fetch. The moment they are
coupled, changing your event model means re-hitting the API.

---

## Phase 6 — First endpoint

Time: ~2 h · ~200 LOC

- [ ] Add Gin, `internal/api/router.go`
- [ ] `GET /health` — returns database status
- [ ] `GET /daily?type=commit&from=2026-01-01&to=2026-08-19`
- [ ] Validate date parameters, return 400 with a clear message on bad input
- [ ] Bind to `127.0.0.1` only

```sql
SELECT local_date, SUM(value) AS value
FROM events
WHERE type = ? AND local_date BETWEEN ? AND ?
GROUP BY local_date
ORDER BY local_date;
```

Response:

```json
[{"date": "2026-01-04", "value": 7}, {"date": "2026-01-05", "value": 3}]
```

**Done when:** `curl` against `/daily` returns real numbers.

**Watch out:** gaps. Days with no events simply do not appear. Decide now whether
the API fills them with zeros or the client does — filling them server-side is
usually less annoying.

---

## Phase 7 — See it

Time: ~1 h · 0 LOC

- [ ] `uvx datasette data.db` (or `pip install datasette`)
- [ ] Browse the `events` table, run a group-by query in the UI
- [ ] Optional: Grafana with the SQLite datasource plugin, one heatmap panel

**Done when:** you have looked at a chart of your own commit history that came
out of your own database.

**This is the milestone that matters.** Everything up to here is one weekend.
Stop, use it for two weeks, then decide whether to continue.

---

## Phase 8 — Second collector: Obsidian

Time: ~5 h · ~250 LOC

No API, no OAuth, no rate limits. Pure local file parsing.

- [ ] `internal/obsidian/scan.go` — walk `Cockpit/`, read markdown frontmatter
- [ ] `external_id` = relative file path + modification time
- [ ] Map frontmatter fields to event types (`book_finished`, `workout`, `todo_done`)
- [ ] Skip files without relevant frontmatter rather than erroring

**Done when:** Obsidian events sit next to GitHub events in the same table and
`/daily` works for both types.

**Watch out:** Syncthing conflict files (`*.sync-conflict-*.md`). Filter them out
or you will double-count.

---

## Phase 9 — Extract the abstraction

Time: ~3 h · ~200 LOC

Only now. You have two implementations and can see what they actually share.

- [ ] `internal/collector/collector.go` — the interface
- [ ] `internal/sync/runner.go` — iterate registered collectors, record
      `sync_state`, continue on individual failures
- [ ] Registry: a slice built in `main.go`
- [ ] Replace the `-sync` flag with `robfig/cron/v3` on a six-hour schedule
- [ ] Keep manual flags for debugging: `-sync-once=github`

```go
type Collector interface {
    Name() string
    Fetch(ctx context.Context, since time.Time) ([]RawItem, error)
    Normalize(raw RawItem) ([]Event, error)
}
```

**Done when:** adding a third collector means one new package and one line in the
registry.

**Watch out:** one collector failing must not abort the others. Log the error into
`sync_state.last_error` and move on.

---

## Phase 10 — OAuth collector: Strava

Time: ~10 h · ~400 LOC

The first genuinely annoying one.

- [ ] Register an application at https://developers.strava.com
- [ ] `golang.org/x/oauth2` config, scope `activity:read_all`
- [ ] One-time authorization flow: temporary local callback handler on
      `127.0.0.1:8081`, exchange code for tokens
- [ ] Persist the refresh token — encrypted, not plaintext in the database
- [ ] Wrap the `TokenSource` so refreshed tokens get written back
- [ ] Events: `workout` with `value = duration`, distance and heart rate in `meta`

**Done when:** the token survives a restart and refreshes on its own after the
access token expires (six hours).

**Watch out:** the token refresh writing back to storage is the part people get
wrong. A `TokenSource` refreshes in memory; if you never persist the new refresh
token you are fine with Strava (it rotates but the old one stays valid) and
broken with providers that invalidate immediately.

---

## Phase 11 — Apple Health

Time: ~12 h · ~400 LOC

- [ ] Export from the Health app → `export.zip`, several hundred MB of XML
- [ ] Parse with `encoding/xml` `Decoder.Token()` streaming — never `Unmarshal`
      the whole file
- [ ] Filter to record types you care about: `HKCategoryTypeIdentifierSleepAnalysis`,
      `HKQuantityTypeIdentifierStepCount`
- [ ] Merge adjacent sleep segments into one night before creating an event
- [ ] iOS Shortcut to automate the export into a synced folder

**Done when:** a night of sleep produces exactly one event, dated to the wake-up
day.

**Watch out:** sleep is exported as dozens of short segments per night with gaps.
Merging them is the actual work, the XML parsing is trivial.

---

## Repository layout

Single Go module, monorepo.

```
lifelog/
├── cmd/
│   └── lifelog/
│       └── main.go
├── internal/
│   ├── config/
│   ├── store/
│   │   ├── migrations/
│   │   ├── queries/          # sqlc input, once you add sqlc
│   │   ├── events.go
│   │   └── raw.go
│   ├── collector/
│   │   └── collector.go      # interface only, from phase 9
│   ├── github/
│   ├── obsidian/
│   ├── strava/
│   ├── sync/
│   │   └── runner.go
│   └── api/
│       ├── router.go
│       └── handlers.go
├── docs/
│   ├── decisions/            # short ADRs
│   └── samples/              # captured API responses
├── web/                      # only if you build your own dashboard
├── .env.example
├── go.mod
└── README.md
```

One package per source, all satisfying the same interface. Everything under
`internal/` so nothing is importable from outside — this is an application, not
a library.

---

## Dependencies

Direct, at the end of phase 11:

```
modernc.org/sqlite
github.com/pressly/goose/v3
github.com/gin-gonic/gin
github.com/robfig/cron/v3
github.com/joho/godotenv
golang.org/x/oauth2
```

Build tools, not runtime dependencies: `sqlc`, `goose` CLI.

Deliberately absent: ORM, PostgreSQL, Redis, message queue, auth framework,
Docker Compose with five services.

---

## Decisions worth writing down

Keep these as short files in `docs/decisions/`. Three sentences each is enough.

- [ ] Why raw payloads are stored separately from events
- [ ] How `local_date` is derived, per event type
- [ ] What `value` and `unit` mean for each event type
- [ ] Why SQLite over PostgreSQL
- [ ] How secrets are stored

The third one is the one you will thank yourself for. By the fourth collector you
will not remember whether a workout's `value` was minutes or seconds.

---

## Recurring gotchas

| Problem | Where it bites | Fix |
| --- | --- | --- |
| Day boundaries | Sleep, late-night commits | One documented rule, applied everywhere |
| Timezone drift | Travel, DST | Store UTC plus an explicit `local_date` |
| Duplicate events on re-sync | Every collector | `UNIQUE (source, external_id)` plus upsert |
| Rate limits | GitHub, Strava | Respect `Retry-After`, sync every six hours not every five minutes |
| Token expiry | OAuth sources | Persist refresh tokens, test a cold restart |
| Large XML | Apple Health | Streaming decoder, never `Unmarshal` |
| Silent collector failure | After phase 9 | Record `last_error` in `sync_state`, surface it in `/health` |

---

## Security

This database is a precise profile of your life. Treat it accordingly.

- [ ] Bind the HTTP server to `127.0.0.1`, never `0.0.0.0`
- [ ] Remote access via Tailscale or WireGuard only, no port forwarding
- [ ] Secrets in `.env`, never committed — verify with `git log -p -- .env`
- [ ] Encrypt OAuth refresh tokens at rest
- [ ] Backup: `sqlite3 data.db ".backup backup.db"` on a schedule, or Litestream

---

## Milestone summary

| Phase | Deliverable | Hours |
| --- | --- | --- |
| 0–7 | GitHub commits visible in a chart | 10–15 |
| 8 | Obsidian data alongside it | 5 |
| 9 | Pluggable collectors, scheduled sync | 3 |
| 10 | Strava with working OAuth | 10 |
| 11 | Sleep and steps from Apple Health | 12 |

Commit after every phase. This project's payoff arrives in years, so visible
progress is what keeps it alive — and it ends up tracking exactly those commits.
