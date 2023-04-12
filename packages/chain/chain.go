// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type NodeConnection interface {
	ChainNodeConn
	Run(ctx context.Context) error
	WaitUntilInitiallySynced(context.Context) error
	GetBech32HRP() iotago.NetworkPrefix
	GetL1Params() *parameters.L1Params
	GetL1ProtocolParams() *iotago.ProtocolParameters
}

type StateFreshness byte

const (
	ActiveOrCommittedState StateFreshness = iota // ActiveState, if exist; Confirmed state otherwise.
	ActiveState                                  // The state the chain build next TX on, can be ahead of ConfirmedState.
	ConfirmedState                               // The state confirmed on L1.
)

func (sf StateFreshness) String() string {
	switch sf {
	case ActiveOrCommittedState:
		return "ActiveOrCommittedState"
	case ActiveState:
		return "ActiveState"
	case ConfirmedState:
		return "ConfirmedState"
	default:
		return fmt.Sprintf("StateFreshness=%v", int(sf))
	}
}

type ChainCore interface {
	ID() isc.ChainID
	// Returns the current latest confirmed alias output and the active one.
	// The active AO can be ahead of the confirmed one by several blocks.
	// Both values can be nil, if the node haven't received an output from
	// L1 yet (after a restart or a chain activation).
	LatestAliasOutput(freshness StateFreshness) (*isc.AliasOutputWithID, error)
	LatestState(freshness StateFreshness) (state.State, error)
	GetCommitteeInfo() *CommitteeInfo // TODO: Review, maybe we can reorganize the CommitteeInfo structure.
	Store() indexedstore.IndexedStore // Use LatestState whenever possible. That will work faster.
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
