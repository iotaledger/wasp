package test_statemgr

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type mockedChain struct {
	chainID                 coretypes.ChainID
	eventStateTransition    *events.Event
	eventRequestProcessed   *events.Event
	onEventStateTransition  func(data *chain.StateTransitionEventData)
	onEventRequestProcessed func(id coretypes.RequestID)
	log                     *logger.Logger
}

func NewMockedChain(chainID coretypes.ChainID, log *logger.Logger) *mockedChain {
	ret := &mockedChain{
		chainID: chainID,
		log:     log,
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.StateTransitionEventData))(params[0].(*chain.StateTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ coretypes.RequestID))(params[0].(coretypes.RequestID))
		}),
		onEventStateTransition: func(d *chain.StateTransitionEventData) {
			log.Infof("onEventStateTransition: %+v", *d)
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

func (m *mockedChain) ID() *coretypes.ChainID {
	return &m.chainID
}

func (m *mockedChain) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("implement me")
}

func (m *mockedChain) ReceiveMessage(i interface{}) {
	panic("implement me")
}

func (m *mockedChain) Events() chain.ChainEvents {
	return m
}

func (m *mockedChain) Processors() *processors.ProcessorCache {
	panic("implement me")
}

func (m *mockedChain) RequestProcessed() *events.Event {
	return m.eventRequestProcessed
}

func (m *mockedChain) StateTransition() *events.Event {
	return m.eventStateTransition
}

func (m *mockedChain) OnStateTransition(f func(data *chain.StateTransitionEventData)) {
	m.onEventStateTransition = f
}

func (m *mockedChain) OnRequestProcessed(f func(id coretypes.RequestID)) {
	m.onEventRequestProcessed = f
}
