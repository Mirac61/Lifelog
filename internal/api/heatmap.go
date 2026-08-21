package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

const (
	cellSize   = 11
	cellGap    = 2
	cellStep   = cellSize + cellGap
	leftMargin = 32
	topMargin  = 20
)

var scale = [5]string{"#363646", "#2D4F67", "#4A7A96", "#7E9CD8", "#A3D4D5"}

func colorFor(v float64) string {
	switch {
	case v <= 0:
		return scale[0]
	case v <= 3:
		return scale[1]
	case v <= 8:
		return scale[2]
	case v <= 15:
		return scale[3]
	default:
		return scale[4]
	}
}

type cell struct {
	X, Y  int
	Color string
	Date  string
	Label string
	Value float64
}

type textLabel struct {
	X, Y int
	Text string
}

type swatch struct {
	X, Y  int
	Color string
}

type heatmapData struct {
	Width    int
	Height   int
	CellSize int
	Cells    []cell
	Months   []textLabel
	Days     []textLabel
	Legend   []swatch
	LessX    int
	MoreX    int
	LegendY  int
}

func buildHeatmap(totals []store.DailyTotal, year int) heatmapData {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	start = start.AddDate(0, 0, -((int(start.Weekday()) + 6) % 7))
	end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
	values := make(map[string]float64, len(totals))
	for _, t := range totals {
		values[t.LocalDate] = t.Value
	}
	lastMonth := time.Month(0)
	week := 0

	var cells []cell
	var months []textLabel
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		row := (int(d.Weekday()) + 6) % 7
		if row == 0 && d.After(start) {
			week++
		}
		if d.Year() != year {
			continue
		}
		if d.Month() != lastMonth && row < 3 {
			lastMonth = d.Month()
			months = append(months, textLabel{X: leftMargin + week*cellStep, Y: 14, Text: d.Format("Jan")})
		}
		date := d.Format("2006-01-02")
		v := values[date]
		cells = append(cells, cell{
			X:     leftMargin + week*cellStep,
			Y:     topMargin + row*cellStep,
			Color: colorFor(v),
			Date:  date,
			Label: d.Format("Mon, 2 Jan 2006"),
			Value: v,
		})
	}
	width := leftMargin + (week+1)*cellStep
	height := topMargin + 7*cellStep + 28
	legendY := height - 6
	var legend []swatch
	for i, c := range scale {
		legend = append(legend, swatch{
			X:     leftMargin + 30 + i*cellStep,
			Y:     legendY - 9,
			Color: c,
		})
	}
	var days []textLabel
	for i, name := range []string{"Mon", "Wed", "Fri"} {
		days = append(days, textLabel{
			X:    0,
			Y:    topMargin + i*2*cellStep + 9,
			Text: name,
		})
	}

	return heatmapData{
		Width:    width,
		Height:   height,
		CellSize: cellSize,
		Cells:    cells,
		Months:   months,
		Days:     days,
		Legend:   legend,
		LessX:    leftMargin,
		MoreX:    leftMargin + 36 + 5*cellStep,
		LegendY:  legendY,
	}
}

func (s *Server) handleHeatmap(w http.ResponseWriter, r *http.Request) {
	eventType := r.URL.Query().Get("type")
	if eventType == "" {
		eventType = "commit"
	}
	year := time.Now().Year()
	if y, err := strconv.Atoi(r.URL.Query().Get("year")); err == nil {
		year = y
	}
	from := fmt.Sprintf("%d-01-01", year)
	to := fmt.Sprintf("%d-12-31", year)
	totals, err := store.DailyTotals(r.Context(), s.db, eventType, from, to)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		return
	}
	data := buildHeatmap(totals, year)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "heatmap.html", data); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}
