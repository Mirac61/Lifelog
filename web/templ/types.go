package templ

import (
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type RepoRow struct {
	Repo    string
	Commits int
	Percent float64
}

type TypeOption struct {
	Value string
	Label string
}

type GitPage struct {
	Active        string
	Year          int
	IsCurrentYear bool
	PrevYear      int
	NextYear      int
	Type          string
	TypeLabel     string
	TypeOptions   []TypeOption
	Stats         store.Stats
	Heatmap       HeatmapData
	Repos         []RepoRow
	PRs           []store.PullRequest
}

type Cell struct {
	X, Y   int
	Level  int
	Date   string
	Label  string
	Value  float64
	Best   bool
	Future bool
}

type TextLabel struct {
	X, Y int
	Text string
}

type Swatch struct {
	X, Y  int
	Level int
}

type HeatmapData struct {
	Width       int
	Height      int
	CellSize    int
	Cells       []Cell
	Months      []TextLabel
	Days        []TextLabel
	Legend      []Swatch
	LessX       int
	MoreX       int
	LegendY     int
	Detail      bool
	HasBest     bool
	DividerX    int
	BestSwatchX int
	BestLabelX  int
}

// bestDaySuffix renders " · <date>" when a best date exists, or nothing.
func bestDaySuffix(date string) string {
	if date == "" {
		return ""
	}
	return " · " + formatShortDate(date)
}

// formatShortDate renders an ISO date as "2 Jan", matching the heatmap's
// month-label format so every date in the git view reads the same way.
func formatShortDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("2 Jan")
}
