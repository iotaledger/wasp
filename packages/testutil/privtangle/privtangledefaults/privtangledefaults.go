package privtangledefaults

const (
	Host     = "http://localhost"
	INXHost  = "localhost"
	BasePort = 16500

	NodePortOffsetRestAPI = iota
	NodePortOffsetPeering
	NodePortOffsetDashboard
	NodePortOffsetProfiling
	NodePortOffsetPrometheus
	NodePortOffsetFaucet
	NodePortOffsetMQTT
	NodePortOffsetIndexer
	NodePortOffsetINX
	NodePortOffsetCoordinator
)
