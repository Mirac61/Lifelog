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

func TestBudgetForDay(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `INSERT INTO todos (text, due_date, status, estimate, created_at) VALUES
		('offen mit Schaetzung', '2026-08-27', 'todo',        120, '2026-08-27'),
		('schon erledigt',       '2026-08-27', 'done',         15, '2026-08-27'),
		('angefangen, ungenau',  '2026-08-27', 'in_progress', NULL, '2026-08-27')`); err != nil {
		t.Fatal(err)
	}

	open, minutes, err := BudgetForDay(ctx, db, "2026-08-27")
	if err != nil {
		t.Fatalf("BudgetForDay: %v", err)
	}
	if open != 2 {
		t.Errorf("open = %d, want 2 (erledigtes zaehlt nicht mit)", open)
	}
	if minutes != 120 {
		t.Errorf("minutes = %d, want 120 (nur die offene Schaetzung)", minutes)
	}
}
