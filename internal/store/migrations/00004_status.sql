-- +goose Up
DROP INDEX idx_todos_done_due;

ALTER TABLE todos ADD COLUMN status TEXT NOT NULL DEFAULT 'todo'
    CHECK (status IN ('todo', 'in_progress', 'done'));

UPDATE todos SET status = 'done' WHERE done = 1;

ALTER TABLE todos DROP COLUMN done;

CREATE INDEX idx_todos_status_due ON todos (status, due_date);

-- +goose Down
DROP INDEX idx_todos_status_due;

ALTER TABLE todos ADD COLUMN done INTEGER NOT NULL DEFAULT 0
    CHECK (done IN (0, 1));

UPDATE todos SET done = 1 WHERE status = 'done';

ALTER TABLE todos DROP COLUMN status;

CREATE INDEX idx_todos_done_due ON todos (done, due_date);
