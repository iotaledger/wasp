package nodeconnmetrics

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	chainLabelNameConst   = "chain"
	msgTypeLabelNameConst = "message_type"
)

func getMetricsLabel(chainAddr iotago.Address, msgType string) prometheus.Labels {
	var chainAddrStr string
	if chainAddr == nil {
		chainAddrStr = ""
	} else {
		chainAddrStr = chainAddr.String()
	}
	return prometheus.Labels{
		chainLabelNameConst:   chainAddrStr,
		msgTypeLabelNameConst: msgType,
	}
}
