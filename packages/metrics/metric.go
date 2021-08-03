package metrics

import "github.com/iotaledger/wasp/packages/iscp"

type GlobalMetrics interface {
	NewChainMetrics(chainID *iscp.ChainID) ChainMetrics
}

type MempoolMetrics interface {
	NewOffLedgerRequest()
	NewOnLedgerRequest()
	ProcessRequest()
}

// type StateManagerMetrics interface {
// }
//
// type ConsensusMetrics interface {
// }

type ChainMetrics interface {
	MempoolMetrics
}
