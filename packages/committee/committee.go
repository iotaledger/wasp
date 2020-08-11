package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm"
)

type Committee interface {
	Params() *Parameters
	Address() *address.Address
	OwnerAddress() *address.Address
	Color() *balance.Color
	Size() uint16
	OwnPeerIndex() uint16
	NumPeers() uint16
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToCommitteePeers(msgType byte, msgData []byte) (uint16, int64)
	SendMsgInSequence(msgType byte, msgData []byte, seqIndex uint16, seq []uint16) (uint16, error)
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
	InitTestRound()
	//
	SetReadyStateManager()
	SetReadyConsensus()
	Dismiss()
	IsDismissed() bool
}

type StateManager interface {
	EvidenceStateIndex(idx uint32)
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
	EventStateTransactionEvidenced(msg *StateTransactionEvidenced)
	EventTimerMsg(TimerTick)
}

var ConstructorNew func(bootupData *registry.BootupData, log *logger.Logger, params *Parameters) Committee

func New(bootupData *registry.BootupData, log *logger.Logger, params *Parameters) Committee {
	return ConstructorNew(bootupData, log, params)
}
