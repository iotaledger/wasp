package testchain

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"go.uber.org/atomic"
)

type MockedChainCore struct {
	T                                    *testing.T
	chainID                              *iscp.ChainID
	processors                           *processors.Cache
	eventStateTransition                 *events.Event
	eventRequestProcessed                *events.Event
	onGlobalStateSync                    func() coreutil.ChainStateSync
	onGetStateReader                     func() state.OptimisticStateReader
	onEventStateTransition               func(data *chain.ChainTransitionEventData)
	onEventRequestProcessed              func(id iscp.RequestID)
	onReceivePeerMessage                 func(*peering.PeerMessage)
	onReceiveDismissChainMsg             func(*messages.DismissChainMsg)
	onReceiveStateTransitionMsg          func(*messages.StateTransitionMsg)
	onReceiveStateCandidateMsg           func(*messages.StateCandidateMsg)
	onReceiveInclusionStateMsg           func(*messages.InclusionStateMsg)
	onReceiveStateMsg                    func(*messages.StateMsg)
	onReceiveVMResultMsg                 func(*messages.VMResultMsg)
	onReceiveAsynchronousCommonSubsetMsg func(*messages.AsynchronousCommonSubsetMsg)
	onReceiveTimerTick                   func(messages.TimerTick)
	onSync                               func(out ledgerstate.OutputID, blockIndex uint32) //nolint:structcheck,unused
	log                                  *logger.Logger
}

func NewMockedChainCore(t *testing.T, chainID *iscp.ChainID, log *logger.Logger) *MockedChainCore {
	receiveFailFun := func(typee string, msg interface{}) {
		t.Fatalf("Receiving of %s is not implemented, but %v is received", typee, msg)
	}
	ret := &MockedChainCore{
		T:          t,
		chainID:    chainID,
		processors: processors.MustNew(processors.NewConfig(inccounter.Processor)),
		log:        log,
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ iscp.RequestID))(params[0].(iscp.RequestID))
		}),
		onEventRequestProcessed: func(id iscp.RequestID) {
			log.Infof("onEventRequestProcessed: %s", id)
		},
		onReceivePeerMessage:        func(msg *peering.PeerMessage) { receiveFailFun("*peering.PeerMessage", msg) },
		onReceiveDismissChainMsg:    func(msg *messages.DismissChainMsg) { receiveFailFun("*messages.DismissChainMs", msg) },
		onReceiveStateTransitionMsg: func(msg *messages.StateTransitionMsg) { receiveFailFun("*messages.StateTransitionMsg", msg) },
		onReceiveStateCandidateMsg:  func(msg *messages.StateCandidateMsg) { receiveFailFun("*messages.StateCandidateMsg", msg) },
		onReceiveInclusionStateMsg:  func(msg *messages.InclusionStateMsg) { receiveFailFun("*messages.InclusionStateMsg", msg) },
		onReceiveStateMsg:           func(msg *messages.StateMsg) { receiveFailFun("*messages.StateMsg", msg) },
		onReceiveVMResultMsg:        func(msg *messages.VMResultMsg) { receiveFailFun("*messages.VMResultMsg", msg) },
		onReceiveAsynchronousCommonSubsetMsg: func(msg *messages.AsynchronousCommonSubsetMsg) {
			receiveFailFun("*messages.AsynchronousCommonSubsetMsg", msg)
		},
		onReceiveTimerTick: func(msg messages.TimerTick) { receiveFailFun("messages.TimerTick", msg) },
	}
	ret.onEventStateTransition = func(msg *chain.ChainTransitionEventData) {
		chain.LogStateTransition(msg, nil, log)
	}
	ret.eventStateTransition.Attach(events.NewClosure(func(data *chain.ChainTransitionEventData) {
		ret.onEventStateTransition(data)
	}))
	ret.eventRequestProcessed.Attach(events.NewClosure(func(id iscp.RequestID) {
		ret.onEventRequestProcessed(id)
	}))
	return ret
}

func (m *MockedChainCore) Log() *logger.Logger {
	return m.log
}

func (m *MockedChainCore) ID() *iscp.ChainID {
	return m.chainID
}

func (m *MockedChainCore) GlobalStateSync() coreutil.ChainStateSync {
	return m.onGlobalStateSync()
}

