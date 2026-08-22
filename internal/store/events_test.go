package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

func TestDailyStats(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	// two streaks: 3 days, then 2 days after a gap
	days := []struct {
		date  string
		value float64
	}{
		{"2026-03-01", 1}, {"2026-03-02", 5}, {"2026-03-03", 2},
		{"2026-03-09", 4}, {"2026-03-10", 3},
	}
	for _, d := range days {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO events (source, type, occurred_at, local_date, value, unit) VALUES ('t','contributions',?,?,?,'count')`,
			d.date+"T12:00:00Z", d.date, d.value); err != nil {
			t.Fatal(err)
		}
	}

	got, err := DailyStats(ctx, db, "contributions", "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	want := Stats{Total: 15, ActiveDays: 5, LongestStreak: 3, BestDate: "2026-03-02", BestValue: 5}
	if got != want {
		t.Errorf("DailyStats = %+v, want %+v", got, want)
	}

	events, err := EventsForDay(ctx, db, "2026-03-02")
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != "contributions" || events[0].Value != 5 {
		t.Errorf("EventsForDay = %+v", events)
	}
}

func TestYearGranularityOffDayAxis(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	// A daily commit and a yearly repo total, both bucketed on Jan 1.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO events (source, type, occurred_at, local_date, value, unit, granularity) VALUES ('github','contributions',?,?,?,'count','day')`,
		"2026-01-01T12:00:00Z", "2026-01-01", 3); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO events (source, type, occurred_at, local_date, value, unit, granularity) VALUES ('github','repo_commit',?,?,?,'commits','year')`,
		"2026-01-01T12:00:00Z", "2026-01-01", 290); err != nil {
		t.Fatal(err)
	}

	totals, err := DailyTotals(ctx, db, "contributions", "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if len(totals) != 1 || totals[0].Value != 3 {
		t.Errorf("DailyTotals(commit) = %+v, want one day of 3", totals)
	}

	repoTotals, err := DailyTotals(ctx, db, "repo_commit", "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if len(repoTotals) != 0 {
		t.Errorf("DailyTotals(repo_commit) = %+v, want empty (year aggregate is not a day)", repoTotals)
	}

	stats, err := DailyStats(ctx, db, "repo_commit", "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatal(err)
	}
	if stats.ActiveDays != 0 || stats.Total != 0 {
		t.Errorf("DailyStats(repo_commit) = %+v, want zero", stats)
	}

	types, err := ListTypes(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	for _, ty := range types {
		if ty == "repo_commit" {
			t.Errorf("ListTypes exposes repo_commit on the day chart: %v", types)
		}
	}

	dayEvents, err := EventsForDay(ctx, db, "2026-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(dayEvents) != 1 || dayEvents[0].Type != "contributions" {
		t.Errorf("EventsForDay(2026-01-01) = %+v, want only the daily commit", dayEvents)
	}
}

func TestInsertEventsPersistsGranularity(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT INTO raw_payloads (source, external_id, fetched_at, payload) VALUES ('github','foo','x','{}')`); err != nil {
		t.Fatal(err)
	}

	yearly := collector.Event{
		Type:        "repo_commit",
		OccurredAt:  time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		LocalDate:   "2026-01-01",
		Granularity: "year",
		Value:       290,
		Unit:        "commits",
		Meta:        json.RawMessage(`{"name":"foo","nameWithOwner":"org/foo","pushedAt":"x"}`),
	}
	if err := InsertEvents(ctx, db, "github", 1, []collector.Event{yearly}); err != nil {
		t.Fatal(err)
	}

	var gran string
	if err := db.QueryRowContext(ctx, `SELECT granularity FROM events WHERE type='repo_commit'`).Scan(&gran); err != nil {
		t.Fatal(err)
	}
	if gran != "year" {
		t.Errorf("InsertEvents stored granularity=%q, want %q", gran, "year")
	}

	// Re-collecting must not overwrite it back to day.
	if err := InsertEvents(ctx, db, "github", 1, []collector.Event{yearly}); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT granularity FROM events WHERE type='repo_commit'`).Scan(&gran); err != nil {
		t.Fatal(err)
	}
	if gran != "year" {
		t.Errorf("re-collect stored granularity=%q, want %q", gran, "year")
	}

	commits, err := ListRepoCommits(ctx, db, 2026)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].NameWithOwner != "org/foo" || commits[0].Value != 290 {
		t.Errorf("ListRepoCommits = %+v, want the single yearly repo row", commits)
	}
}
