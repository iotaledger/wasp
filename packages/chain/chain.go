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

type NodeConnection interface {
	ChainNodeConn
	GetMetrics() nodeconnmetrics.NodeConnectionMetrics
}

type ChainReader interface {
	GetStateReader() state.Store // TODO: Rename to GetStore.
}

type ChainCore interface {
	ChainReader
	ID() *isc.ChainID
	GetCommitteeInfo() *CommitteeInfo // TODO: Review, maybe we can reorganize the CommitteeInfo structure.
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
