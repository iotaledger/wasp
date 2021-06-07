// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"

	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type ChainCore interface {
	ID() *chainid.ChainID
	GetCommitteeInfo() *CommitteeInfo
	ReceiveMessage(interface{})
	Events() ChainEvents
	Processors() *processors.ProcessorCache
	GlobalStateSync() coreutil.GlobalSync
	GetStateReader() state.OptimisticStateReader
	Log() *logger.Logger
}

// ChainEntry interface to access chain from the chain registry side
type ChainEntry interface {
	ReceiveTransaction(*ledgerstate.Transaction)
	ReceiveInclusionState(ledgerstate.TransactionID, ledgerstate.InclusionState)
	ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time)
	ReceiveOutput(output ledgerstate.Output)

	Dismiss(reason string)
	IsDismissed() bool
}

// ChainRequests is an interface to query status of the reqest
type ChainRequests interface {
	GetRequestProcessingStatus(id coretypes.RequestID) RequestProcessingStatus
	EventRequestProcessed() *events.Event
}

type ChainEvents interface {
	RequestProcessed() *events.Event
	StateTransition() *events.Event
	StateSynced() *events.Event
}

type Chain interface {
	ChainCore
	ChainRequests
	ChainEntry
}

// Committee is ordered (indexed 0..size-1) list of peers which run the consensus
type Committee interface {
	Address() ledgerstate.Address
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	DKShare() *tcrypto.DKShare
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte, ts int64)
	IsAlivePeer(peerIndex uint16) bool
	QuorumIsAlive(quorum ...uint16) bool
	PeerStatus() []*PeerStatus
	Attach(chain ChainCore)
	IsReady() bool
	Close()
	RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte))
}

type NodeConnection interface {
	PullBacklog(addr *ledgerstate.AliasAddress)
	PullState(addr *ledgerstate.AliasAddress)
	PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID)
	PostTransaction(tx *ledgerstate.Transaction)
}

type StateManager interface {
	Ready() *ready.Ready
	EventGetBlockMsg(msg *GetBlockMsg)
	EventBlockMsg(msg *BlockMsg)
	EventStateMsg(msg *StateMsg)
	EventOutputMsg(msg ledgerstate.Output)
	EventStateCandidateMsg(msg *StateCandidateMsg)
	EventTimerMsg(msg TimerTick)
	GetStatusSnapshot() *SyncInfo
	Close()
}

type Consensus interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventVMResultCalculated(*VMResultMsg)
	EventSignedResultMsg(*SignedResultMsg)
	EventInclusionsStateMsg(*InclusionStateMsg)
	EventAsynchronousCommonSubsetMsg(msg *AsynchronousCommonSubsetMsg)
	EventVMResultMsg(msg *VMResultMsg)
	EventTimerMsg(TimerTick)
	IsReady() bool
	Close()
	GetStatusSnapshot() *ConsensusInfo
}

type Mempool interface {
	ReceiveRequests(reqs ...coretypes.Request)
	RemoveRequests(reqs ...coretypes.RequestID)
	ReadyNow(nowis ...time.Time) []coretypes.Request
	ReadyFromIDs(nowis time.Time, reqids ...coretypes.RequestID) ([]coretypes.Request, bool)
	HasRequest(id coretypes.RequestID) bool
	Stats() MempoolStats
	WaitRequestInPool(reqid coretypes.RequestID, timeout ...time.Duration) bool // for testing
	WaitInBufferEmpty(timeout ...time.Duration) bool                            // for testing
	Close()
}

type AsynchronousCommonSubsetRunner interface {
	RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte))
	TryHandleMessage(recv *peering.RecvEvent) bool
	Close()
}

type MempoolStats struct {
	TotalPool      int
	Ready          int
	InBufCounter   int
	OutBufCounter  int
	InPoolCounter  int
	OutPoolCounter int
}

type SyncInfo struct {
	Synced                bool
	SyncedBlockIndex      uint32
	SyncedStateHash       hashing.HashValue
	SyncedStateTimestamp  time.Time
	StateOutputBlockIndex uint32
	StateOutputID         ledgerstate.OutputID
	StateOutputHash       hashing.HashValue
	StateOutputTimestamp  time.Time
}

type ConsensusInfo struct {
	StateIndex uint32
	Mempool    MempoolStats
	TimerTick  int
}

type ReadyListRecord struct {
	Request coretypes.Request
	Seen    map[uint16]bool
}

type CommitteeInfo struct {
	Address       ledgerstate.Address
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

type StateTransitionEventData struct {
	VirtualState    state.VirtualState
	ChainOutput     *ledgerstate.AliasOutput
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
