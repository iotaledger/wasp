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
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/metrics/nodeconnmetrics"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util/ready"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type ChainCore interface {
	ID() *iscp.ChainID
	GetCommitteeInfo() *CommitteeInfo
	StateCandidateToStateManager(state.VirtualStateAccess, iotago.OutputID)
	TriggerChainTransition(*ChainTransitionEventData)
	Processors() *processors.Cache
	GlobalStateSync() coreutil.ChainStateSync
	GetStateReader() state.OptimisticStateReader
	GetChainNodes() []peering.PeerStatusProvider     // CommitteeNodes + AccessNodes
	GetCandidateNodes() []*governance.AccessNodeInfo // All the current candidates.
	Log() *logger.Logger

	// FIXME these methods should not be part of the chain interface just for the need of mocking
	//  Mocking interfaces should be available only in the testing environment
	// Most of these methods are made public for mocking in tests
	EnqueueDismissChain(reason string) // This one should really be public
	EnqueueLedgerState(chainOutput *iotago.AliasOutput, timestamp time.Time)
	EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn)
	EnqueueRequestAckMsg(msg *messages.RequestAckMsgIn)
	EnqueueMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn)
	EnqueueMissingRequestMsg(msg *messages.MissingRequestMsg)
	EnqueueTimerTick(tick int)
}

// ChainEntry interface to access chain from the chain registry side
type ChainEntry interface {
	ReceiveTransaction(*iotago.Transaction)
	ReceiveState(stateOutput *iotago.AliasOutput, timestamp time.Time)
	Dismiss(reason string)
	IsDismissed() bool
}

// ChainRequests is an interface to query status of the request
type ChainRequests interface {
	GetRequestProcessingStatus(id iscp.RequestID) RequestProcessingStatus
	AttachToRequestProcessed(func(iscp.RequestID)) (attachID *events.Closure)
	DetachFromRequestProcessed(attachID *events.Closure)
}

type ChainMetrics interface {
	GetNodeConnectionMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics
	GetConsensusWorkflowStatus() ConsensusWorkflowStatus
}

type Chain interface {
	ChainCore
	ChainRequests
	ChainEntry
	ChainMetrics
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
	GetRandomValidators(upToN int) []*cryptolib.PublicKey // TODO: Remove after OffLedgerRequest dissemination is changed.
}

type (
	NodeConnectionHandleTransactionFun func(*iotago.Transaction)
	//NodeConnectionHandleInclusionStateFun     func(iotago.TransactionID, iotago.InclusionState) TODO: refactor
	NodeConnectionHandleOutputFun             func(iotago.Output, *iotago.UTXOInput)
	NodeConnectionHandleUnspentAliasOutputFun func(*iscp.AliasOutputWithID, time.Time)
)

type NodeConnection interface {
	Subscribe(addr iotago.Address)
	Unsubscribe(addr iotago.Address)
	AttachToTransactionReceived(*iotago.AliasAddress, NodeConnectionHandleTransactionFun)
	//AttachToInclusionStateReceived(*iotago.AliasAddress, NodeConnectionHandleInclusionStateFun) TODO: refactor
	AttachToOutputReceived(*iotago.AliasAddress, NodeConnectionHandleOutputFun)
	AttachToUnspentAliasOutputReceived(*iotago.AliasAddress, NodeConnectionHandleUnspentAliasOutputFun)
	PullState(addr *iotago.AliasAddress)
	PullTransactionInclusionState(addr iotago.Address, txid iotago.TransactionID)
	PullConfirmedOutput(addr iotago.Address, outputID *iotago.UTXOInput)
	PostTransaction(tx *iotago.Transaction)
	GetMetrics() nodeconnmetrics.NodeConnectionMetrics
	DetachFromTransactionReceived(*iotago.AliasAddress)
	DetachFromInclusionStateReceived(*iotago.AliasAddress)
	DetachFromOutputReceived(*iotago.AliasAddress)
	DetachFromUnspentAliasOutputReceived(*iotago.AliasAddress)
	Close()
}

type ChainNodeConnection interface {
	AttachToTransactionReceived(NodeConnectionHandleTransactionFun)
	//AttachToInclusionStateReceived(NodeConnectionHandleInclusionStateFun)	TODO: refactor
	AttachToOutputReceived(NodeConnectionHandleOutputFun)
	AttachToUnspentAliasOutputReceived(NodeConnectionHandleUnspentAliasOutputFun)
	PullState()
	PullTransactionInclusionState(txid iotago.TransactionID)
	PullConfirmedOutput(outputID *iotago.UTXOInput)
	PostTransaction(tx *iotago.Transaction)
	GetMetrics() nodeconnmetrics.NodeConnectionMessagesMetrics
	DetachFromTransactionReceived()
	DetachFromInclusionStateReceived()
	DetachFromOutputReceived()
	DetachFromUnspentAliasOutputReceived()
	Close()
}

type StateManager interface {
	Ready() *ready.Ready
	EnqueueGetBlockMsg(msg *messages.GetBlockMsgIn)
	EnqueueBlockMsg(msg *messages.BlockMsgIn)
	EnqueueStateMsg(msg *messages.StateMsg)
	EnqueueOutputMsg(iotago.Output, *iotago.UTXOInput)
	EnqueueStateCandidateMsg(state.VirtualStateAccess, *iotago.UTXOInput)
	EnqueueTimerMsg(msg messages.TimerTick)
	GetStatusSnapshot() *SyncInfo
	SetChainPeers(peers []*cryptolib.PublicKey)
	Close()
}

type Consensus interface {
	EnqueueStateTransitionMsg(state.VirtualStateAccess, *iotago.AliasOutput, time.Time)
	EnqueueSignedResultMsg(*messages.SignedResultMsgIn)
	EnqueueSignedResultAckMsg(*messages.SignedResultAckMsgIn)
	// EnqueueInclusionsStateMsg(iotago.TransactionID, iotago.InclusionState) // TODO does this make sense with hornet?
	EnqueueAsynchronousCommonSubsetMsg(msg *messages.AsynchronousCommonSubsetMsg)
	EnqueueVMResultMsg(msg *messages.VMResultMsg)
	EnqueueTimerMsg(messages.TimerTick)
	IsReady() bool
	Close()
	GetStatusSnapshot() *ConsensusInfo
	GetWorkflowStatus() ConsensusWorkflowStatus
	ShouldReceiveMissingRequest(req iscp.Calldata) bool
}

type AsynchronousCommonSubsetRunner interface {
	RunACSConsensus(value []byte, sessionID uint64, stateIndex uint32, callback func(sessionID uint64, acs [][]byte))
	Close()
}

type WAL interface {
	Write(bytes []byte) error
	Contains(i uint32) bool
	Read(i uint32) ([]byte, error)
}

type SyncInfo struct {
	Synced                bool
	SyncedBlockIndex      uint32
	SyncedStateHash       hashing.HashValue
	SyncedStateTimestamp  time.Time
	StateOutputBlockIndex uint32
	StateOutputID         *iotago.UTXOInput
	StateOutputCommitment hashing.HashValue
	StateOutputTimestamp  time.Time
}

type ConsensusInfo struct {
	StateIndex uint32
	Mempool    mempool.MempoolInfo
	TimerTick  int
}

type ConsensusWorkflowStatus interface {
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

type ReadyListRecord struct {
	Request iscp.Calldata
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
	PubKey    *cryptolib.PublicKey
	NetID     string
	Connected bool
}

type ChainTransitionEventData struct {
	VirtualState    state.VirtualStateAccess
	ChainOutput     *iscp.AliasOutputWithID
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
