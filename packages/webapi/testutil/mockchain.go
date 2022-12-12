package testutil

import (
	"context"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type MockChain struct{}

var _ chain.Chain = &MockChain{}

// LatestAliasOutput implements chain.Chain
func (*MockChain) LatestAliasOutput() (confirmed *isc.AliasOutputWithID, active *isc.AliasOutputWithID) {
	panic("unimplemented")
}

// GetCandidateNodes implements chain.Chain
func (*MockChain) GetCandidateNodes() []*governance.AccessNodeInfo {
	panic("unimplemented")
}

// GetChainNodes implements chain.Chain
func (*MockChain) GetChainNodes() []peering.PeerStatusProvider {
	panic("unimplemented")
}

// GetCommitteeInfo implements chain.Chain
func (*MockChain) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("unimplemented")
}

// GetStateReader implements chain.Chain
func (*MockChain) GetStateReader() state.Store {
	panic("unimplemented")
}

// ID implements chain.Chain
func (*MockChain) ID() isc.ChainID {
	panic("unimplemented")
}

// Log implements chain.Chain
func (*MockChain) Log() *logger.Logger {
	return testlogger.NewSilentLogger("", true)
}

// Processors implements chain.Chain
func (*MockChain) Processors() *processors.Cache {
	return new(processors.Cache)
}

// AwaitRequestProcessed implements chain.Chain
func (*MockChain) AwaitRequestProcessed(ctx context.Context, requestID isc.RequestID) <-chan *blocklog.RequestReceipt {
	panic("unimplemented")
}

// ReceiveOffLedgerRequest implements chain.Chain
func (*MockChain) ReceiveOffLedgerRequest(request isc.OffLedgerRequest, sender *cryptolib.PublicKey) {
}

// ConfigUpdated implements chain.Chain
func (*MockChain) ConfigUpdated(accessNodes []*cryptolib.PublicKey) {
	panic("unimplemented")
}

// GetConsensusPipeMetrics implements chain.Chain
func (*MockChain) GetConsensusPipeMetrics() chain.ConsensusPipeMetrics {
	panic("unimplemented")
}

// GetConsensusWorkflowStatus implements chain.Chain
func (*MockChain) GetConsensusWorkflowStatus() chain.ConsensusWorkflowStatus {
	panic("unimplemented")
}

// GetNodeConnectionMetrics implements chain.Chain
func (*MockChain) GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMetrics {
	panic("unimplemented")
}
