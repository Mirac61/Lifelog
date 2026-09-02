package store

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestListTodosForDay(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	// ein Todo für heute ohne estimate/category, eines überfällig und offen
	if _, err := db.ExecContext(ctx, `INSERT INTO todos (text, due_date, created_at) VALUES
		('heute ohne extras', '2026-08-27', '2026-08-27'),
		('überfällig', '2026-08-25', '2026-08-25')`); err != nil {
		t.Fatal(err)
	}

	todos, err := ListTodosForDay(ctx, db, "2026-08-27")
	if err != nil {
		t.Fatalf("ListTodosForDay: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("want 2 todos (heute + überfällig), got %d", len(todos))
	}
}
