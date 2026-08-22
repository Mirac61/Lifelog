# lifelog — Build Roadmap

Personal data warehouse in Go. One binary, SQLite, collectors pull from external
sources into a unified event store.

**Rule for this project:** build the thinnest possible vertical slice first
(Phase 1–7). Abstract only after the second collector exists.

---

## Current state (August 2026)

The GitHub vertical slice is shipped end to end: one binary serves a browser UI
(dashboard + git pages) over htmx/server-rendered SVG from a single SQLite event
store. The event store is honest about granularity — daily events and yearly
aggregates carry different `granularity`, and only day events feed the day axis.

**Next milestone: Obsidian.** It is the second collector, so it is the real test
of whether the store and dashboard are actually generic — and it unlocks the
Todos and Calendar domains that already have their accent tokens in the CSS.

---

## Deferred design changes (decided, not started)

Two UI decisions were agreed on 2026-08-22 but deliberately parked. Do these
when the Obsidian slice is done and the abstraction (Phase 9) exists.

**1. Remove the heatmap from the dashboard — replace with a rolling daily digest.**
The dashboard and git page currently render the same year-grid heatmap
redundantly. The dashboard is the cross-domain "what is going on?" view, so the
year grid belongs to git.

- Dashboard `/`: drop the year grid, year stats and the year/type selector.
- Show a rolling window instead (today / 7 days / 30 days).
- One aggregated number across **all** event types, plus a per-domain split.
- Below it a day-by-day list (like `/day`, but for the window).
- Optional small bar/line sparkline for the window in place of the 53-column grid.

**2. Add a type filter to the git page (filter by PR).**
Once the heatmap lives only on git, give the git page the type selector the
dashboard used to have — `Contributions` / `Pull requests` (maybe `Commits`).
Hero stats + heatmap follow the selected type; the PR list stays.

Both are pure presentation changes over the existing event data. No new
collectors, no schema change needed.

---

## Phase 0 — Decide and prepare — **done**

- [x] Pick the final project name (`lifelog`)
- [x] Create GitHub repo, private for now
- [x] Add `.gitignore` (Go template + `*.db`, `*.db-wal`, `*.db-shm`, `.env`)
- [x] Skip LICENSE while the repo is private
- [x] Create a **fine-grained** Personal Access Token — resource owner: self,
      repository access: all repositories, permission: Metadata read-only

**Watch out:** the token expires. Put the date in a calendar.

---

## Phase 1 — Skeleton — **done**

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

## Phase 2 — Schema and migrations — **done**

- [x] Add `github.com/pressly/goose/v3`
- [x] `internal/store/migrations/00001_init.sql`
- [x] Embed the migrations directory with `//go:embed`
- [x] Run migrations on startup, before anything else touches the database

Three tables: `raw_payloads`, `events`, `sync_state`.

---

## Phase 3 — Explore the GitHub API by hand — **done**

- [x] Run the contributions query and inspect the response
- [x] Decide the natural unique key: `contributions:<year>`
- [ ] Save a sample response to `docs/samples/github-contributions.json`

**Decision made:** use `contributionCalendar` rather than
`commitContributionsByRepository`. The calendar returns a flat list of
`{date, contributionCount}` that maps almost directly onto an events row. The
per-repository breakdown became a **separate query** — a second `external_id`,
not a replacement.

**Watch out:** `contributionsCollection` accepts a maximum span of one year.
Query one year at a time, or use GraphQL aliases to request several years in a
single request.

---

## Phase 4 — Fetch into raw_payloads — **done**

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

## Phase 5 — Normalize into events — **done**

- [x] `internal/github/normalize.go` — payload → `[]collector.Event`
- [x] `internal/store/raw.go` — `ListRaw` reads payloads back out
- [x] `internal/store/events.go` — `InsertEvents`, delete-then-insert per
      `raw_id`, wrapped in a transaction
- [x] Flag `-normalize` that reprocesses every raw payload for a source

**Decision made:** days with a count of zero are skipped. The absence of a row
means no activity.

**Decision made:** three event types. `contributions` (daily, from the calendar;
this is total contributions, not commits), `pull_request` (daily), and
`repo_commit` (yearly per-repo commit total, from `commitContributionsByRepository`).

