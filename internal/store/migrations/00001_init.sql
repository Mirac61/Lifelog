-- +goose Up
CREATE TABLE raw_payloads (
    id          INTEGER PRIMARY KEY,
    source      TEXT NOT NULL,
    external_id TEXT NOT NULL,
    fetched_at  TEXT NOT NULL,
    payload     TEXT NOT NULL,
    UNIQUE (source, external_id)
);

CREATE TABLE events (
    id          INTEGER PRIMARY KEY,
    source      TEXT NOT NULL,
    type        TEXT NOT NULL,
    occurred_at TEXT NOT NULL,
    local_date  TEXT NOT NULL,
    value       REAL,
    unit        TEXT,
    meta        TEXT,
    raw_id      INTEGER REFERENCES raw_payloads (id) ON DELETE CASCADE
);

CREATE INDEX idx_events_date_type ON events (local_date, type);
CREATE INDEX idx_events_raw ON events (raw_id);

CREATE TABLE sync_state (
    source       TEXT PRIMARY KEY,
    last_sync_at TEXT NOT NULL,
    last_error   TEXT
);

-- +goose Down
DROP TABLE sync_state;
DROP TABLE events;
DROP TABLE raw_payloads;
