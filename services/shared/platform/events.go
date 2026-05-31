package platform

import "time"

type EventEnvelope struct {
	EventID    string         `json:"event_id"`
	EventName  string         `json:"event_name"`
	OccurredAt time.Time      `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}
