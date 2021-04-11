package mock_chain

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type MockedChainCore struct {
	chainID                 coretypes.ChainID
	eventStateTransition    *events.Event
	eventRequestProcessed   *events.Event
	onEventStateTransition  func(data *chain.StateTransitionEventData)
	onEventRequestProcessed func(id coretypes.RequestID)
	onReceiveMessage        func(i interface{})
	log                     *logger.Logger
}

func NewMockedChainCore(chainID coretypes.ChainID, log *logger.Logger) *MockedChainCore {
	ret := &MockedChainCore{
		chainID: chainID,
		log:     log,
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.StateTransitionEventData))(params[0].(*chain.StateTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		onEventStateTransition: func(msg *chain.StateTransitionEventData) {
			chain.LogStateTransition(msg, log)
		},
		onEventRequestProcessed: func(id coretypes.RequestID) {
			log.Infof("onEventRequestProcessed: %s", id)
		},
	}
	ret.eventStateTransition.Attach(events.NewClosure(func(data *chain.StateTransitionEventData) {
		ret.onEventStateTransition(data)
	}))

	ret.eventRequestProcessed.Attach(events.NewClosure(func(id coretypes.RequestID) {
		ret.onEventRequestProcessed(id)
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
	panic("implement me")
}

func (m *MockedChainCore) RequestProcessed() *events.Event {
	return m.eventRequestProcessed
}

func (m *MockedChainCore) StateTransition() *events.Event {
	return m.eventStateTransition
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
