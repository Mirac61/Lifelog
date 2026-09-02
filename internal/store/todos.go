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
	Status      string
	CreatedAt   string
	CompletedAt sql.NullString
	Estimate    sql.NullInt64
	Category    sql.NullString
}

func ListTodos(ctx context.Context, db *sql.DB) ([]Todo, error) {
	const query = `SELECT id, text, due_date, status FROM todos ORDER BY CASE status
			WHEN 'in_progress' THEN 0 WHEN 'todo' THEN 1 ELSE 2 END, due_date IS NULL, due_date`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.DueDate, &t.Status); err != nil {
			return nil, fmt.Errorf("scan todo: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todos: %w", err)
	}
	return todos, nil
}

func ListTodosForDay(ctx context.Context, db *sql.DB, day string) ([]Todo, error) {
	const query = `SELECT id, text, due_date, status, estimate, category FROM todos	
								 WHERE due_date = ? OR (due_date < ? AND status != 'done')`
	rows, err := db.QueryContext(ctx, query, day, day)
	if err != nil {
		return nil, fmt.Errorf("query todo list: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Text, &t.DueDate, &t.Status, &t.Estimate, &t.Category); err != nil {
			return nil, fmt.Errorf("scan todo list: %w", err)
		}
		todos = append(todos, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate todo list: %w", err)
	}
	return todos, nil
}

func CreateTodo(ctx context.Context, db *sql.DB, t Todo) (Todo, error) {
	const query = `INSERT INTO todos (text, due_date, estimate, category, created_at) VALUES (?, ?, ?, ?, ?)
			RETURNING id, text, due_date, status, estimate, category`
	var out Todo
	if err := db.QueryRowContext(ctx, query, t.Text, t.DueDate, t.Estimate, t.Category, time.Now()).
		Scan(&out.ID, &out.Text, &out.DueDate, &out.Status, &out.Estimate, &out.Category); err != nil {
		return Todo{}, fmt.Errorf("creating todo: %w", err)
	}
	return out, nil
}

func SwitchTodo(ctx context.Context, db *sql.DB, id int64, status string) (Todo, error) {
	const query = `UPDATE todos SET status = ? WHERE id = ? RETURNING id, text, due_date, status, estimate, category`
	var t Todo
	if err := db.QueryRowContext(ctx, query, status, id).Scan(&t.ID, &t.Text, &t.DueDate, &t.Status, &t.Estimate, &t.Category); err != nil {
		return Todo{}, fmt.Errorf("switching todo status: %w", err)
	}
	return t, nil
}

// DeleteTodo returns the day the todo belonged to, so the caller can
// recompute that day's budget.
func DeleteTodo(ctx context.Context, db *sql.DB, id int64) (sql.NullString, error) {
	const query = `DELETE FROM todos WHERE id = ? RETURNING due_date`
	var due sql.NullString
	if err := db.QueryRowContext(ctx, query, id).Scan(&due); err != nil {
		return due, fmt.Errorf("deleting todo: %w", err)
	}
	return due, nil
}

// PlannedForDay sums the estimates of everything still unfinished.
func PlannedForDay(ctx context.Context, db *sql.DB, day string) (int, error) {
	todos, err := ListTodosForDay(ctx, db, day)
	if err != nil {
		return 0, err
	}
	var planned int
	for _, t := range todos {
		if t.Estimate.Valid && t.Status != "done" {
			planned += int(t.Estimate.Int64)
		}
	}
	return planned, nil
}
