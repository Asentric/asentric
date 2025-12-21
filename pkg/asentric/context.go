package asentric

import "time"

type Context interface {
	Event() Event

	BlockNumber() uint64
	TxHash() string
	Timestamp() time.Time

	Tags() map[string]string
}
