package collector

import (
	"encoding/json"
	"time"
)

type RawItem struct {
	ExternalID string
	OccurredAt time.Time
	Payload    json.RawMessage
}

type Event struct {
	Type       string
	OccurredAt time.Time
	// LocalDate is the start of the period this event covers; Granularity
	// says how long that period is ("day" or "year").
	LocalDate   string
	Granularity string
	Value       float64
	Unit        string
	Meta        json.RawMessage
}
