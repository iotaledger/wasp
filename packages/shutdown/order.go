package shutdown

const (
	PriorityDatabase = iota

	PriorityPeering
	PriorityNodeConnection
	PriorityWebAPI
	PriorityBadgerGarbageCollection
)
