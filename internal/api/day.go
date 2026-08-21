package api

import (
	"net/http"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
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

	data := struct {
		Date   string
		Events []store.DayEvent
	}{day.Format("Mon, 2 Jan 2006"), events}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "day.html", data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}
