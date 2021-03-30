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
	MsgStateIndexPingPong      = 0 + peering.FirstUserMsgCode
	MsgNotifyRequests          = 1 + peering.FirstUserMsgCode
	MsgNotifyFinalResultPosted = 2 + peering.FirstUserMsgCode
	MsgStartProcessingRequest  = 3 + peering.FirstUserMsgCode
	MsgSignedHash              = 4 + peering.FirstUserMsgCode
	MsgGetBatch                = 5 + peering.FirstUserMsgCode
	MsgStateUpdate             = 6 + peering.FirstUserMsgCode
	MsgBatchHeader             = 7 + peering.FirstUserMsgCode
)

type TimerTick int

// all peer messages have this
type PeerMsgHeader struct {
	// is set upon receive the message
	SenderIndex uint16
	// state index in the context of which the message is sent
	BlockIndex uint32
}

// Ping is sent to receive Pong
type StateIndexPingPongMsg struct {
	PeerMsgHeader
	RSVP bool
}

// message is sent to the leader of the state processing
// it is sent upon state change or upon arrival of the new request
// the receiving operator will ignore repeating messages
type NotifyReqMsg struct {
	PeerMsgHeader
	// list of request ids ordered by the time of arrival
	RequestIDs []coretypes.RequestID
}

// message is sent by the leader to all peers immediately after the final transaction is posted
// to the tangle. Main purpose of the message is to prevent unnecessary leader rotation
// in long confirmation times
// Final signature is sent to prevent possibility for a leader node to lie (is it necessary)
type NotifyFinalResultPostedMsg struct {
	PeerMsgHeader
	TxId ledgerstate.TransactionID
}

// message is sent by the leader to other peers to initiate request processing
// other peers are expected to check is timestamp is acceptable then
// process request batch and sign the result hash with the timestamp proposed by the leader
type StartProcessingBatchMsg struct {
	PeerMsgHeader
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
	PeerMsgHeader
	// timestamp of this message. Field is set upon receive the message to sender's timestamp
	Timestamp int64
	// returns hash of all req ids
	BatchHash hashing.HashValue
	// original timestamp, the parameter for calculations, which is signed as part of the essence
	OrigTimestamp int64
	// hash of the signed data (essence)
	EssenceHash hashing.HashValue
	// signature
	SigShare tbdn.SigShare
}

// request block of updates from peer. Used in syn process
type GetBlockMsg struct {
	PeerMsgHeader
}

// the header of the block message sent by peers in the process of syncing
// it is sent as a first message while syncing a batch
type BlockHeaderMsg struct {
	PeerMsgHeader
	Size                uint16
	AnchorTransactionID ledgerstate.TransactionID
}

// state update sent to peer. Used in sync process, as part of batch
type StateUpdateMsg struct {
	PeerMsgHeader
	// state update
	StateUpdate state.StateUpdate
	// position in a batch
	IndexInTheBlock uint16
}

// state manager notifies consensus operator about changed state
// only sent internally within committee
// state transition is always from state N to state N+1
type StateTransitionMsg struct {
	// new variable state
	VariableState state.VirtualState
	// corresponding state transaction
	ChainOutput *ledgerstate.AliasOutput
	//
	Timestamp time.Time
	// processed requests
	RequestIDs []*coretypes.RequestID
	// is the state index last seen
	Synchronized bool
}

// message of complete batch. Is sent by consensus operator to the state manager as a VM result
// - state manager to itself when batch is completed after syncing
type PendingBlockMsg struct {
	Block state.Block
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
