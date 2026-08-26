package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

type DailyTotal struct {
	LocalDate string
	Value     float64
}

type DayEvent struct {
	Type  string
	Value float64
	Unit  string
}

type RepoCommit struct {
	Name          string
	NameWithOwner string
	OccurredAt    string
	LocalDate     string
	Value         float64
}

type Stats struct {
	Total         float64
	ActiveDays    int
	LongestStreak int
	BestDate      string
	BestValue     float64
}

type PullRequest struct {
	OccurredAt string
	LocalDate  string
	Title      string `json:"title"`
	URL        string `json:"url"`
	Repo       string `json:"repository"`
	State      string `json:"state"`
}

func InsertEvents(ctx context.Context, db *sql.DB, source string, rawID int64, events []collector.Event) error {
	const delQuery = `DELETE FROM events WHERE raw_id = ?`
	const insertQuery = `INSERT INTO events (source, type, occurred_at, local_date, value, unit, meta, raw_id, granularity) VALUES (?,?,?,?,?,?,?,?,?)`
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, delQuery, rawID); err != nil {
		return fmt.Errorf("delete events: %w", err)
	}

	for _, e := range events {
		granularity := "day"
		if e.Granularity == "year" {
			granularity = "year"
		}
		if _, err := tx.ExecContext(ctx, insertQuery, source, e.Type, e.OccurredAt.Format(time.RFC3339), e.LocalDate, e.Value, e.Unit, e.Meta, rawID, granularity); err != nil {
			return fmt.Errorf("insert events: %w", err)
		}
	}
	return tx.Commit()
}

func DailyTotals(ctx context.Context, db *sql.DB, eventType, from, to string) ([]DailyTotal, error) {
	// repo_commit is yearly, not day-granularity, so it's excluded here.
	const query = `SELECT local_date, SUM(value)
		FROM events
		WHERE type = ? AND granularity = 'day' AND local_date BETWEEN ? AND ?
		GROUP BY local_date
		ORDER BY local_date`

	rows, err := db.QueryContext(ctx, query, eventType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var totals []DailyTotal
	for rows.Next() {
		var t DailyTotal
		if err := rows.Scan(&t.LocalDate, &t.Value); err != nil {
			return nil, fmt.Errorf("scan daily total: %w", err)
		}
		totals = append(totals, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily totals: %w", err)
	}
	return totals, nil
}

func ListTypes(ctx context.Context, db *sql.DB) ([]string, error) {
	const query = `SELECT DISTINCT type FROM events WHERE granularity = 'day' ORDER BY type`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var types []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("fetching types: %w", err)
		}
		types = append(types, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating types: %w", err)
	}
	return types, nil
}

func DailyStats(ctx context.Context, db *sql.DB, eventType, from, to string) (Stats, error) {
	totals, err := DailyTotals(ctx, db, eventType, from, to)
	if err != nil {
		return Stats{}, fmt.Errorf("daily stats: %w", err)
	}

	var s Stats
	var prev time.Time
	streak := 0
	for _, t := range totals {
		if t.Value <= 0 {
			continue
		}
		day, err := time.Parse("2006-01-02", t.LocalDate)
		if err != nil {
			return Stats{}, fmt.Errorf("parse local date %q: %w", t.LocalDate, err)
		}
		s.Total += t.Value
		s.ActiveDays++
		if t.Value > s.BestValue {
			s.BestValue, s.BestDate = t.Value, t.LocalDate
		}
		if !prev.IsZero() && day.Sub(prev) == 24*time.Hour {
			streak++
		} else {
			streak = 1
		}
		if streak > s.LongestStreak {
			s.LongestStreak = streak
		}
		prev = day
	}
	return s, nil
}

func EventsForDay(ctx context.Context, db *sql.DB, date string) ([]DayEvent, error) {
	const query = `SELECT type, COALESCE(value, 0), COALESCE(unit, '')
		FROM events
		WHERE local_date = ? AND granularity = 'day'
		ORDER BY occurred_at, type`

	rows, err := db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []DayEvent
	for rows.Next() {
		var e DayEvent
		if err := rows.Scan(&e.Type, &e.Value, &e.Unit); err != nil {
			return nil, fmt.Errorf("scan day event: %w", err)
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate day events: %w", err)
	}
	return events, nil
}

func ListRepoCommits(ctx context.Context, db *sql.DB, year int) ([]RepoCommit, error) {
	const query = `SELECT occurred_at, local_date, value, meta
		FROM events
		WHERE source = ?
		  AND type = ?
		  AND local_date BETWEEN ? AND ?
		ORDER BY value DESC`

	from := fmt.Sprintf("%d-01-01", year)
	to := fmt.Sprintf("%d-12-31", year)
	rows, err := db.QueryContext(ctx, query, "github", "repo_commit", from, to)
	if err != nil {
		return nil, fmt.Errorf("query repo commits: %w", err)
	}
	defer rows.Close()

	var commits []RepoCommit
	for rows.Next() {
		var commit RepoCommit
		var meta json.RawMessage
		if err := rows.Scan(&commit.OccurredAt, &commit.LocalDate, &commit.Value, &meta); err != nil {
			return nil, fmt.Errorf("scan repo commit: %w", err)
		}

		var details struct {
			Name          string `json:"name"`
			NameWithOwner string `json:"nameWithOwner"`
		}
		if err := json.Unmarshal(meta, &details); err != nil {
			return nil, fmt.Errorf("decode repo metadata: %w", err)
		}
		commit.Name = details.Name
		commit.NameWithOwner = details.NameWithOwner
		commits = append(commits, commit)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate repo commits: %w", err)
	}

	return commits, nil
}

// EventYearRange returns the earliest and latest year with a github event.
// Both are 0 when the source has no events yet.
func EventYearRange(ctx context.Context, db *sql.DB) (first, last int, err error) {
	const query = `SELECT MIN(substr(local_date, 1, 4)), MAX(substr(local_date, 1, 4))
		FROM events
		WHERE source = ?`

	var minYear, maxYear sql.NullInt64
	if err := db.QueryRowContext(ctx, query, "github").Scan(&minYear, &maxYear); err != nil {
		return 0, 0, fmt.Errorf("query event year range: %w", err)
	}
	if !minYear.Valid {
		return 0, 0, nil
	}
	return int(minYear.Int64), int(maxYear.Int64), nil
}

func ListPullRequests(ctx context.Context, db *sql.DB, year int) ([]PullRequest, error) {
	const query = `SELECT occurred_at, local_date, meta FROM events WHERE source = ? 
		AND type = ? 
		AND local_date BETWEEN ? AND ? 
		ORDER BY occurred_at DESC`
	from := fmt.Sprintf("%d-01-01", year)
	to := fmt.Sprintf("%d-12-31", year)
	rows, err := db.QueryContext(ctx, query, "github", "pull_request", from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var prs []PullRequest
	var meta json.RawMessage
	for rows.Next() {
		var pr PullRequest
		if err := rows.Scan(&pr.OccurredAt, &pr.LocalDate, &meta); err != nil {
			return nil, fmt.Errorf("scan pr contributions: %w", err)
		}
		if err := json.Unmarshal(meta, &pr); err != nil {
			return nil, fmt.Errorf("decode pr metadata: %w", err)
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pull requests: %w", err)
	}
	return prs, nil
}
