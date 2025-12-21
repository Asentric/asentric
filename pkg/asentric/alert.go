package asentric

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Alert struct {
	Rule     string
	Severity Severity

	Title   string
	Message string

	TxHash      string
	BlockNumber uint64

	Labels map[string]string
}
