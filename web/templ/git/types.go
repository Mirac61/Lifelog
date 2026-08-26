package git

import (
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
