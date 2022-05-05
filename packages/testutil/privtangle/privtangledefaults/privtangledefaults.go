package privtangledefaults

const (
	Host     = "http://localhost"
	BasePort = 16500

	NodePortOffsetRestAPI = iota
	NodePortOffsetPeering
	NodePortOffsetDashboard
	NodePortOffsetProfiling
	NodePortOffsetPrometheus
	NodePortOffsetFaucet
	NodePortOffsetMQTT
	NodePortOffsetMQTTWebSocket
	NodePortOffsetIndexer
	NodePortOffsetINX
)