func (m *MockedChainCore) GetStateReader() state.OptimisticStateReader {
	return m.onGetStateReader()
}

func (m *MockedChainCore) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("implement me")
}

func (m *MockedChainCore) AttachToPeerMessages(fun func(recv *peering.RecvEvent)) {} // TODO

func (m *MockedChainCore) EnqueDismissChain(reason string) {} // TODO

func (m *MockedChainCore) StateCandidateToStateManager(state.VirtualStateAccess, ledgerstate.OutputID) {
} // TODO

func (m *MockedChainCore) ReceiveMessage(msg interface{}) {
	switch msgTypecasted := msg.(type) {
	case *peering.PeerMessage:
		m.onReceivePeerMessage(msgTypecasted)
	case *messages.DismissChainMsg:
		m.onReceiveDismissChainMsg(msgTypecasted)
	case *messages.StateTransitionMsg:
		m.onReceiveStateTransitionMsg(msgTypecasted)
	case *messages.StateCandidateMsg:
		m.onReceiveStateCandidateMsg(msgTypecasted)
	case *messages.InclusionStateMsg:
		m.onReceiveInclusionStateMsg(msgTypecasted)
	case *messages.StateMsg:
		m.onReceiveStateMsg(msgTypecasted)
	case *messages.VMResultMsg:
		m.onReceiveVMResultMsg(msgTypecasted)
	case *messages.AsynchronousCommonSubsetMsg:
		m.onReceiveAsynchronousCommonSubsetMsg(msgTypecasted)
	case messages.TimerTick:
		m.onReceiveTimerTick(msgTypecasted)
	}
}

func (m *MockedChainCore) Events() chain.ChainEvents {
	return m
}

func (m *MockedChainCore) Processors() *processors.Cache {
	return m.processors
}

func (m *MockedChainCore) RequestProcessed() *events.Event {
	return m.eventRequestProcessed
}

func (m *MockedChainCore) ChainTransition() *events.Event {
	return m.eventStateTransition
}

func (m *MockedChainCore) OnStateTransition(f func(data *chain.ChainTransitionEventData)) {
	m.onEventStateTransition = f
}

func (m *MockedChainCore) OnRequestProcessed(f func(id iscp.RequestID)) {
	m.onEventRequestProcessed = f
}

func (m *MockedChainCore) OnReceivePeerMessage(f func(*peering.PeerMessage)) {
	m.onReceivePeerMessage = f
}

func (m *MockedChainCore) OnReceiveDismissChainMsg(f func(*messages.DismissChainMsg)) {
	m.onReceiveDismissChainMsg = f
}

func (m *MockedChainCore) OnReceiveStateTransitionMsg(f func(*messages.StateTransitionMsg)) {
	m.onReceiveStateTransitionMsg = f
}

func (m *MockedChainCore) OnReceiveStateCandidateMsg(f func(*messages.StateCandidateMsg)) {
	m.onReceiveStateCandidateMsg = f
}

func (m *MockedChainCore) OnReceiveInclusionStateMsg(f func(*messages.InclusionStateMsg)) {
	m.onReceiveInclusionStateMsg = f
}

func (m *MockedChainCore) OnReceiveStateMsg(f func(*messages.StateMsg)) {
	m.onReceiveStateMsg = f
}

func (m *MockedChainCore) OnReceiveVMResultMsg(f func(*messages.VMResultMsg)) {
	m.onReceiveVMResultMsg = f
}

func (m *MockedChainCore) OnReceiveAsynchronousCommonSubsetMsg(f func(*messages.AsynchronousCommonSubsetMsg)) {
	m.onReceiveAsynchronousCommonSubsetMsg = f
}

func (m *MockedChainCore) OnReceiveTimerTick(f func(messages.TimerTick)) {
	m.onReceiveTimerTick = f
}

func (m *MockedChainCore) OnGetStateReader(f func() state.OptimisticStateReader) {
	m.onGetStateReader = f
}

func (m *MockedChainCore) OnGlobalStateSync(f func() coreutil.ChainStateSync) {
	m.onGlobalStateSync = f
}

func (m *MockedChainCore) GlobalSolidIndex() *atomic.Uint32 {
	return nil
}

func (m *MockedChainCore) ReceiveOffLedgerRequest(_ *request.OffLedger, _ string) {
}
