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

**Next milestone: native Todos.** Decided 2026-08-23: the Obsidian collector is
shelved — disliking the tool plus the file-scanning/frontmatter complexity it
would need, when Lifelog can just own the data itself. Todos becomes the first
domain where Lifelog is the source of truth, not a mirror of an external
service. See Phase 8 below.

---

## Deferred design changes (decided, not started)

**1. Remove the heatmap from the dashboard — replace with a rolling daily digest.**
Agreed 2026-08-22, still parked — do this when the Obsidian slice is done and
the abstraction (Phase 9) exists. The dashboard and git page currently render
the same year-grid heatmap redundantly. The dashboard is the cross-domain
"what is going on?" view, so the year grid belongs to git.

- Dashboard `/`: drop the year grid, year stats and the year/type selector.
- Show a rolling window instead (today / 7 days / 30 days).
- One aggregated number across **all** event types, plus a per-domain split.
- Below it a day-by-day list (like `/day`, but for the window).
- Optional small bar/line sparkline for the window in place of the 53-column grid.

This is a pure presentation change over existing event data. No new
collectors, no schema change needed.

**2. ~~Add a type filter to the git page~~ — done 2026-08-23.** Git page now
has a Contributions/Pull Requests pill toggle plus a year stepper (`‹ 2026 ›`)
next to the page title, state kept in sync via an htmx OOB swap
(`gitcontrols.html`).

**3. Git page polish — flagged 2026-08-23 in a design pass over the pill/stepper
work, not started.**

- Heatmap renders the full year through Dec 31 even for the current year, so
  future days (which haven't happened yet) look visually identical to past
  days with zero commits — no way to tell "no activity" from "hasn't
  happened". It also leaves ~40% of the chart empty and dead on the right.
  Fix: cap the current year's grid at today instead of Dec 31 (past years
  render in full); shrink the SVG width to match so the empty space goes
  away too.
- PR status dots (`.pr-dot[data-state]`) render the same flat grey for every
  PR regardless of merged/open/closed. `state` is already fetched from
  GitHub (`fetch.go`) and decoded (`prPullRequest.State` in `normalize.go`),
  but `NormalizePRContributions` never puts it into the event's `meta` JSON,
  so it never reaches the template. **No migration needed** — `state` is
  already sitting in the cached `raw_payloads`. Fix: add
  `"state": node.PullRequest.State` to the meta map in
  `NormalizePRContributions`, then `-normalize` re-derives every PR event
  from payloads already on disk, no re-fetch required.
- Once state actually reaches the dot, bump `.pr-dot` from 7px to ~9px —
  at the current size the colour split is unreadable anyway.
- Heatmap legend reads "less" / "more" in lowercase — the only text on the
  page not in Sentence/Title Case. Align the casing, or drop the legend text
  since the colour ramp is fairly self-explanatory without it.

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

## Phase 8 — Native Todos — **next**

Time: ~3 h · ~150 LOC

Decided 2026-08-23, replaces the Obsidian collector. Unlike every other domain
so far, Todos is **not** fed by a collector — there is no external source, no
`raw_payloads`, no `-sync`/`-normalize`. Lifelog is the source of truth and the
data is mutable (a todo gets checked off, not appended-to like an event).
That is a deliberate first: everything up to here has been read-only
aggregation, this is the first thing you create *in* Lifelog.

- [ ] New table `todos` (own migration, not the `events` table — it doesn't
      fit the append-only event shape): `id, text, due_date NULL, done INTEGER,
      created_at, completed_at NULL`
- [ ] `internal/store/todos.go` — `ListTodos`, `CreateTodo`, `ToggleTodo`
- [ ] `GET /todos` — the nav link already exists in `layout.html`
      (`.nav-item` for Todos), it just 404s today; only the handler is missing
- [ ] Page: add-todo form (text + optional due date), list sorted by due date
      (overdue first, undated last), htmx toggle on click like the day-detail
      pattern
