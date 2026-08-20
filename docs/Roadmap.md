# lifelog — Build Roadmap

Personal data warehouse in Go. One binary, SQLite, collectors pull from external
sources into a unified event store.

**Rule for this project:** build the thinnest possible vertical slice first
(Phase 1–7). Abstract only after the second collector exists.

---

## Phase 0 — Decide and prepare 

- [x] Pick the final project name (`lifelog`)
- [x] Create GitHub repo, private for now
- [x] Add `.gitignore` (Go template + `*.db`, `*.db-wal`, `*.db-shm`, `.env`)
- [x] Skip LICENSE while the repo is private
- [x] Create a **fine-grained** Personal Access Token — resource owner: self,
      repository access: all repositories, permission: Metadata read-only

**Watch out:** the token expires. Put the date in a calendar.

---

## Phase 1 — Skeleton 

- [x] `go mod init github.com/Mirac61/lifelog`
- [x] `cmd/lifelog/main.go` — entry point, nothing but startup
- [x] `internal/config/config.go` — struct loaded from environment
- [x] Load `.env` at startup (`github.com/joho/godotenv`)
- [x] `log/slog` default handler
- [x] Open SQLite with `modernc.org/sqlite`, WAL mode via connection string
- [x] Ping the database, log success, exit cleanly on SIGINT

**Watch out:** the `.env` file is only read by the Go binary. For shell commands
like curl, load it explicitly: `set -a; source .env; set +a`.

---

## Phase 2 — Schema and migrations 

- [x] Add `github.com/pressly/goose/v3`
- [x] `internal/store/migrations/00001_init.sql`
- [x] Embed the migrations directory with `//go:embed`
- [x] Run migrations on startup, before anything else touches the database

Three tables: `raw_payloads`, `events`, `sync_state`.

**Watch out:** `//go:embed` must sit directly above the variable with no blank
line, or it is treated as a plain comment and the FS stays empty.

---

## Phase 3 — Explore the GitHub API by hand 

- [x] Run the contributions query and inspect the response
- [x] Decide the natural unique key: `contributions:<year>`
- [ ] Save a sample response to `docs/samples/github-contributions.json`

**Note:** the GraphQL Explorer was removed from GitHub's docs in November 2025.
Use curl, or a client like Altair/Insomnia pointed at
`https://api.github.com/graphql`. The schema reference lives at
https://docs.github.com/en/graphql/reference.

**Decision made:** use `contributionCalendar` rather than
`commitContributionsByRepository`. The calendar returns a flat list of
`{date, contributionCount}`, which maps almost directly onto an events row.
The per-repository breakdown becomes a **separate query later** — a second
`external_id`, not a replacement.

**Watch out:** `contributionsCollection` accepts a maximum span of one year.
Query one year at a time, or use GraphQL aliases to request several years in a
single request.

---

## Phase 4 — Fetch into raw_payloads 

- [x] `internal/github/client.go` — HTTP client with a 30 s timeout, bearer token
- [x] `internal/github/fetch.go` — `FetchContributions(ctx, year)`
- [x] `internal/collector/collector.go` — shared `RawItem` type
- [x] `internal/store/raw.go` — `InsertRaw` with `ON CONFLICT ... DO UPDATE`
- [x] Wire it into `main.go` behind `-sync`
- [ ] Write `sync_state` after a successful run — **still open**

**Watch out:** GitHub GraphQL returns HTTP 200 with an `errors` array on failure.
Check that array explicitly.

**Watch out:** struct fields need `json:"..."` tags on *every* field that crosses
the JSON boundary. A missing tag on `Variables` sends `Variables` instead of
`variables`, and the server reports it as an invalid variable value.

---

## Phase 5 — Normalize into events 

- [x] `internal/github/normalize.go` — payload → `[]collector.Event`
- [x] `internal/store/raw.go` — `ListRaw` reads payloads back out
- [x] `internal/store/events.go` — `InsertEvents`, delete-then-insert per `raw_id`,
      wrapped in a transaction
- [x] Flag `-normalize` that reprocesses every raw payload for a source

**Verified:** 113 events, `SUM(value) = 653`, matching `totalContributions`
from the payload.

**Decision made:** days with a count of zero are skipped. The absence of a row
means no activity. This keeps the table at roughly a third of the size and
removes the future-dated zeros GitHub returns for the rest of the year.

---

## Phase 6 — Query layer and HTTP server

Time: ~3 h · ~250 LOC

**Stack decision: standard library, no framework.** Go 1.22+ `net/http` has
method-aware routing (`mux.HandleFunc("GET /path", h)`) and path parameters.
For a handful of routes with no middleware chain that is enough. Gin adds a
dependency for nothing here; Fiber/fasthttp would additionally break
compatibility with `http.Handler`, `httptest`, and `net/http/pprof` in exchange
for microseconds that are invisible next to a SQLite query.

- [ ] `internal/store/events.go` — `DailyTotals(ctx, db, eventType, from, to)`
      returning `[]DailyTotal{LocalDate string, Value float64}`
