-- +goose Up
CREATE TABLE todos (
    id           INTEGER PRIMARY KEY,
    text         TEXT NOT NULL,
    due_date     TEXT,
    done         INTEGER NOT NULL DEFAULT 0 CHECK (done IN (0, 1)),
    created_at   TEXT NOT NULL,
    completed_at TEXT
);

CREATE INDEX idx_todos_done_due ON todos (done, due_date);

-- +goose Down
DROP INDEX idx_todos_done_due;
DROP TABLE todos;
