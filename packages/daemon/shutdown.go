package daemon

// Please add the dependencies if you add your own priority here.
// Otherwise investigating deadlocks at shutdown is much more complicated.

const (
	PriorityCloseDatabase = iota // no dependencies
	PriorityDatabaseHealth
	PriorityNodeConnection
	PriorityPeering
	PriorityChains
	PriorityWebAPI
	PriorityPrometheus
	PriorityPublisher
	PriorityDashboard
	PriorityProfilingRecorder
)
