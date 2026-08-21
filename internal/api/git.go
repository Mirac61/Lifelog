package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Mirac61/lifelog/internal/store"
)

type repoRow struct {
	Repo    string
	Commits int
	Percent float64
}

type gitPage struct {
	Active  string
	Year    int
	Years   []int
	Stats   store.Stats
	Heatmap heatmapData
	Repos   []repoRow
}

func (s *Server) loadGitPage(ctx context.Context, year int) (gitPage, error) {
	firstYear, err := store.FirstRepoCommitYear(ctx, s.db)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching commit years: %w", err)
	}
	if firstYear == 0 {
		firstYear = year
	}
	var years []int
	for y := year; y >= firstYear; y-- {
		years = append(years, y)
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

	view, err := s.view(ctx, "commit", year)
	if err != nil {
		return gitPage{}, fmt.Errorf("fetching view: %w", err)
	}
	// Die Git-Seite hat keine Tagesdetail-Ansicht, Zellen bleiben inert.
	view.Heatmap.Detail = false
	return gitPage{
		Active:  "git",
		Year:    year,
		Years:   years,
		Stats:   view.Stats,
		Heatmap: view.Heatmap,
		Repos:   rows,
	}, nil
}

func (s *Server) handleGit(w http.ResponseWriter, r *http.Request) {
	page, err := s.loadGitPage(r.Context(), time.Now().Year())
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

	page, err := s.loadGitPage(r.Context(), year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.renderPartial(w, "gitview.html", page)
}
