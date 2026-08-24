package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type repoRow struct {
	Repo    string
	Commits int
	Percent float64
}

type viewData struct {
	Stats   store.Stats
	Heatmap heatmapData
}

// typeLabel is the human-facing name for a pill; typeLabels holds the ones
// the two day-granularity collectors actually emit; anything else falls
// back to its raw value so a new collector still renders instead of erroring.
var typeLabels = map[string]string{
	"contributions": "Contributions",
	"pull_request":  "Pull Requests",
}

type typeOption struct {
	Value string
	Label string
}

type gitPage struct {
	Active      string
	Year        int
	PrevYear    int
	NextYear    int
	Type        string
	TypeLabel   string
	TypeOptions []typeOption
	Stats       store.Stats
	Heatmap     heatmapData
	Repos       []repoRow
	PRs         []store.PullRequest
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

func (s *Server) loadGitPage(ctx context.Context, year int, eventType string) (gitPage, error) {
	firstYear, lastYear, err := store.EventYearRange(ctx, s.db)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching commit years: %w", err)
	}
	if firstYear == 0 {
		firstYear = year
	}
	if lastYear < year {
		lastYear = year
	}

	types, err := store.ListTypes(ctx, s.db)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching event types: %w", err)
	}
	if !slices.Contains(types, eventType) {
		eventType = "contributions"
		if len(types) > 0 && !slices.Contains(types, eventType) {
			eventType = types[0]
		}
	}
	typeOptions := make([]typeOption, len(types))
	typeLabel := eventType
	for i, t := range types {
		label, ok := typeLabels[t]
		if !ok {
			label = t
		}
		if t == eventType {
			typeLabel = label
		}
		typeOptions[i] = typeOption{Value: t, Label: label}
	}

	var prevYear, nextYear int
	if year-1 >= firstYear {
		prevYear = year - 1
	}
	if year+1 <= lastYear {
		nextYear = year + 1
	}

	commits, err := store.ListRepoCommits(ctx, s.db, year)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching commits: %w", err)
	}
	var rows []repoRow
	if len(commits) > 0 {
		max := commits[0].Value
		for _, c := range commits {
			rows = append(rows, repoRow{
				Repo:    c.NameWithOwner,
				Commits: int(c.Value),
				Percent: c.Value / max * 100,
			})
		}
	}

	view, err := s.view(ctx, eventType, year)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching view: %w", err)
	}

	prs, err := store.ListPullRequests(ctx, s.db, year)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching pull requests: %w", err)
	}

	// Die Git-Seite hat keine Tagesdetail-Ansicht, Zellen bleiben inert.
	view.Heatmap.Detail = false
	return gitPage{
		Active:      "git",
		Year:        year,
		PrevYear:    prevYear,
		NextYear:    nextYear,
		Type:        eventType,
		TypeLabel:   typeLabel,
		TypeOptions: typeOptions,
		Stats:       view.Stats,
		Heatmap:     view.Heatmap,
		Repos:       rows,
		PRs:         prs,
	}, nil
}

func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	page, err := s.loadGitPage(r.Context(), time.Now().Year(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.renderPage(w, "git.html", page)
}

func (s *Server) handleGitView(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	if value := r.URL.Query().Get("year"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			http.Error(w, "invalid year", http.StatusBadRequest)
			return
		}
		year = parsed
	}

	page, err := s.loadGitPage(r.Context(), year, r.URL.Query().Get("type"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPartial(w, "gitview.html", page)
}