- [ ] `internal/store/events.go` — `ListTypes(ctx, db)` returning the distinct
      `type` values present, so the dashboard discovers sources on its own
- [ ] `internal/api/server.go` — `http.ServeMux`, bound to `cfg.ListenAddress`
- [ ] `GET /health` — database ping plus the last sync state per source
- [ ] Graceful shutdown: `srv.Shutdown(ctx)` driven by the existing SIGINT context
- [ ] Remove the `-sync`/`-normalize` early `return` path so the server actually
      starts when no flag is given

```sql
SELECT local_date, SUM(value) AS value
FROM events
WHERE type = ? AND local_date BETWEEN ? AND ?
GROUP BY local_date
ORDER BY local_date;
```

**Done when:** `curl localhost:8080/health` answers and the process shuts down
cleanly on Ctrl-C.

**Watch out:** gaps. Days with no events do not appear in the result. The
heatmap renderer needs a value for every cell, so fill the gaps — either in
`DailyTotals` or in the renderer, but decide once and write it down.

---

## Phase 7 — Heatmap in the browser

Time: ~4 h · ~300 LOC

**Stack decision: htmx plus server-rendered SVG.** No JSON API, no build step,
no npm. Handlers return HTML fragments; htmx swaps them into the page. A single
`<script>` tag from a CDN, or vendored into `web/`.

- [ ] `internal/api/heatmap.go` — render an SVG grid from `[]DailyTotal`
- [ ] `GET /` — full page: type selector plus the heatmap for the default type
- [ ] `GET /heatmap?type=commit&year=2026` — fragment only, for htmx swaps
- [ ] `html/template` for the page shell; the SVG can be built with a strings
      builder or a template, whichever reads better
- [ ] Type selector populated from `ListTypes` — new collectors appear by
      themselves, no code change

Grid layout: 53 columns × 7 rows. Column is the week index, row is the weekday.
Each day is a `<rect>` with `x = week*13`, `y = weekday*13`, size 11.

Colour scale — five buckets, Kanagawa palette to match the rest of your setup:

| Count | Meaning |
| --- | --- |
| 0 | empty cell |
| 1–3 | lightest |
| 4–8 | |
| 9–15 | |
| 16+ | darkest |

**Done when:** you open `localhost:8080` and see your own commit history as a
heatmap, and switching the type in the selector swaps the grid without a page
reload.

**Watch out:** the first cell of the year is not necessarily a Sunday. Start the
grid at the Sunday on or before 1 January and leave the leading cells empty, the
same way GitHub does.

**This is the milestone that matters.** Everything before it is plumbing. Use it
for two weeks before continuing.

---

## Phase 8 — Second collector: Obsidian

Time: ~5 h · ~250 LOC

No API, no OAuth, no rate limits. Pure local file parsing — and the real test of
whether the dashboard is actually generic.

- [ ] `internal/obsidian/scan.go` — walk `Cockpit/`, read markdown frontmatter
- [ ] `external_id` = relative file path + modification time
- [ ] `internal/obsidian/normalize.go` — frontmatter → events
      (`book_finished`, `workout`, `todo_done`)
- [ ] Skip files without relevant frontmatter rather than erroring
- [ ] Extend `-sync` and `-normalize` to take a source name

**Done when:** Obsidian events sit next to GitHub events in the same table and
the type selector shows both without any dashboard code changing.

**Watch out:** Syncthing conflict files (`*.sync-conflict-*.md`). Filter them out
or you will double-count.

---

## Phase 9 — Extract the abstraction

Time: ~3 h · ~200 LOC

Only now. You have two implementations and can see what they actually share.
The `collector` package already exists but currently holds only the shared
`RawItem` and `Event` types — this phase adds the interface.

- [ ] `internal/collector/collector.go` — the `Collector` interface
- [ ] `internal/sync/runner.go` — iterate registered collectors, record
      `sync_state`, continue on individual failures
- [ ] Registry: a slice built in `main.go`
- [ ] Add `robfig/cron/v3` on a six-hour schedule alongside the HTTP server
- [ ] Keep manual flags for debugging: `-sync-once=github`

```go
type Collector interface {
    Name() string
    Fetch(ctx context.Context, since time.Time) ([]RawItem, error)
    Normalize(raw RawItem) ([]Event, error)
}
```

**Watch out:** one collector failing must not abort the others. Log the error
into `sync_state.last_error` and surface it in `/health`.

---

## Phase 9.5 — More GitHub queries

Time: ~4 h · ~250 LOC

Cheap once the pattern exists: each is one constant plus one function.

- [ ] `contributionYears` — which years have data, so the first sync can
      backfill everything instead of only the current year
- [ ] `commitContributionsByRepository` — per-repo daily breakdown, giving
      context to the calendar numbers
- [ ] Repository list with `pushedAt` and `defaultBranchRef.target.history.totalCount`
      — commits per project and when each was last touched

**Watch out:** `history.totalCount` counts every commit in the default branch,
not only yours. Identical for solo repos, wrong for shared ones. Filter with
`history(author: {id: ...})` using your own `viewer { id }` if it matters.

---

## Phase 10 — OAuth collector: Strava

