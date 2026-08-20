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
	//contributionYearsQuery          = ``
	//pullRequestsQuery               = ``
	//commitContributionsByRepository = ``
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
