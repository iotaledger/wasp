package nodeconnmetrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

const (
	chainLabelNameConst   = "chain"
	msgTypeLabelNameConst = "message_type"
)

func getMetricsLabel(chainID *isc.ChainID, msgType string) prometheus.Labels {
	var chainIDStr string
	if chainID == nil {
		chainIDStr = ""
	} else {
		chainIDStr = chainID.String()
	}
	return prometheus.Labels{
		chainLabelNameConst:   chainIDStr,
		msgTypeLabelNameConst: msgType,
	}
}
