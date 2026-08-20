package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

type RawRow struct {
	ID         int64
	ExternalID string
	Payload    json.RawMessage
}

func InsertRaw(ctx context.Context, db *sql.DB, source string, item collector.RawItem) error {
	const query = `
INSERT INTO raw_payloads (source, external_id, fetched_at, payload)
VALUES (?, ?, ?, ?)
ON CONFLICT (source, external_id) DO UPDATE SET
    payload = excluded.payload,
    fetched_at = excluded.fetched_at`

	_, err := db.ExecContext(ctx, query, source, item.ExternalID,
		item.OccurredAt.Format(time.RFC3339), item.Payload)
	return err
}

func ListRaw(ctx context.Context, db *sql.DB, source string) ([]RawRow, error) {
	const query = `SELECT id, external_id, payload FROM raw_payloads WHERE source = ?`
	rows, err := db.QueryContext(ctx, query, source)
	if err != nil {
		return nil, fmt.Errorf("query raw payloads: %w", err)
	}
	defer rows.Close()
	var items []RawRow
	for rows.Next() {
		var row RawRow
		if err := rows.Scan(&row.ID, &row.ExternalID, &row.Payload); err != nil {
			return nil, fmt.Errorf("scan raw payload: %w", err)
		}
		items = append(items, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate raw payloads: %w", err)
	}
	return items, nil
}
