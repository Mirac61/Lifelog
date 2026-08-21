package api

import (
	"context"
	"fmt"
	"log/slog"
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
	return viewData{stats, buildHeatmap(totals, year)}, nil
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
		Year    int
		Years   []int
		Types   []string
		Stats   store.Stats
		Heatmap heatmapData
	}{year, years, types, view.Stats, view.Heatmap}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("render index", "error", err)
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}
