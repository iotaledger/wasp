// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
)

// Message types for the committee communications.
const (
	MsgGetBlock = 1 + peering.FirstUserMsgCode + iota
	MsgBlock
	MsgSignedResult
	MsgSignedResultAck
	MsgOffLedgerRequest
	MsgMissingRequestIDs
	MsgMissingRequest
	MsgRequestAck
)

type TimerTick int

// StateTransitionMsg Notifies chain about changed state
type StateTransitionMsg struct {
	// new variable state
	State state.VirtualStateAccess
	// corresponding state transaction
	StateOutput *ledgerstate.AliasOutput
	//
	StateTimestamp time.Time
}

// StateCandidateMsg Consensus sends the finalized next state to StateManager
type StateCandidateMsg struct {
	State             state.VirtualStateAccess
	ApprovingOutputID ledgerstate.OutputID
}

// VMResultMsg Consensus -> Consensus. VM sends result of async task started by Consensus to itself
type VMResultMsg struct {
	Task *vm.VMTask
}

// AsynchronousCommonSubsetMsg
type AsynchronousCommonSubsetMsg struct {
	ProposedBatchesBin [][]byte
	SessionID          uint64
}

// InclusionStateMsg txstream plugin sends inclusions state of the transaction to ConsensusOld
type InclusionStateMsg struct {
	TxID  ledgerstate.TransactionID
	State ledgerstate.InclusionState
}

// StateMsg txstream plugin sends the only existing AliasOutput in the chain's address to StateManager
type StateMsg struct {
	ChainOutput *ledgerstate.AliasOutput
	Timestamp   time.Time
}