Time: ~10 h · ~400 LOC

- [ ] Register an application at https://developers.strava.com
- [ ] `golang.org/x/oauth2` config, scope `activity:read_all`
- [ ] One-time authorization flow: temporary local callback handler
- [ ] Persist the refresh token — encrypted, not plaintext in the database
- [ ] Wrap the `TokenSource` so refreshed tokens get written back
- [ ] Events: `workout` with `value = duration`, distance and heart rate in `meta`

**Done when:** the token survives a restart and refreshes on its own.

---

## Phase 11 — Apple Health

Time: ~12 h · ~400 LOC

- [ ] Export from the Health app → several hundred MB of XML
- [ ] Parse with `encoding/xml` `Decoder.Token()` streaming — never `Unmarshal`
- [ ] Filter to record types you care about: sleep analysis, step count
- [ ] Merge adjacent sleep segments into one night before creating an event
- [ ] iOS Shortcut to automate the export into a synced folder

**Done when:** a night of sleep produces exactly one event, dated to the
wake-up day.

**Watch out:** the iterator pattern here is the same one you already used for
`sql.Rows` — `Next()`, read, `Err()` at the end.

---

## Repository layout

Single Go module, monorepo. Go lives at the repo root — no `backend/` directory.

```
lifelog/
├── cmd/lifelog/main.go
├── internal/
│   ├── api/
│   │   ├── server.go
│   │   ├── handlers.go
│   │   └── heatmap.go
│   ├── collector/collector.go     # shared RawItem, Event, later the interface
│   ├── config/config.go
│   ├── github/
│   │   ├── client.go              # how to talk to GraphQL
│   │   ├── fetch.go               # what to ask for
│   │   └── normalize.go           # payload → events
│   ├── obsidian/
│   ├── strava/
│   ├── store/
│   │   ├── migrations/
│   │   ├── db.go
│   │   ├── migrate.go
│   │   ├── raw.go
│   │   └── events.go
│   └── sync/runner.go
├── docs/
│   ├── decisions/
│   └── samples/
├── web/                           # htmx, any static assets
├── .env.example
└── go.mod
```

Dependency direction: `github` and `store` both point at `collector`, never at
each other. One package per source, all satisfying the same interface.

---

## Dependencies

Current:

```
modernc.org/sqlite
github.com/pressly/goose/v3
github.com/joho/godotenv
```

Later:

```
github.com/robfig/cron/v3      # phase 9
golang.org/x/oauth2            # phase 10
```

Deliberately absent: **HTTP framework** (stdlib `net/http` is enough), ORM,
PostgreSQL, Redis, message queue, auth framework, npm toolchain.

---

## Decisions worth writing down

Short files in `docs/decisions/`, three sentences each.

- [ ] Why raw payloads are stored separately from events
- [ ] Why zero-count days are not stored
- [ ] How `local_date` is derived, per event type
- [ ] What `value` and `unit` mean for each event type
- [ ] Why SQLite over PostgreSQL
- [ ] Why stdlib `net/http` over Gin or Fiber
- [ ] How secrets are stored

The fourth one is the one you will thank yourself for. By the fourth collector
you will not remember whether a workout's `value` was minutes or seconds.

---

## Recurring gotchas

| Problem | Where it bites | Fix |
| --- | --- | --- |
| Missing JSON tag | Any struct crossing the wire | Tag every field; a wrong name fails silently |
| Discarded return value | Everywhere | `go vet ./...` catches most of it |
| Day boundaries | Sleep, late-night commits | One documented rule, applied everywhere |
| Timezone drift | Travel, DST | Store UTC plus an explicit `local_date` |
| Duplicate events on re-sync | Every collector | `UNIQUE (source, external_id)` plus upsert |
| Rate limits | GitHub, Strava | Sync every six hours, not every five minutes |
| Token expiry | Fine-grained PAT, OAuth | Surface `last_error` in `/health` |
| Large XML | Apple Health | Streaming decoder, never `Unmarshal` |

---

## Security

This database is a precise profile of your life.

- [ ] Bind the HTTP server to `127.0.0.1`, never `0.0.0.0`
- [ ] Remote access via Tailscale or WireGuard only, no port forwarding
- [ ] Secrets in `.env`, never committed — verify with `git log -p -- .env`
- [ ] Encrypt OAuth refresh tokens at rest
- [ ] Backup: `sqlite3 data.db ".backup backup.db"` on a schedule, or Litestream

---

## Milestone summary

| Phase | Deliverable | Hours | Status |
| --- | --- | --- | --- |
| 0–5 | GitHub commits in a normalized events table | ~12 | done |
| 6–7 | Heatmap in the browser | ~7 | next |
| 8 | Obsidian data alongside it | 5 | |
| 9 | Pluggable collectors, scheduled sync | 3 | |
| 9.5 | Repo breakdown and multi-year backfill | 4 | |
| 10 | Strava with working OAuth | 10 | |
| 11 | Sleep and steps from Apple Health | 12 | |

Commit after every phase. This project's payoff arrives in years, so visible
progress is what keeps it alive — and it ends up tracking exactly those commits.