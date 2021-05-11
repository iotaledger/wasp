// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/vm"
)

// Message types for the committee communications.
const (
	MsgNotifyRequests          = 0 + peering.FirstUserMsgCode
	MsgNotifyFinalResultPosted = 1 + peering.FirstUserMsgCode
	MsgStartProcessingRequest  = 2 + peering.FirstUserMsgCode
	MsgSignedHash              = 3 + peering.FirstUserMsgCode
	MsgGetBlock                = 4 + peering.FirstUserMsgCode
	MsgBlock                   = 5 + peering.FirstUserMsgCode
	MsgSignedResult            = 6 + peering.FirstUserMsgCode
)

type TimerTick int

// message is sent to the leader of the state processing
// it is sent upon state change or upon arrival of the new request
// the receiving operator will ignore repeating messages
type NotifyReqMsg struct {
	SenderIndex   uint16
	StateOutputID ledgerstate.OutputID
	RequestIDs    []coretypes.RequestID
}

// message is sent by the leader to all peers immediately after the final transaction is posted
// to the tangle. Main purpose of the message is to prevent unnecessary leader rotation
// in long confirmation times
// Final signature is sent to prevent possibility for a leader node to lie (is it necessary)
type NotifyFinalResultPostedMsg struct {
	SenderIndex   uint16
	StateOutputID ledgerstate.OutputID
	TxId          ledgerstate.TransactionID
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request batch and sign the result hash with the timestamp proposed by the leader
type StartProcessingBatchMsg struct {
	SenderIndex   uint16
	StateOutputID ledgerstate.OutputID
	// timestamp of the message. Field is set upon receive the message to sender's timestamp
	Timestamp int64
	// batch of request ids
	RequestIDs []coretypes.RequestID
	// reward address
	FeeDestination coretypes.AgentID
}

// after calculations the result peer responds to the start processing msg
// with SignedHashMsg, which contains result hash and signatures
type SignedHashMsg struct {
	SenderIndex   uint16
	StateOutputID ledgerstate.OutputID
	Timestamp     int64
	BatchHash     hashing.HashValue
	OrigTimestamp int64
	EssenceHash   hashing.HashValue
	SigShare      tbdn.SigShare
}

type SignedResultMsg struct {
	SenderIndex uint16
	EssenceHash hashing.HashValue
	SigShare    tbdn.SigShare
}

// GetBlockMsg StateManager queries specific block data from another peer (access node)
type GetBlockMsg struct {
	SenderNetID string
	BlockIndex  uint32
}

// BlockMsg StateManager in response to GetBlockMsg sends block data to the querying node's StateManager
type BlockMsg struct {
	SenderNetID string
	BlockBytes  []byte
}

// DismissChainMsg sent by component to the chain core in case of major setback
type DismissChainMsg struct {
	Reason string
}

// StateTransitionMsg StateManager -> ConsensusOld. Notifies consensus about changed state
type StateTransitionMsg struct {
	// new variable state
	State state.VirtualState
	// corresponding state transaction
	StateOutput *ledgerstate.AliasOutput
	//
	StateTimestamp time.Time
}

// StateCandidateMsg ConsensusOld -> StateManager. ConsensusOld sends the finalized next state to StateManager
type StateCandidateMsg struct {
	State             state.VirtualState
	ApprovingOutputID ledgerstate.OutputID
}

// VMResultMsg ConsensusOld -> ConsensusOld. VM sends result of async task started by ConsensusOld to itself
type VMResultMsg struct {
	Task   *vm.VMTask
	Leader uint16
}

// AsynchronousCommonSubsetMsg
type AsynchronousCommonSubsetMsg struct {
	ProposedBatchesBin [][]byte
}

// InclusionStateMsg nodeconn plugin sends inclusions state of the transaction to ConsensusOld
type InclusionStateMsg struct {
	TxID  ledgerstate.TransactionID
	State ledgerstate.InclusionState
}

// StateMsg nodeconn plugin sends the only existing AliasOutput in the chain's address to StateManager
type StateMsg struct {
	ChainOutput *ledgerstate.AliasOutput
	Timestamp   time.Time
}
