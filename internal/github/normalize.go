package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

// NormalizeContributions
type contributionDay struct {
	Date              string `json:"date"`
	ContributionCount int    `json:"contributionCount"`
}

type week struct {
	ContributionDays []contributionDay `json:"contributionDays"`
}

type contributionCalendar struct {
	Weeks []week `json:"weeks"`
}

type contributionsCollection struct {
	ContributionCalendar contributionCalendar `json:"contributionCalendar"`
}

type viewer struct {
	ContributionsCollection contributionsCollection `json:"contributionsCollection"`
}

type contributionsResponse struct {
	Viewer viewer `json:"viewer"`
}

// NormalizeContributionYears
type yearsContributionsCollection struct {
	ContributionYears []int `json:"contributionYears"`
}

type yearsViewer struct {
	ContributionsCollection yearsContributionsCollection `json:"contributionsCollection"`
}

type yearsResponse struct {
	Viewer yearsViewer `json:"viewer"`
}

// NormalizePRContributions
type prRepository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type prPullRequest struct {
	Title      string       `json:"title"`
	URL        string       `json:"url"`
	Repository prRepository `json:"repository"`
}

type prNode struct {
	OccurredAt  string        `json:"occurredAt"`
	PullRequest prPullRequest `json:"pullRequest"`
}

type prContributions struct {
	TotalCount int      `json:"totalCount"`
	Nodes      []prNode `json:"nodes"`
}

type prContributionsCollection struct {
	PullRequestContributions prContributions `json:"pullRequestContributions"`
}

type prViewer struct {
	ContributionsCollection prContributionsCollection `json:"contributionsCollection"`
}

type prResponse struct {
	Viewer prViewer `json:"viewer"`
}

// NormalizeRepoContributions
type repRepository struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
	PushedAt      string `json:"pushedAt"`
}

type repContributions struct {
	TotalCount int `json:"totalCount"`
}

type repContributionsByRepository struct {
	Repository    repRepository    `json:"repository"`
	Contributions repContributions `json:"contributions"`
}

type repContributionsCollection struct {
	StartedAt                       string                         `json:"startedAt"`
	EndedAt                         string                         `json:"endedAt"`
	CommitContributionsByRepository []repContributionsByRepository `json:"commitContributionsByRepository"`
}

type repViewer struct {
	ContributionsCollection repContributionsCollection `json:"contributionsCollection"`
}

type repResponse struct {
	Viewer repViewer `json:"viewer"`
}

func NormalizeContributions(payload json.RawMessage) ([]collector.Event, error) {
	var response contributionsResponse
	err := json.Unmarshal(payload, &response)
	if err != nil {
		return nil, fmt.Errorf("decode contributions: %w", err)
	}
	var events []collector.Event
	weeks := response.Viewer.ContributionsCollection.ContributionCalendar.Weeks
	for _, w := range weeks {
		for _, d := range w.ContributionDays {
			if d.ContributionCount == 0 {
				continue
			}
			occurredAt, err := time.Parse("2006-01-02", d.Date)
			if err != nil {
				return nil, fmt.Errorf("parse date %s: %w", d.Date, err)
			}
			events = append(events, collector.Event{
				Type:       "commit",
				OccurredAt: occurredAt,
				LocalDate:  d.Date,
				Value:      float64(d.ContributionCount),
				Unit:       "count",
			})
		}
	}
	return events, nil
}

func NormalizeContributionYears(payload json.RawMessage) ([]int, error) {
	var response yearsResponse
	err := json.Unmarshal(payload, &response)
	if err != nil {
		return nil, fmt.Errorf("decode contribution year: %w", err)
	}
	return response.Viewer.ContributionsCollection.ContributionYears, nil
}

func NormalizePRContributions(payload json.RawMessage) ([]collector.Event, error) {
	var response prResponse
	err := json.Unmarshal(payload, &response)
	if err != nil {
		return nil, fmt.Errorf("decode PR contributions: %w", err)
	}
	var events []collector.Event
	nodes := response.Viewer.ContributionsCollection.PullRequestContributions.Nodes
	for _, node := range nodes {
		title := node.PullRequest.Title
		url := node.PullRequest.URL
		repository := node.PullRequest.Repository.NameWithOwner
		occurredAt, err := time.Parse(time.RFC3339, node.OccurredAt)
		if err != nil {
			return nil, fmt.Errorf("parse date %s: %w", node.OccurredAt, err)
		}
		meta, err := json.Marshal(map[string]string{
			"title":      title,
			"url":        url,
			"repository": repository,
		})
		if err != nil {
			return nil, fmt.Errorf("encode PR metadata: %w", err)
		}

		events = append(events, collector.Event{
			Type:       "pull_request",
			OccurredAt: occurredAt,
			LocalDate:  occurredAt.Format("2006-01-02"),
			Value:      1,
			Unit:       "pr",
			Meta:       meta,
		})
	}
	return events, nil
}

func NormalizeRepoContributions(payload json.RawMessage) ([]collector.Event, error) {
	var response repResponse
	if err := json.Unmarshal(payload, &response); err != nil {
		return nil, fmt.Errorf("decode repo contributions: %w", err)
	}

	collection := response.Viewer.ContributionsCollection
	startedAt, err := time.Parse(time.RFC3339, collection.StartedAt)
	if err != nil {
		return nil, fmt.Errorf("parse startedAt %q: %w", collection.StartedAt, err)
	}

	var events []collector.Event
	for _, repo := range collection.CommitContributionsByRepository {
		meta, err := json.Marshal(map[string]string{
			"name":          repo.Repository.Name,
			"nameWithOwner": repo.Repository.NameWithOwner,
			"pushedAt":      repo.Repository.PushedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("encode repo metadata: %w", err)
		}

		events = append(events, collector.Event{
			Type:       "repo_commit",
			OccurredAt: startedAt,
			LocalDate:  startedAt.Format("2006-01-02"),
			Value:      float64(repo.Contributions.TotalCount),
			Unit:       "commits",
			Meta:       meta,
		})
	}
	return events, nil
}
