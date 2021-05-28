package testchain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type MockedChainCore struct {
	chainID                 coretypes.ChainID
	processors              *processors.ProcessorCache
	eventStateTransition    *events.Event
	eventRequestProcessed   *events.Event
	eventStateSynced        *events.Event
	onEventStateTransition  func(data *chain.StateTransitionEventData)
	onEventRequestProcessed func(id coretypes.RequestID)
	onEventStateSynced      func(id ledgerstate.OutputID, blockIndex uint32)
	onReceiveMessage        func(i interface{})
	onSync                  func(out ledgerstate.OutputID, blockIndex uint32)
	log                     *logger.Logger
}

func NewMockedChainCore(chainID coretypes.ChainID, log *logger.Logger) *MockedChainCore {
	ret := &MockedChainCore{
		chainID:    chainID,
		processors: processors.MustNew(),
		log:        log,
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.StateTransitionEventData))(params[0].(*chain.StateTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		eventStateSynced: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(outputID ledgerstate.OutputID, blockIndex uint32))(params[0].(ledgerstate.OutputID), params[1].(uint32))
		}),
		onEventStateTransition: func(msg *chain.StateTransitionEventData) {
			chain.LogStateTransition(msg, log)
		},
		onEventRequestProcessed: func(id coretypes.RequestID) {
			log.Infof("onEventRequestProcessed: %s", id)
		},
		onEventStateSynced: func(outputID ledgerstate.OutputID, blockIndex uint32) {
			chain.LogSyncedEvent(outputID, blockIndex, log)
		},
	}
	ret.eventStateTransition.Attach(events.NewClosure(func(data *chain.StateTransitionEventData) {
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

func (m *MockedChainCore) ID() *coretypes.ChainID {
	return &m.chainID
}

func (m *MockedChainCore) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("implement me")
}

func (m *MockedChainCore) ReceiveMessage(i interface{}) {
	m.onReceiveMessage(i)
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

func (m *MockedChainCore) StateTransition() *events.Event {
	return m.eventStateTransition
}

func (m *MockedChainCore) StateSynced() *events.Event {
	return m.eventStateSynced
}

func (m *MockedChainCore) OnStateTransition(f func(data *chain.StateTransitionEventData)) {
	m.onEventStateTransition = f
}

func (m *MockedChainCore) OnRequestProcessed(f func(id coretypes.RequestID)) {
	m.onEventRequestProcessed = f
}

func (m *MockedChainCore) OnReceiveMessage(f func(i interface{})) {
	m.onReceiveMessage = f
}

func (m *MockedChainCore) OnStateSynced(f func(out ledgerstate.OutputID, blockIndex uint32)) {
	m.onEventStateSynced = f
}
