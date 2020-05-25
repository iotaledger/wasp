package committee

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
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
	EventResultCalculated(result *VMOutput)
	EventSignedHashMsg(msg *SignedHashMsg)
	EventTimerMsg(msg TimerTick)
}

type Processor interface {
	Run(inputs VMInputs) VMOutput
}

type VMInputs interface {
	// account address
	Address() *address.Address
	// color of the smart contracts
	Color() *balance.Color
	// balances/outputs of the account address (scid.Address())
	// imposed by the leader
	Balances() map[valuetransaction.ID][]*balance.Balance
	// reward address or nil of no rewards collected
	RewardAddress() *address.Address
	// timestamp imposed by the leader
	Timestamp() time.Time
	// batch of requests to run sequentially. .
	RequestMsg() []*RequestMsg
	// the context state transaction
	StateTransaction() *sctransaction.Transaction
	// the context variable state
	VariableState() state.VariableState
	// log for VM
	Log() *logger.Logger
}

type VMOutput struct {
	// references to inouts
	Inputs VMInputs
	// result transaction (not signed)
	// accumulated final result of batch processing. It means the result transaction as inputs
	// has all outputs to the SC account address from all request
	// similarly outputs are consolidated, for example it should contain the only output of the SC token
	ResultTransaction *sctransaction.Transaction
	// state update corresponding to requests
	StateUpdates []state.StateUpdate
}

var ConstructorNew func(scdata *registry.SCMetaData, log *logger.Logger) (Committee, error)

func New(scdata *registry.SCMetaData, log *logger.Logger) (Committee, error) {
	return ConstructorNew(scdata, log)
}

func BatchHash(vimp VMInputs) hashing.HashValue {
	var buf bytes.Buffer
	for _, msg := range vimp.RequestMsg() {
		buf.Write(msg.RequestId()[:])
	}
	return *hashing.HashData(buf.Bytes())
}
