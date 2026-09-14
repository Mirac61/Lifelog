package api

import (
	"net/http"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/dashboard"
)

func (s *Server) handleDay(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, "invalid date", http.StatusBadRequest)
		return
	}
	events, err := store.EventsForDay(r.Context(), s.db, date)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	render(w, r, dashboard.Day(events, day))
}