- [ ] Dashboard `/`: short "open todos" list, ties into the rolling-digest
      rework already parked above (deferred design change #1)
- [ ] `--accent: #9CCB86` for `body.domain-todos` is already in `style.css`
      from the original 4-domain palette — nothing to add there

**Done when:** a todo can be added with a due date, shown on `/todos` and on
the dashboard, and checked off without a page reload.

---

## Phase 8b — Notes / Braindump — **later, only if actually needed**

Parked 2026-08-23. Free-text notes for projects and braindumping, e.g. for
planning docs. Not started, not committed to — build it only if the "just a
todo list" itch turns out to be not enough on its own.

- Markdown stored as text in SQLite (`notes(id, title, body, tags,
  created_at, updated_at)`), **not** files on disk — files-on-disk is the
  Obsidian complexity this is meant to avoid.
- Render server-side with `goldmark` (pure Go, no npm/build step, fits the
  existing htmx/server-rendered-SVG pattern) — no client-side markdown lib.
- Search: SQLite `FTS5` over title + body is enough; no need for anything
  fancier at this scale.
- Tags: simple, e.g. a comma-separated column or a join table — not a
  full taxonomy.
- Explicitly **not** doing backlinks (`[[note]]` cross-references + a link
  graph/index) — real complexity (parse wiki-links, maintain the index, build
  UI for it) for a feature that only pays off with a much bigger note corpus
  than "projects + braindump" implies.
- If this ships, todos-with-a-due-date could optionally also be authored
  inline in a note via a convention like `- [ ] text @2026-08-25`, parsed out
  into the `todos` table from Phase 8 on save. That is an alternate entry
  path layered on top of Phase 8, not a prerequisite for it — Phase 8's
  `todos` table and UI must work standalone first.

---

## Phase 9 — Extract the abstraction

Time: ~3 h · ~200 LOC

Only after a **second real collector** exists (Strava below, since Obsidian
is shelved — Todos in Phase 8 is native, not a collector, so it doesn't count
here). Two implementations, so you can see what they actually share.

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

## Phase 12 — Gym tracking via openGym — later (extended)

Time: ~4 h · ~200 LOC · blocked on running an openGym instance

**Decision made (2026-08-23):** log workouts in
[openGym](https://gitea.com/DuarteSantos/openGym) (self-hosted, AGPL, free)
instead of Hevy (subscription, API paywalled) or LiftLog (local-only, would mean
manual exports every time). openGym covers gym stats only — heart rate, sleep
and steps stay with Apple Health (Phase 11), Strava stays Phase 10.

Setup outside this repo (one-time, ~30 min):

- [ ] Deploy openGym: VPS (~3–4 €/mo Hetzner/Netcup) or home box + Tailscale Serve
      (`tailscale serve` gives valid TLS on `*.ts.net`, which passkeys require)
- [ ] iPhone: install the PWA from Safari ("Add to Home Screen"), create profile
      with Face ID passkey
- [ ] Pin the deployed image/version — the project is days old and solo-maintained

Collector work in lifelog:

- [ ] `internal/gym/read.go` — read `state-<uid>.json` straight from the openGym
      `./data/` directory (shared volume/SSH/Tailscale file access). Do **not**
      use `GET /api/data`: auth is passkey session cookies, not usable headless.
      Same shape as `raw_payloads`: store the whole JSON, normalize separately.
- [ ] `external_id` = workout date + start time (the JSON has no stable id field)
- [ ] `internal/gym/normalize.go` — one `workout` event per entry: `value` =
      session duration in minutes (same unit rule as Strava, Phase 10),
      exercises/sets/volume into `meta`
- [ ] Bodyweight entries become their own event type (`bodyweight`), not folded
      into workouts
- [ ] `-sync-once=gym` works like the other sources once Phase 9 exists

What the source gives us (`state-<uid>.json`): `workouts[]` (date, duration,
exercises, sets as weight × reps, warm-up flags, optional RIR/RPE), `routines[]`,
`bodyweight[]` (date + kg), week schedule. No heart rate/sleep/steps — by design.

**Done when:** finishing a session in the openGym PWA puts a `workout` event on
the dashboard heatmap without any manual step.

**Watch out:** the state schema will drift while the project is young. Normalize
defensively — ignore unknown fields, never hard-fail on a new shape — and keep
openGym's one-tap JSON export as the documented escape hatch.

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
│   ├── gym/                   # (Phase 12, reads openGym state JSON)
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
| 12 | Gym workouts from self-hosted openGym | 4 | |

Commit after every phase. This project's payoff arrives in years, so visible
progress is what keeps it alive — and it ends up tracking exactly those commits.
