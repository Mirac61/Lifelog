-- +goose Up
-- local_date has always meant "start of the period"; nothing recorded how long
-- that period is. Daily events and per-repo yearly totals were indistinguishable,
-- so the heatmap drew a whole year of repo commits on Jan 1.
ALTER TABLE events ADD COLUMN granularity TEXT NOT NULL DEFAULT 'day'
    CHECK (granularity IN ('day', 'year'));

UPDATE events SET granularity = 'year' WHERE type = 'repo_commit';

CREATE INDEX idx_events_granularity_type ON events (granularity, type);

-- +goose Down
DROP INDEX idx_events_granularity_type;
ALTER TABLE events DROP COLUMN granularity;
