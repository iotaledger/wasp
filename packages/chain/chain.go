// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"time"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type NodeConnection interface {
	ChainNodeConn
	GetMetrics() nodeconnmetrics.NodeConnectionMetrics
	Run(ctx context.Context)
}

type ChainCore interface {
	ID() isc.ChainID
	// Returns the current latest confirmed alias output and the active one.
	// The active AO can be ahead of the confirmed one by several blocks.
	// Both values can be nil, if the node haven't received an output from
	// L1 yet (after a restart or a chain activation).
	LatestAliasOutput() (confirmed, active *isc.AliasOutputWithID)
	GetCommitteeInfo() *CommitteeInfo // TODO: Review, maybe we can reorganize the CommitteeInfo structure.
	GetStateReader() state.Store      // TODO: Rename to GetStore.
	Processors() *processors.Cache
	GetChainNodes() []peering.PeerStatusProvider     // CommitteeNodes + AccessNodes
	GetCandidateNodes() []*governance.AccessNodeInfo // All the current candidates.
	Log() *logger.Logger
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
