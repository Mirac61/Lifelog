package store

import (
	"context"
	"path/filepath"
	"testing"
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
			`INSERT INTO events (source, type, occurred_at, local_date, value, unit) VALUES ('t','commit',?,?,?,'count')`,
			d.date+"T12:00:00Z", d.date, d.value); err != nil {
			t.Fatal(err)
		}
	}

	got, err := DailyStats(ctx, db, "commit", "2026-01-01", "2026-12-31")
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
	if len(events) != 1 || events[0].Type != "commit" || events[0].Value != 5 {
		t.Errorf("EventsForDay = %+v", events)
	}
}
