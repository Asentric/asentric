package domain

// Event represents an ABI-decoded event
// Fields are decoded and ready for rule evaluation
// No topic-level exposure to rules
type Event struct {
	Name   string
	Fields map[string]any
}

