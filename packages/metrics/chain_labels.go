package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/iotaledger/wasp/packages/isc"
)

const (
	labelNameChain                                  = "chain"
	labelNamePipeName                               = "pipe_name"
	labelNameMessageType                            = "message_type"
	labelNameInMilestone                            = "in_milestone"
	labelNameInStateOutputMetrics                   = "in_state_output"
	labelNameInAliasOutputMetrics                   = "in_alias_output"
	labelNameInOutputMetrics                        = "in_output"
	labelNameInOnLedgerRequestMetrics               = "in_on_ledger_request"
	labelNameInTxInclusionStateMetrics              = "in_tx_inclusion_state"
	labelNameOutPublishStateTransactionMetrics      = "out_publish_state_transaction"
	labelNameOutPublishGovernanceTransactionMetrics = "out_publish_gov_transaction"
	labelNameOutPullLatestOutputMetrics             = "out_pull_latest_output"
	labelNameOutPullTxInclusionStateMetrics         = "out_pull_tx_inclusion_state"
	labelNameOutPullOutputByIDMetrics               = "out_pull_output_by_id"
	labelTxPublishResult                            = "result"
	labelNameWebapiRequestOperation                 = "api_req_type"
	labelNameWebapiRequestStatusCode                = "api_req_status_code"
	labelNameWebapiEvmRPCSuccess                    = "success"
)

func getChainLabels(chainID isc.ChainID) prometheus.Labels {
	return prometheus.Labels{
		labelNameChain: chainID.String(),
	}
}
