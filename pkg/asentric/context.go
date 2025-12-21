package asentric

type Context struct {
	Event *Event

	Network string
	ChainID int64
	Block   uint64

	// future: historical snapshot, cache, dll
}
