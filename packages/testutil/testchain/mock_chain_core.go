package testchain

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"go.uber.org/atomic"
)

type MockedChainCore struct {
	T                                    *testing.T
	chainID                              chainid.ChainID
	processors                           *processors.ProcessorCache
	eventStateTransition                 *events.Event
	eventRequestProcessed                *events.Event
	eventStateSynced                     *events.Event
	onGlobalStateSync                    func() coreutil.ChainStateSync
	onGetStateReader                     func() state.OptimisticStateReader
	onEventStateTransition               func(data *chain.ChainTransitionEventData)
	onEventRequestProcessed              func(id coretypes.RequestID)
	onEventStateSynced                   func(id ledgerstate.OutputID, blockIndex uint32)
	onReceivePeerMessage                 func(*peering.PeerMessage)
	onReceiveDismissChainMsg             func(*chain.DismissChainMsg)
	onReceiveStateTransitionMsg          func(*chain.StateTransitionMsg)
	onReceiveStateCandidateMsg           func(*chain.StateCandidateMsg)
	onReceiveInclusionStateMsg           func(*chain.InclusionStateMsg)
	onReceiveStateMsg                    func(*chain.StateMsg)
	onReceiveVMResultMsg                 func(*chain.VMResultMsg)
	onReceiveAsynchronousCommonSubsetMsg func(*chain.AsynchronousCommonSubsetMsg)
	onReceiveTimerTick                   func(chain.TimerTick)
	onSync                               func(out ledgerstate.OutputID, blockIndex uint32)
	log                                  *logger.Logger
}

func NewMockedChainCore(t *testing.T, chainID chainid.ChainID, log *logger.Logger) *MockedChainCore {
	receiveFailFun := func(typee string, msg interface{}) {
		t.Fatalf("Receiving of %s is not implemented, but %v is received", typee, msg)
	}
	ret := &MockedChainCore{
		T:          t,
		chainID:    chainID,
		processors: processors.MustNew(),
		log:        log,
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		eventStateSynced: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(outputID ledgerstate.OutputID, blockIndex uint32))(params[0].(ledgerstate.OutputID), params[1].(uint32))
		}),
		onEventRequestProcessed: func(id coretypes.RequestID) {
			log.Infof("onEventRequestProcessed: %s", id)
		},
		onEventStateSynced: func(outputID ledgerstate.OutputID, blockIndex uint32) {
			chain.LogSyncedEvent(outputID, blockIndex, log)
		},
		onReceivePeerMessage:        func(msg *peering.PeerMessage) { receiveFailFun("*peering.PeerMessage", msg) },
		onReceiveDismissChainMsg:    func(msg *chain.DismissChainMsg) { receiveFailFun("*chain.DismissChainMs", msg) },
		onReceiveStateTransitionMsg: func(msg *chain.StateTransitionMsg) { receiveFailFun("*chain.StateTransitionMsg", msg) },
		onReceiveStateCandidateMsg:  func(msg *chain.StateCandidateMsg) { receiveFailFun("*chain.StateCandidateMsg", msg) },
		onReceiveInclusionStateMsg:  func(msg *chain.InclusionStateMsg) { receiveFailFun("*chain.InclusionStateMsg", msg) },
		onReceiveStateMsg:           func(msg *chain.StateMsg) { receiveFailFun("*chain.StateMsg", msg) },
		onReceiveVMResultMsg:        func(msg *chain.VMResultMsg) { receiveFailFun("*chain.VMResultMsg", msg) },
		onReceiveAsynchronousCommonSubsetMsg: func(msg *chain.AsynchronousCommonSubsetMsg) {
			receiveFailFun("*chain.AsynchronousCommonSubsetMsg", msg)
		},
		onReceiveTimerTick: func(msg chain.TimerTick) { receiveFailFun("chain.TimerTick", msg) },
	}
	ret.onEventStateTransition = func(msg *chain.ChainTransitionEventData) {
		chain.LogStateTransition(msg, nil, log)
	}
	ret.eventStateTransition.Attach(events.NewClosure(func(data *chain.ChainTransitionEventData) {
		ret.onEventStateTransition(data)
	}))
	ret.eventRequestProcessed.Attach(events.NewClosure(func(id coretypes.RequestID) {
		ret.onEventRequestProcessed(id)
	}))
	ret.eventStateSynced.Attach(events.NewClosure(func(outid ledgerstate.OutputID, blockIndex uint32) {
		ret.onEventStateSynced(outid, blockIndex)
	}))
	return ret
}

