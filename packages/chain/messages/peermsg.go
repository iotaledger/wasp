// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package messages

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
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

// StateTransitionMsg Notifies chain about changed
type StateTransitionMsg struct {
	// is transition a governance update
	IsGovernance bool
	// new variable state
	State state.VirtualStateAccess
	// corresponding state transaction
	StateOutput *isc.AliasOutputWithID
	//
	StateTimestamp time.Time
}

// StateCandidateMsg Consensus sends the finalized next state to StateManager
type StateCandidateMsg struct {
	State             state.VirtualStateAccess
	ApprovingOutputID *iotago.UTXOInput
}

// Level 1 sends new state output to state manager
type OutputMsg struct {
	Output iotago.Output
	ID     *iotago.UTXOInput
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
type TxInclusionStateMsg struct {
	TxID  iotago.TransactionID
	State string
}

// StateMsg txstream plugin sends the only existing AliasOutput in the chain's address to StateManager
type StateMsg struct {
	ChainOutput *isc.AliasOutputWithID
	Timestamp   time.Time
}
