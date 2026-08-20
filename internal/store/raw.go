package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

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
