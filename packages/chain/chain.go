// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type ChainInfo struct { // TODO: ...
	ChainID *isc.ChainID
}

// type Chain interface { // TODO: ...
// 	ChainNode
// 	// TODO: Info() *ChainInfo
// 	// TODO: OffLedgerRequest(req ...)
// 	// TODO: GetCurrentCommittee.
// 	// TODO: GetCurrentAccessNodes.
// }

type NodeConnection interface {
	// TODO: node.ChainNodeConn
	GetMetrics() nodeconnmetrics.NodeConnectionMetrics
}

// func New(
// 	ctx context.Context,
// 	chainID *isc.ChainID,
// 	chainStore state.Store,
// 	nodeConn node.ChainNodeConn,
// 	nodeIdentity *cryptolib.KeyPair,
// 	processorsConfig *processors.Config,
// 	net peering.NetworkProvider,
// 	log *logger.Logger,
// ) (Chain, error) {
// 	var dkRegistry registry.DKShareRegistryProvider // TODO: Get it somehow.
// 	var cmtLogStore cmtLog.Store                   // TODO: Get it somehow.
// 	var smBlockWAL smGPAUtils.BlockWAL             // TODO: Get it somehow.
// 	return node.New(ctx, chainID, chainStore, nodeConn, nodeIdentity, processorsConfig, dkRegistry, cmtLogStore, smBlockWAL, net, log)
// }

type ChainCore interface {
	ID() *isc.ChainID
	GetCommitteeInfo() *CommitteeInfo
	// StateCandidateToStateManager(state.VirtualStateAccess, *iotago.UTXOInput)
	// TriggerChainTransition(*ChainTransitionEventData)
	Processors() *processors.Cache
	// GlobalStateSync() coreutil.ChainStateSync
	GetStateReader() state.Store                     // TODO: Rename to GetStore.
	GetChainNodes() []peering.PeerStatusProvider     // CommitteeNodes + AccessNodes
	GetCandidateNodes() []*governance.AccessNodeInfo // All the current candidates.
	Log() *logger.Logger
	// EnqueueDismissChain(reason string)
	// EnqueueAliasOutput(*isc.AliasOutputWithID)
}

type ConsensusPipeMetrics interface { // TODO: Review it.
	GetEventStateTransitionMsgPipeSize() int
	GetEventPeerLogIndexMsgPipeSize() int
	GetEventACSMsgPipeSize() int
	GetEventVMResultMsgPipeSize() int
	GetEventTimerMsgPipeSize() int
}

type ConsensusWorkflowStatus interface { // TODO: Review it.
	IsStateReceived() bool
	IsBatchProposalSent() bool
	IsConsensusBatchKnown() bool
	IsVMStarted() bool
	IsVMResultSigned() bool
	IsTransactionFinalized() bool
	IsTransactionPosted() bool
	IsTransactionSeen() bool
	IsInProgress() bool
	GetBatchProposalSentTime() time.Time
	GetConsensusBatchKnownTime() time.Time
	GetVMStartedTime() time.Time
	GetVMResultSignedTime() time.Time
	GetTransactionFinalizedTime() time.Time
	GetTransactionPostedTime() time.Time
	GetTransactionSeenTime() time.Time
	GetCompletedTime() time.Time
	GetCurrentStateIndex() uint32
}
