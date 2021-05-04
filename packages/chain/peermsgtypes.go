// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto/tbdn"
	"github.com/iotaledger/wasp/packages/vm"
	"time"
)

// Message types for the committee communications.
const (
	MsgNotifyRequests          = 0 + peering.FirstUserMsgCode
	MsgNotifyFinalResultPosted = 1 + peering.FirstUserMsgCode
	MsgStartProcessingRequest  = 2 + peering.FirstUserMsgCode
	MsgSignedHash              = 3 + peering.FirstUserMsgCode
	MsgGetBlock                = 4 + peering.FirstUserMsgCode
	MsgBlock                   = 5 + peering.FirstUserMsgCode
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

// request block of updates from peer. Used in sync process
type GetBlockMsg struct {
	SenderIndex uint16
	BlockIndex  uint32
}

type BlockMsg struct {
	SenderIndex uint16
	BlockBytes  []byte
}

// DismissChainMsg sent by component to the chain core in case of major setback
type DismissChainMsg struct {
	Reason string
}

// state manager notifies consensus about changed state
// only sent internally within committee
// state transition is always from state N to state N+1
type StateTransitionMsg struct {
	// new variable state
	VariableState state.VirtualState
	// corresponding state transaction
	ChainOutput *ledgerstate.AliasOutput
	//
	Timestamp time.Time
}

// message of complete batch. Is sent by consensus operator to the state manager as a VM result
// - state manager to itself when batch is completed after syncing
type StateCandidateMsg struct {
	State state.VirtualState
}

// VMResultMsg is the message sent by the async VM task to the chan object upon success full finish
type VMResultMsg struct {
	Task   *vm.VMTask
	Leader uint16
}

type InclusionStateMsg struct {
	TxID  ledgerstate.TransactionID
	State ledgerstate.InclusionState
}

type StateMsg struct {
	ChainOutput *ledgerstate.AliasOutput
	Timestamp   time.Time
}
