package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type viewData struct {
	Stats   store.Stats
	Heatmap heatmapData
}

func (s *Server) view(ctx context.Context, eventType string, year int) (viewData, error) {
	from, to := fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)

	totals, err := store.DailyTotals(ctx, s.db, eventType, from, to)
	if err != nil {
		return viewData{}, fmt.Errorf("daily totals: %w", err)
	}
	stats, err := store.DailyStats(ctx, s.db, eventType, from, to)
	if err != nil {
		return viewData{}, fmt.Errorf("daily stats: %w", err)
	}
	hm := buildHeatmap(totals, year, stats.BestDate)
	hm.Detail = true
	return viewData{stats, hm}, nil
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	types, err := store.ListTypes(r.Context(), s.db)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	if len(types) == 0 {
		types = []string{"commit"}
	}

	year := time.Now().Year()
	years := []int{year, year - 1, year - 2, year - 3}

	view, err := s.view(r.Context(), types[0], year)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}

	data := struct {
		Active  string
		Year    int
		Years   []int
		Type    string
		Types   []string
		Stats   store.Stats
		Heatmap heatmapData
	}{"dashboard", year, years, types[0], types, view.Stats, view.Heatmap}

	s.renderPage(w, "dashboard.html", data)
}
