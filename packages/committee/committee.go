package committee

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm"
	"time"
)

type Committee interface {
	Address() address.Address
	Color() balance.Color
	Size() uint16
	OwnPeerIndex() uint16
	MetaData() *registry.SCMetaData
	OpenQueue()
	Dismiss()
	IsDismissed() bool
	SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error
	SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time)
	SendMsgInSequence(msgType byte, msgData []byte, seqIndex uint16, seq []uint16) (uint16, error)
	IsAlivePeer(peerIndex uint16) bool
	ReceiveMessage(msg interface{})
	InitTestRound()
}

type StateManager interface {
	CheckSynchronizationStatus(idx uint32) bool
	EventGetBatchMsg(msg *GetBatchMsg)
	EventBatchHeaderMsg(msg *BatchHeaderMsg)
	EventStateUpdateMsg(msg *StateUpdateMsg)
	EventStateTransactionMsg(msg StateTransactionMsg)
	EventTimerMsg(msg TimerTick)
}

type Operator interface {
	EventStateTransitionMsg(msg *StateTransitionMsg)
	EventBalancesMsg(balances BalancesMsg)
	EventRequestMsg(reqMsg *RequestMsg)
	EventNotifyReqMsg(msg *NotifyReqMsg)
	EventStartProcessingReqMsg(msg *StartProcessingReqMsg)
	EventResultCalculated(result *vm.VMTask)
	EventSignedHashMsg(msg *SignedHashMsg)
	EventTimerMsg(msg TimerTick)
}

var ConstructorNew func(scdata *registry.SCMetaData, log *logger.Logger) (Committee, error)

func New(scdata *registry.SCMetaData, log *logger.Logger) (Committee, error) {
	return ConstructorNew(scdata, log)
}
