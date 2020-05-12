package shutdown

const (
	PriorityDatabase = iota

	PriorityPeering
	PriorityNodeConnection
	PriorityDispatcher
	PriorityWebAPI
	PriorityBadgerGarbageCollection
)
