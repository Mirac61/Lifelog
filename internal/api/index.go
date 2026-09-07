package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/dashboard"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	types, err := store.ListTypes(r.Context(), s.db)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	eventType := "contributions"
	if len(types) > 0 {
		eventType = types[0]
	}

	year := time.Now().Year()
	from, to := fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
	stats, err := store.DailyStats(r.Context(), s.db, eventType, from, to)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	render(w, r, dashboard.Page(stats, year))
}
