package api

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/todo"
)

func (s *Server) handleTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	status := r.PathValue("status")
	if !slices.Contains([]string{"todo", "in_progress", "done"}, status) {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	t, err := store.SwitchTodo(r.Context(), s.db, id, status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "toggle failed", http.StatusInternalServerError)
		return
	}
	s.renderRow(w, r, t)
}

func (s *Server) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	t := parseTodoInput(r.FormValue("text"))
	if t.Text == "" {
		http.Error(w, "empty todo", http.StatusBadRequest)
		return
	}
	if date := r.FormValue("date"); date != "" {
		t.DueDate = sql.NullString{String: date, Valid: true}
	}
	created, err := store.CreateTodo(r.Context(), s.db, t)
	if err != nil {
		http.Error(w, "create failed", http.StatusInternalServerError)
		return
	}
	s.renderRow(w, r, created)
}

func (s *Server) handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	due, err := store.DeleteTodo(r.Context(), s.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "todo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	// Empty body swaps the row away; only the budget comes back, out of band.
	render(w, r, todo.Planned(s.planned(r, due), true))
}

// renderRow returns the changed row plus an out-of-band budget update.
func (s *Server) renderRow(w http.ResponseWriter, r *http.Request, t store.Todo) {
	render(w, r, todo.ItemUpdate(t, s.planned(r, t.DueDate)))
}

func (s *Server) planned(r *http.Request, due sql.NullString) int {
	if !due.Valid {
		return 0
	}
	n, err := store.PlannedForDay(r.Context(), s.db, due.String)
	if err != nil {
		return 0
	}
	return n
}

// A bare number stays text, so "Kapitel 90 lesen" is not read as an estimate.
var durationRe = regexp.MustCompile(`^(?:(\d+)h)?(?:(\d+)m)?$`)

// parseTodoInput reads "Kaffee kochen 15m #haushalt" into text, estimate and category.
func parseTodoInput(input string) store.Todo {
	var t store.Todo
	var words []string
	for _, w := range strings.Fields(input) {
		switch {
		case strings.HasPrefix(w, "#") && len(w) > 1:
			t.Category = sql.NullString{String: w[1:], Valid: true}
		case durationRe.MatchString(w) && strings.ContainsAny(w, "hm"):
			m := durationRe.FindStringSubmatch(w)
			hours, _ := strconv.ParseInt(m[1], 10, 64)
			mins, _ := strconv.ParseInt(m[2], 10, 64)
			t.Estimate = sql.NullInt64{Int64: hours*60 + mins, Valid: true}
		default:
			words = append(words, w)
		}
	}
	t.Text = strings.Join(words, " ")
	return t
}
