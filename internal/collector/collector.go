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
	LocalDate  string
	Value      float64
	Unit       string
	Meta       json.RawMessage
}
