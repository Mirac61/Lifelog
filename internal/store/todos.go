package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Todo struct {
	ID          int64
	Text        string
	DueDate     sql.NullString
	Done        bool
	CreatedAt   string
	CompletedAt sql.NullString
}

func ListTodos(ctx context.Context, db *sql.DB) ([]Todo, error) {
	const query = `SELECT id, text, due_date, done FROM todos ORDER BY done, due_date IS NULL, due_date`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.DueDate, &t.Done); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}
	return todos, nil
}

func CreateTodo(ctx context.Context, db *sql.DB, text string, dueDate *string) error {
	const query = `INSERT INTO todos (text, due_date, created_at) VALUES (?, ?, ?)`
	if _, err := db.ExecContext(ctx, query, text, dueDate, time.Now()); err != nil {
		return fmt.Errorf("creating todo: %w", err)
	}
	return nil
}

/* func SwitchTodo(ctx context.Context, db *sql.DB, id int64) error {
	// TODO: A migration should be implemented to switch statuses from todo, in progress and done
} */