---

## Phase 6 — Query layer and HTTP server — **done**

- [x] `internal/store/events.go` — `DailyTotals`, `DailyStats`, `ListTypes`,
      `EventsForDay`, `ListRepoCommits`, `ListPullRequests`
- [x] `internal/api/server.go` — `http.ServeMux`, bound to `cfg.ListenAddress`
- [x] `GET /health` — database ping (surfacing `sync_state` is still open)
- [x] Graceful shutdown: `srv.Shutdown(ctx)` driven by the existing SIGINT context
- [x] Remove the `-sync`/`-normalize` early `return` path so the server actually
      starts when no flag is given

---

## Phase 7 — Heatmap in the browser — **done**

**Stack decision: htmx plus server-rendered SVG.** No JSON API, no build step,
no npm. Handlers return HTML fragments; htmx swaps them into the page.

- [x] `internal/api/heatmap.go` — render an SVG grid from `[]DailyTotal`
- [x] `GET /` — full page: type selector plus the heatmap for the default type
- [x] `GET /heatmap?type=...&year=...` — fragment only, for htmx swaps
- [x] `html/template` for the page shell; SVG built in Go
- [x] Type selector populated from `ListTypes` — new collectors appear by
      themselves, no code change

Grid layout: 53 columns × 7 rows. Column is the week index, row is the weekday.
Each day is a `<rect>`; the grid starts at the Sunday on or before 1 January.

**Note / deviation from the plan:** the ramp thresholds (5 levels) live in
`levelFor` in Go; the colours are mixed in oklab from `--accent` in CSS, per
domain (blue for git, achromatic paper for the cross-domain dashboard).

**Shipped beyond the original plan** (same slice, no new collectors):

- Git page (`/git`): hero stats, contributions heatmap, PR list, per-repo
  rankings, year selector
- `/day` in-page detail via htmx — cells on the dashboard are clickable;
  git-page cells are inert because the partial renders without `hx-get`
- `granularity` column (`day`/`year`) + migration, so yearly `repo_commit`
  totals never land on the day axis
- `-backfill` via `contributionYears`, plus `FetchRepoCommits` and
  `FetchPRContributions`

---

## Phase 8 — Second collector: Obsidian — **next**

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

Only after Obsidian. You now have two implementations and can see what they
actually share.

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
into `sync_state.last_error` and surface it in `/health`. This also closes the
two open items from Phase 4 and 6.

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

---

## Repository layout

Single Go module, monorepo. Go lives at the repo root — no `backend/` directory.

```
lifelog/
├── cmd/lifelog/main.go        # flags, startup sync/normalize, HTTP server
├── cmd/lifelog/sync.go        # syncYear + normalizeAll (github)
├── internal/
│   ├── api/                   # server.go, index.go, heatmap.go, git.go, day.go
│   ├── collector/collector.go # shared RawItem, Event (interface comes in Phase 9)
│   ├── config/config.go
│   ├── github/                # client.go, fetch.go, normalize.go (+ tests)
│   ├── obsidian/              # (Phase 8)
│   ├── strava/                # (Phase 10)
│   ├── store/                 # db.go, migrate.go, raw.go, events.go, migrations/
│   └── sync/runner.go         # (Phase 9)
├── docs/
│   ├── Roadmap.md
│   └── samples/               # (open)
├── web/
│   ├── templates/             # layout, pages, partials
│   └── static/                # style.css, fonts
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
| Missing JSON tag | Any struct crossing the wire | Tag every field |
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
| 0–5 | GitHub data in a normalized events table | ~12 | done |
| 6 | Query layer + HTTP server | ~3 | done |
| 7 | Heatmap + dashboard, htmx | ~4 | done |
| 7.5 | Git page, `/day`, granularity, backfill (unplanned) | ~4 | done |
| 8 | Obsidian data alongside it | 5 | next |
| 9 | Pluggable collectors, scheduled sync | 3 | |
| 10 | Strava with working OAuth | 10 | |
| 11 | Sleep and steps from Apple Health | 12 | |

Commit after every phase. This project's payoff arrives in years, so visible
progress is what keeps it alive — and it ends up tracking exactly those commits.
