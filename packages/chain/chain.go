// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"fmt"
	"time"

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
	ID() *coretypes.ChainID
	GetCommitteeInfo() *CommitteeInfo
	ReceiveMessage(interface{})
	Events() ChainEvents
	Processors() *processors.ProcessorCache
}

type Chain interface {
	ChainCore

	ReceiveTransaction(*ledgerstate.Transaction)
	ReceiveInclusionState(ledgerstate.TransactionID, ledgerstate.InclusionState)
	ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time)
	ReceiveOutput(output ledgerstate.Output)

	Dismiss(reason string)
	IsDismissed() bool

	ChainRequests
}

type ChainEvents interface {
	RequestProcessed() *events.Event
	StateTransition() *events.Event
	StateSynced() *events.Event
}

// Committee is ordered (indexed 0..size-1) list of peers which run the consensus and the whoel chain
type Committee interface {
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	DKShare() *tcrypto.DKShare
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte, ts int64) uint16
	IsAlivePeer(peerIndex uint16) bool
	QuorumIsAlive(quorum ...uint16) bool
	PeerStatus() []*PeerStatus
	OnPeerMessage(fun func(recv *peering.RecvEvent))
	IsReady() bool
	Close()
}

// TODO temporary wrapper for Committee need replacement for all peers, not only committee.
//  Must be close to GroupProvider but less functions
type PeerGroupProvider interface {
	NumPeers() uint16
	NumIsAlive(quorum uint16) bool
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendToAllUntilFirstError(msgType byte, msgData []byte) uint16
}

type ChainRequests interface {
	GetRequestProcessingStatus(id coretypes.RequestID) RequestProcessingStatus
	EventRequestProcessed() *events.Event
}

type NodeConnection interface {
	PullBacklog(addr *ledgerstate.AliasAddress)
	PullState(addr *ledgerstate.AliasAddress)
	PullConfirmedTransaction(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	PullTransactionInclusionState(addr ledgerstate.Address, txid ledgerstate.TransactionID)
	PullConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID)
	PostTransaction(tx *ledgerstate.Transaction, fromSc ledgerstate.Address, fromLeader uint16)
}

type StateManager interface {
	Ready() *ready.Ready
	SetPeers(PeerGroupProvider)
	EventGetBlockMsg(msg *GetBlockMsg)
	EventBlockMsg(msg *BlockMsg)
	EventStateMsg(msg *StateMsg)
	EventOutputMsg(msg ledgerstate.Output)
	EventStateCandidateMsg(msg StateCandidateMsg)
	EventTimerMsg(msg TimerTick)
	GetSyncInfo() *SyncInfo
	Close()
}

type Consensus interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventNotifyReqMsg(*NotifyReqMsg)
	EventStartProcessingBatchMsg(*StartProcessingBatchMsg)
	EventResultCalculated(msg *VMResultMsg)
	EventSignedHashMsg(*SignedHashMsg)
	EventNotifyFinalResultPostedMsg(*NotifyFinalResultPostedMsg)
	EventTransactionInclusionStateMsg(msg *InclusionStateMsg)
	EventTimerMsg(TimerTick)
	Close()
}

type Mempool interface {
	ReceiveRequest(req coretypes.Request)
	MarkSeenByCommitteePeer(reqid *coretypes.RequestID, peerIndex uint16)
	ClearSeenMarks()
	GetReadyList(seenThreshold uint16) []coretypes.Request
	GetReadyListFull(seenThreshold uint16) []*ReadyListRecord
	TakeAllReady(nowis time.Time, reqids ...coretypes.RequestID) ([]coretypes.Request, bool)
	RemoveRequests(reqs ...coretypes.RequestID)
	HasRequest(id coretypes.RequestID) bool
	Stats() (int, int, int)
	Close()
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
