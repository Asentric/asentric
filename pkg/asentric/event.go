package asentric

import "time"

type EventType string

type Event struct {
	ID        string
	Type      EventType
	Timestamp time.Time

	Actor  string
	Target string

	Attributes map[string]any
}
