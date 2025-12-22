package asentric

// Severity represents alert severity level.
// This is a strict set of values - only these are valid.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// ExecutionRef contains reference to the execution context.
// This is populated by the engine, not by rules.
type ExecutionRef struct {
	TxHash      string
	BlockNumber uint64
}

// Alert represents the output of a rule evaluation.
// Alerts are:
// - Serializable (JSON-safe)
// - Immutable after creation
// - Infrastructure-agnostic (no delivery logic)
//
// Alert does NOT include:
// - Chain IDs (runtime responsibility)
// - Network information
// - Delivery metadata
type Alert struct {
	// Rule is the name of the rule that generated this alert
	Rule string

	// Severity indicates the severity level
	Severity Severity

	// Title is a short summary of the alert
	Title string

	// Description provides detailed information
	Description string

	// Ref contains execution context (populated by engine)
	Ref *ExecutionRef

	// Metadata contains additional key-value data
	// MUST be JSON-serializable
	// MUST NOT be mutated after Alert creation
	Metadata map[string]any
}
