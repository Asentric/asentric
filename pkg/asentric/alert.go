package asentric

import "time"

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

type Alert struct {
	ID       string
	RuleID   string
	Severity Severity

	Summary   string
	Evidence  *Event
	Timestamp time.Time
}
