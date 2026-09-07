package api

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
	"github.com/Mirac61/lifelog/web/templ/git"
)

// How an event type is written in prose; unlisted types fall back to their
// raw value.
var typeNames = map[string]struct{ Plural, Singular string }{
	"contributions": {"Beiträge", "ein Beitrag"},
	"pull_request":  {"Pull Requests", "ein Pull Request"},
}

func (s *Server) yearStats(ctx context.Context, eventType string, year int) (store.Stats, error) {
	from, to := fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
	return store.DailyStats(ctx, s.db, eventType, from, to)
}

func (s *Server) loadGitPage(ctx context.Context, year int, eventType string) (git.GitPage, error) {
	firstYear, lastYear, err := store.EventYearRange(ctx, s.db)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("fetching commit years: %w", err)
	}
	if firstYear == 0 {
		firstYear = year
	}
	if lastYear < year {
		lastYear = year
	}

	types, err := store.ListTypes(ctx, s.db)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("fetching event types: %w", err)
	}
	if !slices.Contains(types, eventType) {
		eventType = "contributions"
		if len(types) > 0 && !slices.Contains(types, eventType) {
			eventType = types[0]
		}
	}
	names, ok := typeNames[eventType]
	if !ok {
		names.Plural, names.Singular = eventType, eventType
	}
	typeOptions := make([]git.TypeOption, len(types))
	for i, t := range types {
		label := t
		if n, ok := typeNames[t]; ok {
			label = n.Plural
		}
		typeOptions[i] = git.TypeOption{Value: t, Label: label}
	}

	var prevYear, nextYear int
	if year-1 >= firstYear {
		prevYear = year - 1
	}
	if year+1 <= lastYear {
		nextYear = year + 1
	}

	from, to := fmt.Sprintf("%d-01-01", year), fmt.Sprintf("%d-12-31", year)
	totals, err := store.DailyTotals(ctx, s.db, eventType, from, to)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("daily totals: %w", err)
	}
	stats, err := s.yearStats(ctx, eventType, year)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("daily stats: %w", err)
	}
	var prev store.Stats
	if prevYear != 0 {
		if prev, err = s.yearStats(ctx, eventType, prevYear); err != nil {
			return git.GitPage{}, fmt.Errorf("previous year stats: %w", err)
		}
	}
	repos, err := store.ListRepoCommits(ctx, s.db, year)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("fetching commits: %w", err)
	}
	prs, err := store.ListPullRequests(ctx, s.db, year)
	if err != nil {
		return git.GitPage{}, fmt.Errorf("fetching pull requests: %w", err)
	}

	view := gitView{
		year:      year,
		prevYear:  prevYear,
		label:     names.Plural,
		singular:  names.Singular,
		eventType: eventType,
		stats:     stats,
		prev:      prev,
		totals:    totals,
		repos:     repos,
		prs:       prs,
		months:    buildMonths(totals, year),
	}
	repoRows, repoRest := view.repoRows()
	prGroups, prCount, prRest := view.prGroups()

	return git.GitPage{
		Year:        year,
		PrevYear:    prevYear,
		NextYear:    nextYear,
		Type:        eventType,
		TypeLabel:   names.Plural,
		TypeOptions: typeOptions,
		HeadMeta:    view.headMeta(),
		Lead:        view.lead(),
		Heatmap:     buildHeatmap(totals, year, stats, names.Plural),
		Months:      view.months,
		Repos:       repoRows,
		RepoRest:    repoRest,
		PRs:         prGroups,
		PRCount:     prCount,
		PRRest:      prRest,
		Insights:    view.insights(),
	}, nil
}

// gitQuery reads the year and type a request asks for; both are optional and
// fall back to the current year and the default type.
func gitQuery(r *http.Request) (int, string, error) {
	year := time.Now().Year()
	if value := r.URL.Query().Get("year"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return 0, "", fmt.Errorf("invalid year")
		}
		year = parsed
	}
	return year, r.URL.Query().Get("type"), nil
}

func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	year, eventType, err := gitQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, err := s.loadGitPage(r.Context(), year, eventType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, git.Git(page))
}

func (s *Server) handleGitView(w http.ResponseWriter, r *http.Request) {
	year, eventType, err := gitQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	page, err := s.loadGitPage(r.Context(), year, eventType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	render(w, r, git.YearSwap(page))
}
