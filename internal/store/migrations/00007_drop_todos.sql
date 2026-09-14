-- +goose Up
DROP TABLE todos;

-- +goose Down
-- +goose StatementBegin
SELECT RAISE(FAIL, 'todos are gone for good, see Roadmap M2');
-- +goose StatementEnd
