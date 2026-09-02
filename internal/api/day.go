package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	views "github.com/Mirac61/lifelog/web/templ"
	"github.com/Mirac61/lifelog/web/templ/todo"
)

// loadTodos counts only unfinished todos into Planned: the time still ahead.
func (s *Server) loadTodos(ctx context.Context, day time.Time) (todo.DayPage, error) {
	todos, err := store.ListTodosForDay(ctx, s.db, day.Format("2006-01-02"))
	if err != nil {
		return todo.DayPage{}, err
	}
	page := todo.DayPage{Day: day, Todos: todos}
	for _, t := range todos {
		if t.Estimate.Valid && t.Status != "done" {
			page.Planned += int(t.Estimate.Int64)
		}
	}
	return page, nil
}

func (s *Server) handleDay(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	todos, err := s.loadTodos(r.Context(), day)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	events, err := store.EventsForDay(r.Context(), s.db, date)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	render(w, r, views.Day(events, todos))
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	day := time.Now()
	if date := r.URL.Query().Get("date"); date != "" {
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil {
			http.Error(w, "invalid date", http.StatusBadRequest)
			return
		}
		day = parsed
	}
	page, err := s.loadTodos(r.Context(), day)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	render(w, r, todo.Page(page))
}
