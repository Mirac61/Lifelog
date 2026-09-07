package git

// GitPage is one year of GitHub activity, already shaped for the design's
// panes. Everything here is a string or a number the template prints as is —
// the arithmetic and the wording happen in internal/api.
type GitPage struct {
	Year        int
	PrevYear    int
	NextYear    int
	Type        string
	TypeLabel   string
	TypeOptions []TypeOption

	HeadMeta string // "739 Beiträge · 13 Repos · 41 Pull Requests"

	Lead     Lead
	Heatmap  Heatmap
	Months   []MonthBar
	Repos    []RepoRow
	RepoRest string // "Top 5 von 13"
	PRs      []PRGroup
	PRCount  string // "41 im Jahr"
	PRRest   string // "31 weitere Pull Requests aus Januar bis Juli."
	Insights []Insight
}

type TypeOption struct {
	Value string
	Label string
}

// Sentence is a line with one emphasised part. Em is the accent-coloured
// middle; both halves may be empty.
type Sentence struct {
	Before string
	Em     string
	After  string
}

// Lead is the opening pane: the year's figure, one sentence, four numbers.
type Lead struct {
	Total    string
	Caption  string // "Beiträge in 2026"
	Sentence Sentence
	Stats    []Figure
}

type Figure struct {
	Value string
	Label string
}

// MonthBar is one column of the monthly chart. Height is in px against the
// design's 72px peak.
type MonthBar struct {
	Label  string
	Value  string
	Height int
	Peak   bool
}

// Heatmap is the day grid: seven rows, one column per week, filled
// column-first. Weeks sizes the month row above it.
type Heatmap struct {
	Weeks  int
	Months []MonthSpan
	Cells  []HeatCell
	Label  string // aria-label, since the cells themselves are decorative
}

// MonthSpan places a month label over the weeks it covers (1-based grid
// columns, To exclusive).
type MonthSpan struct {
	Label string
	From  int
	To    int
}

// HeatCell is one day. Void days lie outside the year or still ahead of it;
// they take no colour and carry no tooltip.
type HeatCell struct {
	Level int // 0 = no activity, 1..4 = the design's ramp
	Void  bool
	Best  bool
	Date  string
	Tip   string
}

type RepoRow struct {
	Rank    string
	Owner   string // "Mirac61/", printed muted before the name
	Name    string
	Percent float64
	Value   string
}

// PRGroup is a month of pull requests, newest month first.
type PRGroup struct {
	Month string
	Rows  []PRRow
}

type PRRow struct {
	Date  string // "26 Aug"
	State string // open · merged · closed, drives the dot colour
	Label string
	Title string
	Repo  string
	URL   string
}

// Insight is one line of the closing pane: a figure and the sentence it
// carries. Plain prose — <em> is the accent inside .gh-sentence only, and
// would come out italic here.
type Insight struct {
	Key  string
	Text string
}
