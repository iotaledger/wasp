package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

const (
	labelNameChain                             = "chain"
	labelNamePipeName                          = "pipe_name"
	labelNameMessageType                       = "message_type"
	labelNameInAnchorMetrics                   = "in_anchor"
	labelNameInOnLedgerRequestMetrics          = "in_on_ledger_request"
	labelNameOutPublishStateTransactionMetrics = "out_publish_state_transaction"
	labelTxPublishResult                       = "result"
	labelNameWebapiRequestOperation            = "api_req_type"
	labelNameWebapiRequestStatusCode           = "api_req_status_code"
	labelNameWebapiEvmRPCSuccess               = "success"
)

func getChainLabels(chainID isc.ChainID) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain: chainID.String(),
	}
}
