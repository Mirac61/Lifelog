-- +goose Up
ALTER TABLE todos ADD COLUMN estimate INTEGER;

-- +goose Down
ALTER TABLE todos DROP COLUMN estimate;
