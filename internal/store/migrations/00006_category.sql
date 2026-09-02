-- +goose Up
ALTER TABLE todos ADD COLUMN category TEXT;

-- +goose Down
ALTER TABLE todos DROP COLUMN category; 
