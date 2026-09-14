# Lifelog

A personal data warehouse in Go. One binary, SQLite, collectors pulling
external sources into a shared event store.

## Why

The question this is built around: **how did my day go, and what could I do
better?**

Raw numbers don't answer that. "12 commits" is neither good nor bad as long as
no yardstick stands next to it. Every source knows *what* happened. None of
them knows *what kind of day it was*.

So the project draws a hard line:

> Lifelog does not produce data a machine could have collected. It only asks
> for what no sensor measures.

That rules out task management. Tasks live in Reminders and Obsidian, and do
so better. It lets exactly two things in: **context** (what kind of day was
this) and **rating** (how was it, how productive was it). Both are typed, never
computed.

## What works today

One source is wired up end to end: GitHub.

- Contribution heatmap, rendered server-side as SVG — no charting library
- Pull request events and per-repo yearly commit totals
- Per-day detail view
- Backfill across every year GitHub has activity for

That's it. "Data warehouse" describes the target shape, not the current state.
The rating layer that makes the numbers mean something is the next milestone
but one — see [the roadmap](docs/Roadmap.md).

## How it works

```
GitHub GraphQL  ──fetch──▶  raw_payloads  ──normalize──▶  events  ──▶  UI
```

**Raw payloads are kept.** Fetching and interpreting are separate steps.
`-normalize` rebuilds the entire `events` table from stored payloads with no
network access, which makes schema changes cheap. `InsertEvents` deletes
`WHERE raw_id = ?` and re-inserts inside one transaction, so re-normalizing
never duplicates.

**`granularity` carries the period length.** `local_date` has always meant
"start of the period" — but nothing recorded how long that period is. Daily
events and per-repo yearly totals were indistinguishable, and the heatmap drew
a whole year of repo commits on January 1st. One column fixed it:

```sql
CREATE TABLE events (
    id          INTEGER PRIMARY KEY,
    source      TEXT NOT NULL,
    type        TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    local_date  TEXT NOT NULL,   -- start of the period
    value       REAL,
    unit        TEXT,
    meta        TEXT,            -- JSON, source-specific
    raw_id      INTEGER REFERENCES raw_payloads (id) ON DELETE CASCADE,
    granularity TEXT NOT NULL DEFAULT 'day'
                CHECK (granularity IN ('day', 'year'))
);
```

| type | granularity | unit | value |
|---|---|---|---|
| `contributions` | `day` | `count` | contributions that day |
| `pull_request` | `day` | `pr` | `1` |
| `repo_commit` | `year` | `commits` | commits in that repo that year |

**No framework.** Routing is the standard library's `http.ServeMux` with Go
1.22 method patterns. htmx is vendored rather than pulled from npm. CSS, fonts,
htmx and the migrations are all `go:embed`-ed, so the binary is self-contained
and `CGO_ENABLED=0` builds a static one.

Four direct dependencies:

- [templ](https://templ.guide) for typed HTML
- [goose](https://github.com/pressly/goose) for migrations
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) for a pure-Go driver
- godotenv for loading `.env`

Deliberately absent:

- HTTP framework
- ORM
- PostgreSQL
- Redis
- message queueing
- auth framework
- npm

## Running it

You need Go 1.26+ and a GitHub token with default read scope.

```sh
cp .env.example .env      # put your token in it
go tool templ generate
go run ./cmd/lifelog      # http://127.0.0.1:8080
```

`go tool templ generate` is not optional. templ is pinned as a
[tool dependency](https://go.dev/doc/modules/managing-dependencies#tools) in
`go.mod`, so there is nothing to install globally — but the generated
`*_templ.go` files checked into the repo may lag behind the `.templ` sources.
The Dockerfile regenerates them at build time for the same reason.

### Docker

```sh
docker compose up -d --build
```

Publishes to `127.0.0.1:8080` only, never `0.0.0.0`. The image has a
healthcheck on `/health`. One-off runs go through the entrypoint:

```sh
docker compose run lifelog -backfill
```

### Flags

Every flag does its work and exits. With no flag, the HTTP server starts.

| Flag | Effect |
|---|---|
| `-sync -year N` | fetch one year from GitHub into `raw_payloads`. Does *not* normalize. |
| `-normalize` | rebuild `events` from stored payloads. No network. |
| `-backfill` | sync every year with activity, then normalize once. |
| *(none)* | serve HTTP, plus a background sync of the current year on startup. |

## Configuration

Read from the environment, or from a `.env` file in the working directory.

| Variable | Required | Default |
|---|---|---|
| `GITHUB_TOKEN` | **yes** | — |
| `DATABASE_PATH` | no | `data.db` |
| `LISTEN_ADDRESS` | no | `127.0.0.1:8080` |

Two things worth knowing:

- `GITHUB_TOKEN` is required **even to just serve the UI**. Startup calls
  `config.Load()` before anything else and exits non-zero without it.
- Every GraphQL query goes through `viewer`, so the token owner is always who
  gets synced. There is no username to configure.

Migrations run automatically on every start. There is no goose binary to
install and no migrate step to remember.

## Routes

| Route | |
|---|---|
| `GET /` | dashboard |
| `GET /day?date=YYYY-MM-DD` | one day in detail |
| `GET /git?year=&type=` | contribution heatmap and repo list |
| `GET /git/view` | htmx fragment for switching year or type |
| `GET /health` | pings the database; used by the container healthcheck |
| `GET /static/` | embedded CSS, fonts, htmx |

## Development

```sh
# watch templates, rebuild, live-reload through a proxy
go tool templ generate -watch -cmd 'go run ./cmd/lifelog' \
  -proxy 'http://localhost:8080' -open-browser=false

# what CI would run, if there were CI
go tool templ generate && go build ./... && go vet ./... && go test ./...

# static binary
CGO_ENABLED=0 go build -ldflags='-s -w' -o bin/lifelog ./cmd/lifelog
```

The same commands are wired up as tasks in `.zed/tasks.json`.

## Scope

Some things are ruled out on purpose, not merely unbuilt:

- **Task management** — Reminders and Obsidian already do it better.
- **A computed daily score** — the weights would be guesswork, and you end up
  optimizing for them. 12 commits on 4 hours of sleep would score well.
- **Multi-user, login, cloud** — "rating: 7" means nothing between two people.
  No network effect, just attack surface on a profile of a life.

The [roadmap](docs/Roadmap.md) has the full list with the reasoning, plus the
milestones. M2 (removing todos) is done; next up is M5 (Reminders as a source),
then M1 (the rating layer) — the order is deliberately not numeric.

## Security

This database is a precise profile of a life. It is treated that way:

- Binds to `127.0.0.1` by default, in the container too.
- Remote access belongs on Tailscale or WireGuard — not a port forward.
- Secrets live in `.env`, never to be committed.

## Layout

```
cmd/lifelog/       entrypoint, flags, sync orchestration
internal/api/      HTTP handlers, SVG heatmap rendering
internal/store/    SQLite access, embedded goose migrations
internal/github/   GraphQL client, fetch and normalize
internal/collector/  shared event types
internal/config/   environment configuration
web/templ/         templ components
web/static/        CSS, fonts, htmx — embedded into the binary
docs/Roadmap.md    where this is going, and why
```

`DesignPrototype/` is generated reference output from a design tool. It is
marked `linguist-vendored` and is **not part of the build** — only `web/` gets
embedded.

The canonical repository is on
[Codeberg](https://codeberg.org/Mirac61/Lifelog); GitHub is a push mirror.
Issues belong on Codeberg.

---

This README is in English. The user interface and `docs/Roadmap.md` are in
German, because this is a tool its author uses in German.
