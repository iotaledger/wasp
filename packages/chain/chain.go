// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/chain/mempool"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type ChainCore interface {
	ID() *iscp.ChainID
	GetCommitteeInfo() *CommitteeInfo
	StateCandidateToStateManager(state.VirtualStateAccess, iotago.OutputID)
	Events() ChainEvents
	Processors() *processors.Cache
	GlobalStateSync() coreutil.ChainStateSync
	GetStateReader() state.OptimisticStateReader
	Log() *logger.Logger

	// Most of these methods are made public for mocking in tests
	EnqueueDismissChain(reason string) // This one should really be public
	Enqueueiotago(chainOutput *iotago.AliasOutput, timestamp time.Time)
	EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn)
	EnqueueRequestAckMsg(msg *messages.RequestAckMsgIn)
	EnqueueMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn)
	EnqueueMissingRequestMsg(msg *messages.MissingRequestMsg)
	EnqueueTimerTick(tick int)
}

// ChainEntry interface to access chain from the chain registry side
type ChainEntry interface {
	ReceiveTransaction(*iotago.Transaction)
	ReceiveInclusionState(iotago.TransactionID, iotago.InclusionState)
	ReceiveState(stateOutput *iotago.AliasOutput, timestamp time.Time)
	ReceiveOutput(output iotago.Output)

	Dismiss(reason string)
	IsDismissed() bool
}

// ChainRequests is an interface to query status of the request
type ChainRequests interface {
	GetRequestProcessingStatus(id iscp.RequestID) RequestProcessingStatus
	EventRequestProcessed() *events.Event
}

type ChainEvents interface {
	RequestProcessed() *events.Event
	ChainTransition() *events.Event
}

type Chain interface {
	ChainCore
	ChainRequests
	ChainEntry
}

// Committee is ordered (indexed 0..size-1) list of peers which run the consensus
type Committee interface {
	Address() iotago.Address
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	DKShare() *tcrypto.DKShare
	IsAlivePeer(peerIndex uint16) bool
	QuorumIsAlive(quorum ...uint16) bool
	PeerStatus() []*PeerStatus
	IsReady() bool
	Close()
	RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte))
	GetOtherValidatorsPeerIDs() []string
	GetRandomValidators(upToN int) []string
}

type NodeConnection interface {
	PullBacklog(addr *iotago.AliasAddress)
	PullState(addr *iotago.AliasAddress)
	PullConfirmedTransaction(addr iotago.Address, txid iotago.TransactionID)
	PullTransactionInclusionState(addr iotago.Address, txid iotago.TransactionID)
	PullConfirmedOutput(addr iotago.Address, outputID iotago.OutputID)
	PostTransaction(tx *iotago.Transaction)
}

type StateManager interface {
	Ready() *ready.Ready
	EnqueueGetBlockMsg(msg *messages.GetBlockMsgIn)
	EnqueueBlockMsg(msg *messages.BlockMsgIn)
	EnqueueStateMsg(msg *messages.StateMsg)
	EnqueueOutputMsg(msg iotago.Output)
	EnqueueStateCandidateMsg(state.VirtualStateAccess, iotago.OutputID)
	EnqueueTimerMsg(msg messages.TimerTick)
	GetStatusSnapshot() *SyncInfo
	Close()
}

type Consensus interface {
	EnqueueStateTransitionMsg(state.VirtualStateAccess, *iotago.AliasOutput, time.Time)
	EnqueueSignedResultMsg(*messages.SignedResultMsgIn)
	EnqueueSignedResultAckMsg(*messages.SignedResultAckMsgIn)
	EnqueueInclusionsStateMsg(iotago.TransactionID, iotago.InclusionState)
	EnqueueAsynchronousCommonSubsetMsg(msg *messages.AsynchronousCommonSubsetMsg)
	EnqueueVMResultMsg(msg *messages.VMResultMsg)
	EnqueueTimerMsg(messages.TimerTick)
	IsReady() bool
	Close()
	GetStatusSnapshot() *ConsensusInfo
	ShouldReceiveMissingRequest(req iscp.Request) bool
}

type AsynchronousCommonSubsetRunner interface {
	RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte))
	Close()
}

type SyncInfo struct {
	Synced                bool
	SyncedBlockIndex      uint32
	SyncedStateHash       hashing.HashValue
	SyncedStateTimestamp  time.Time
	StateOutputBlockIndex uint32
	StateOutputID         iotago.OutputID
	StateOutputHash       hashing.HashValue
	StateOutputTimestamp  time.Time
}

type ConsensusInfo struct {
	StateIndex uint32
	Mempool    mempool.MempoolInfo
	TimerTick  int
}

type ReadyListRecord struct {
	Request iscp.Request
	Seen    map[uint16]bool
}

type CommitteeInfo struct {
	Address       iotago.Address
	Size          uint16
	Quorum        uint16
	QuorumIsAlive bool
	PeerStatus    []*PeerStatus
}

type PeerStatus struct {
	Index     int
	PeeringID string
	IsSelf    bool
	Connected bool
}

type ChainTransitionEventData struct {
	VirtualState    state.VirtualStateAccess
	ChainOutput     *iotago.AliasOutput
	OutputTimestamp time.Time
}

func (p *PeerStatus) String() string {
	return fmt.Sprintf("%+v", *p)
}

type RequestProcessingStatus int

const (
	RequestProcessingStatusUnknown = RequestProcessingStatus(iota)
	RequestProcessingStatusBacklog
	RequestProcessingStatusCompleted
)

const (
	// TimerTickPeriod time tick for consensus and state manager objects
	TimerTickPeriod = 100 * time.Millisecond
)

const (
	PeerMsgTypeMissingRequestIDs = iota
	PeerMsgTypeMissingRequest
	PeerMsgTypeOffLedgerRequest
	PeerMsgTypeRequestAck
)
