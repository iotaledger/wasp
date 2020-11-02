package chain

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/processors"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm"
)

type Chain interface {
	ID() *coretypes.ChainID
	OwnerAddress() *address.Address
	Color() *balance.Color
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	NumPeers() uint16
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToCommitteePeers(msgType byte, msgData []byte, ts int64) uint16
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
	InitTestRound()
	HasQuorum() bool
	PeerStatus() []*PeerStatus
	//
	SetReadyStateManager()
	SetReadyConsensus()
	Dismiss()
	IsDismissed() bool
	GetRequestProcessingStatus(*coretypes.RequestID) RequestProcessingStatus
	// chain processors
	Processors() *processors.ProcessorCache
}

type PeerStatus struct {
	Index     int
	PeeringID string
	IsSelf    bool
	Connected bool
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

type StateManager interface {
	EvidenceStateIndex(idx uint32)
	EventStateIndexPingPongMsg(msg *StateIndexPingPongMsg)
	EventGetBatchMsg(msg *GetBatchMsg)
	EventBatchHeaderMsg(msg *BatchHeaderMsg)
	EventStateUpdateMsg(msg *StateUpdateMsg)
	EventStateTransactionMsg(msg *StateTransactionMsg)
	EventPendingBatchMsg(msg PendingBatchMsg)
	EventTimerMsg(msg TimerTick)
}

type Operator interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventBalancesMsg(BalancesMsg)
	EventRequestMsg(*RequestMsg)
	EventNotifyReqMsg(*NotifyReqMsg)
	EventStartProcessingBatchMsg(*StartProcessingBatchMsg)
	EventResultCalculated(*vm.VMTask)
	EventSignedHashMsg(*SignedHashMsg)
	EventNotifyFinalResultPostedMsg(*NotifyFinalResultPostedMsg)
	EventTransactionInclusionLevelMsg(msg *TransactionInclusionLevelMsg)
	EventTimerMsg(TimerTick)
	//
	IsRequestInBacklog(*coretypes.RequestID) bool
}

var ConstructorNew func(bootupData *registry.BootupData, log *logger.Logger, onActivation func()) Chain

func New(bootupData *registry.BootupData, log *logger.Logger, onActivation func()) Chain {
	return ConstructorNew(bootupData, log, onActivation)
}
