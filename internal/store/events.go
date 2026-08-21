package store

import (
	"context"
	"database/sql"
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

func InsertEvents(ctx context.Context, db *sql.DB, source string, rawID int64, events []collector.Event) error {
	const delQuery = `DELETE FROM events WHERE raw_id = ?`
	const insertQuery = `INSERT INTO events (source, type, occurred_at, local_date, value, unit, meta, raw_id) VALUES (?,?,?,?,?,?,?,?)`
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, delQuery, rawID); err != nil {
		return fmt.Errorf("delete events: %w", err)
	}

	for _, e := range events {
		if _, err := tx.ExecContext(ctx, insertQuery, source, e.Type, e.OccurredAt.Format(time.RFC3339), e.LocalDate, e.Value, e.Unit, e.Meta, rawID); err != nil {
			return fmt.Errorf("insert events: %w", err)
		}
	}
	return tx.Commit()
}

func DailyTotals(ctx context.Context, db *sql.DB, eventType, from, to string) ([]DailyTotal, error) {
	const query = `SELECT local_date, SUM(value)
		FROM events
		WHERE type = ? AND local_date BETWEEN ? AND ?
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
	const query = `SELECT DISTINCT type FROM events ORDER BY type`

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

type Stats struct {
	Total         float64
	ActiveDays    int
	LongestStreak int
	BestDate      string
	BestValue     float64
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
		WHERE local_date = ?
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
