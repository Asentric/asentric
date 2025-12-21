package asentric

type EventType string

const (
	EventBlock EventType = "block"
	EventLog   EventType = "log"
)

type Event interface {
	Type() EventType
	Raw() any
}
