package github

import (
	"context"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

const (
	contributionsQuery = `query($from: DateTime!, $to: DateTime!) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      contributionCalendar {
        totalContributions
        weeks {
          contributionDays { date contributionCount }
        }
      }
    }
  }
}`
	contributionYearsQuery = `query {
  viewer {
    contributionsCollection {
      contributionYears
    }
  }
}`
	pullRequestsQuery = `query($from: DateTime!, $to: DateTime!) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      pullRequestContributions(first: 100) {
        totalCount
        nodes {
          occurredAt
          pullRequest {
            title
            url
            repository {
              nameWithOwner
            }
          }
        }
      }
    }
  }
}`

	commitContributionsByRepository = `query($from: DateTime!, $to: DateTime!) {
  viewer {
    contributionsCollection(from: $from, to: $to) {
      startedAt
      endedAt
      commitContributionsByRepository(maxRepositories: 100) {
        repository {
          name
          nameWithOwner
          pushedAt
        }
        contributions(orderBy: {direction: DESC, field: COMMIT_COUNT}) {
          totalCount
        }
      }
    }
  }
}`
)

func (c *Client) FetchContributions(ctx context.Context, year int) (collector.RawItem, error) {
	from := fmt.Sprintf("%d-01-01T00:00:00Z", year)
	to := fmt.Sprintf("%d-12-31T23:59:59Z", year)
	variables := map[string]any{"from": from, "to": to}
	payload, err := c.Query(ctx, contributionsQuery, variables)
	if err != nil {
		return collector.RawItem{}, err
	}
	return collector.RawItem{
		ExternalID: fmt.Sprintf("contributions:%d", year),
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}, nil
}

func (c *Client) FetchRepoCommits(ctx context.Context, year int) (collector.RawItem, error) {
	from := fmt.Sprintf("%d-01-01T00:00:00Z", year)
	to := fmt.Sprintf("%d-12-31T23:59:59Z", year)
	variables := map[string]any{"from": from, "to": to}
	payload, err := c.Query(ctx, commitContributionsByRepository, variables)
	if err != nil {
		return collector.RawItem{}, err
	}
	return collector.RawItem{
		ExternalID: fmt.Sprintf("repo-commits:%d", year),
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}, nil
}
