package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSwitchTodoAndOrdering(t *testing.T) {
	ctx := context.Background()
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}

	due := "2026-03-01"
	for _, text := range []string{"a", "b", "c"} {
		if err := CreateTodo(ctx, db, text, &due); err != nil {
			t.Fatal(err)
		}
	}

	todos, err := ListTodos(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 3 || todos[0].Status != "todo" {
		t.Fatalf("fresh todos = %+v, want three with status todo", todos)
	}

	// Swapped bind arguments would leave both of these untouched and still return nil.
	if err := SwitchTodo(ctx, db, todos[0].ID, "done"); err != nil {
		t.Fatal(err)
	}
	if err := SwitchTodo(ctx, db, todos[1].ID, "in_progress"); err != nil {
		t.Fatal(err)
	}

	got, err := ListTodos(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"in_progress", "todo", "done"}
	for i, status := range want {
		if got[i].Status != status {
			t.Errorf("ListTodos[%d].Status = %q, want %q (order: %+v)", i, got[i].Status, status, got)
		}
	}

	if err := SwitchTodo(ctx, db, todos[2].ID, "bogus"); err == nil {
		t.Error("SwitchTodo accepted an invalid status, want CHECK constraint error")
	}
}
