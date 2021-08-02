package metrics

import "github.com/iotaledger/wasp/packages/iscp"

type MempoolMetrics interface {
	NewOffLedgerRequest()
	NewOnLedgerRequest()
	ProcessRequest()
	NewChainMetrics(chainID *iscp.ChainID) *Metrics
}

// type StateManagerMetrics interface {
// }
//
// type ConsensusMetrics interface {
// }

type ChainMetrics interface {
	MempoolMetrics
}
