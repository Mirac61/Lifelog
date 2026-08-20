package github

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mirac61/lifelog/internal/collector"
)

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
