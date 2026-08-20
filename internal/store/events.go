package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

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
