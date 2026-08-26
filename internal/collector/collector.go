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
	// LocalDate is this event's period start; Granularity is the period length ("day" or "year").
	LocalDate   string
	Granularity string
	Value       float64
	Unit        string
	Meta        json.RawMessage
}
