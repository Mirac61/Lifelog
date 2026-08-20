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
