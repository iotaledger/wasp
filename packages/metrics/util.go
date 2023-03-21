package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

func getChainLabels(chainID isc.ChainID) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain: chainID.String(),
	}
}

func getChainMessageTypeLabels(chainID isc.ChainID, msgType string) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain:       chainID.String(),
		labelNameMessageType: msgType,
	}
}
