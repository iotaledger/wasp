package shutdown

const (
	PriorityDatabase = iota

	PriorityPeering
	PriorityWebAPI
	PriorityBadgerGarbageCollection
)