func (m *MockedChainCore) Log() *logger.Logger {
	return m.log
}

func (m *MockedChainCore) ID() *chainid.ChainID {
	return &m.chainID
}

func (c *MockedChainCore) GlobalStateSync() coreutil.ChainStateSync {
	return c.onGlobalStateSync()
}

func (m *MockedChainCore) GetStateReader() state.OptimisticStateReader {
	return m.onGetStateReader()
}

func (m *MockedChainCore) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("implement me")
}

func (m *MockedChainCore) ReceiveMessage(msg interface{}) {
	switch msg.(type) {
	case *peering.PeerMessage:
		m.onReceivePeerMessage(msg.(*peering.PeerMessage))
	case *chain.DismissChainMsg:
		m.onReceiveDismissChainMsg(msg.(*chain.DismissChainMsg))
	case *chain.StateTransitionMsg:
		m.onReceiveStateTransitionMsg(msg.(*chain.StateTransitionMsg))
	case *chain.StateCandidateMsg:
		m.onReceiveStateCandidateMsg(msg.(*chain.StateCandidateMsg))
	case *chain.InclusionStateMsg:
		m.onReceiveInclusionStateMsg(msg.(*chain.InclusionStateMsg))
	case *chain.StateMsg:
		m.onReceiveStateMsg(msg.(*chain.StateMsg))
	case *chain.VMResultMsg:
		m.onReceiveVMResultMsg(msg.(*chain.VMResultMsg))
	case *chain.AsynchronousCommonSubsetMsg:
		m.onReceiveAsynchronousCommonSubsetMsg(msg.(*chain.AsynchronousCommonSubsetMsg))
	case chain.TimerTick:
		m.onReceiveTimerTick(msg.(chain.TimerTick))
	}
}

func (m *MockedChainCore) Events() chain.ChainEvents {
	return m
}

func (m *MockedChainCore) Processors() *processors.ProcessorCache {
	return m.processors
}

func (m *MockedChainCore) RequestProcessed() *events.Event {
	return m.eventRequestProcessed
}

func (m *MockedChainCore) ChainTransition() *events.Event {
	return m.eventStateTransition
}

func (m *MockedChainCore) StateSynced() *events.Event {
	return m.eventStateSynced
}

func (m *MockedChainCore) OnStateTransition(f func(data *chain.ChainTransitionEventData)) {
	m.onEventStateTransition = f
}

func (m *MockedChainCore) OnRequestProcessed(f func(id coretypes.RequestID)) {
	m.onEventRequestProcessed = f
}

func (m *MockedChainCore) OnReceivePeerMessage(f func(*peering.PeerMessage)) {
	m.onReceivePeerMessage = f
}

func (m *MockedChainCore) OnReceiveDismissChainMsg(f func(*chain.DismissChainMsg)) {
	m.onReceiveDismissChainMsg = f
}

func (m *MockedChainCore) OnReceiveStateTransitionMsg(f func(*chain.StateTransitionMsg)) {
	m.onReceiveStateTransitionMsg = f
}

func (m *MockedChainCore) OnReceiveStateCandidateMsg(f func(*chain.StateCandidateMsg)) {
	m.onReceiveStateCandidateMsg = f
}

func (m *MockedChainCore) OnReceiveInclusionStateMsg(f func(*chain.InclusionStateMsg)) {
	m.onReceiveInclusionStateMsg = f
}

func (m *MockedChainCore) OnReceiveStateMsg(f func(*chain.StateMsg)) {
	m.onReceiveStateMsg = f
}

func (m *MockedChainCore) OnReceiveVMResultMsg(f func(*chain.VMResultMsg)) {
	m.onReceiveVMResultMsg = f
}

func (m *MockedChainCore) OnReceiveAsynchronousCommonSubsetMsg(f func(*chain.AsynchronousCommonSubsetMsg)) {
	m.onReceiveAsynchronousCommonSubsetMsg = f
}

func (m *MockedChainCore) OnReceiveTimerTick(f func(chain.TimerTick)) {
	m.onReceiveTimerTick = f
}

func (m *MockedChainCore) OnStateSynced(f func(out ledgerstate.OutputID, blockIndex uint32)) {
	m.onEventStateSynced = f
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

func (m *MockedChainCore) ReceiveOffLedgerRequest(req *request.RequestOffLedger) {
}
