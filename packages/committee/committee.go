package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm"
)

type Committee interface {
	Address() *address.Address
	OwnerAddress() *address.Address
	Color() *balance.Color
	Size() uint16
	Quorum() uint16
	OwnPeerIndex() uint16
	NumPeers() uint16
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToCommitteePeers(msgType byte, msgData []byte) (uint16, int64)
	SendMsgInSequence(msgType byte, msgData []byte, seqIndex uint16, seq []uint16) (uint16, error)
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
	InitTestRound()
	HasQuorum() bool
	ConnectedPeers() []uint16
	//
	SetReadyStateManager()
	SetReadyConsensus()
	Dismiss()
	IsDismissed() bool
	GetRequestProcessingStatus(*sctransaction.RequestId) RequestProcessingStatus
}

type RequestProcessingStatus int

const (
	RequestProcessingStatusUnknown = RequestProcessingStatus(iota)
	RequestProcessingStatusBacklog
	RequestProcessingStatusCompleted
)

type StateManager interface {
	EvidenceStateIndex(idx uint32)
	EventPingPongMsg(msg *PingPongMsg)
	EventGetBatchMsg(msg *GetBatchMsg)
	EventBatchHeaderMsg(msg *BatchHeaderMsg)
	EventStateUpdateMsg(msg *StateUpdateMsg)
	EventStateTransactionMsg(msg *StateTransactionMsg)
	EventPendingBatchMsg(msg PendingBatchMsg)
	EventTimerMsg(msg TimerTick)
}

type Operator interface {
	EventProcessorReady(ProcessorIsReady)
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
	IsRequestInBacklog(*sctransaction.RequestId) bool
}

var ConstructorNew func(bootupData *registry.BootupData, log *logger.Logger, onActivation func()) Committee

func New(bootupData *registry.BootupData, log *logger.Logger, onActivation func()) Committee {
	return ConstructorNew(bootupData, log, onActivation)
}
