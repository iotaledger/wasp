// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testchain

import (
	"testing"

	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/contracts/native/inccounter"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/coreprocessors"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/processors"
)

type MockedChainCore struct {
	T                       *testing.T
	chainID                 *iscp.ChainID
	processors              *processors.Cache
	eventStateTransition    *events.Event
	eventRequestProcessed   *events.Event
	getNetIDsFun            func() []string
	onGlobalStateSync       func() coreutil.ChainStateSync
	onGetStateReader        func() state.OptimisticStateReader
	onEventStateTransition  func(data *chain.ChainTransitionEventData)
	onEventRequestProcessed func(id iscp.RequestID)
	onSendPeerMsg           func(netID string, msgReceiver byte, msgType byte, msgData []byte)
	onStateCandidate        func(state state.VirtualStateAccess, outputID *iotago.UTXOInput)
	onDismissChain          func(reason string)
	onAliasOutput           func(chainOutput *iscp.AliasOutputWithID)
	onOffLedgerRequest      func(msg *messages.OffLedgerRequestMsgIn)
	onMissingRequestIDs     func(msg *messages.MissingRequestIDsMsgIn)
	onMissingRequest        func(msg *messages.MissingRequestMsg)
	onTimerTick             func(tick int)
	onSync                  func(out iotago.OutputID, blockIndex uint32) //nolint:structcheck,unused
	log                     *logger.Logger
}

var _ chain.ChainCore = &MockedChainCore{}

func NewMockedChainCore(t *testing.T, chainID *iscp.ChainID, log *logger.Logger) *MockedChainCore {
	receiveFailFun := func(typee string, msg interface{}) {
		t.Fatalf("Receiving of %s is not implemented, but %v is received", typee, msg)
	}
	ret := &MockedChainCore{
		T:          t,
		chainID:    chainID,
		processors: processors.MustNew(coreprocessors.Config().WithNativeContracts(inccounter.Processor)),
		log:        log.Named("mc-" + chainID.AsAddress().String()[2:8]),
		getNetIDsFun: func() []string {
			t.Fatalf("List of netIDs is not known")
			return []string{}
		},
		eventStateTransition: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ *chain.ChainTransitionEventData))(params[0].(*chain.ChainTransitionEventData))
		}),
		eventRequestProcessed: events.NewEvent(func(handler interface{}, params ...interface{}) {
			handler.(func(_ iscp.RequestID))(params[0].(iscp.RequestID))
		}),
		onEventRequestProcessed: func(id iscp.RequestID) {
			log.Infof("onEventRequestProcessed: %s", id)
		},
		onSendPeerMsg: func(netID string, msgReceiver byte, msgType byte, msgData []byte) {
			t.Fatalf("Sending to peer msg not implemented, netID=%v, receiver=%v, msgType=%v", netID, msgReceiver, msgType)
		},
		onStateCandidate: func(state state.VirtualStateAccess, outputID *iotago.UTXOInput) {
			t.Fatalf("Receiving state candidate not implemented, outputID=%v", iscp.OID(outputID))
		},
		onDismissChain: func(reason string) { t.Fatalf("Dismissing chain not implemented, reason=%v", reason) },
		onAliasOutput: func(chainOutput *iscp.AliasOutputWithID) {
			t.Fatalf("Receiving alias output not implemented, chain output ID=%v", iscp.OID(chainOutput.ID()))
		},
		onOffLedgerRequest:  func(msg *messages.OffLedgerRequestMsgIn) { receiveFailFun("*messages.OffLedgerRequestMsgIn", msg) },
		onMissingRequestIDs: func(msg *messages.MissingRequestIDsMsgIn) { receiveFailFun("*messages.MissingRequestIDsMsgIn", msg) },
		onMissingRequest:    func(msg *messages.MissingRequestMsg) { receiveFailFun("*messages.MissingRequestMsg", msg) },
		onTimerTick:         func(tick int) { t.Fatalf("Receiving timer tick not implemented: index=%v", tick) },
	}
	ret.onEventStateTransition = func(msg *chain.ChainTransitionEventData) {
		chain.LogStateTransition(msg.VirtualState.BlockIndex(), iscp.OID(msg.ChainOutput.ID()), state.RootCommitment(msg.VirtualState.TrieNodeStore()), nil, log)
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

func (m *MockedChainCore) StateCandidateToStateManager(virtualState state.VirtualStateAccess, outputID *iotago.UTXOInput) {
	m.onStateCandidate(virtualState, outputID)
}

func (m *MockedChainCore) EnqueueDismissChain(reason string) {
	m.onDismissChain(reason)
}

func (m *MockedChainCore) EnqueueAliasOutput(chainOutput *iscp.AliasOutputWithID) {
	m.onAliasOutput(chainOutput)
}

func (m *MockedChainCore) EnqueueOffLedgerRequestMsg(msg *messages.OffLedgerRequestMsgIn) {
	m.onOffLedgerRequest(msg)
}

func (m *MockedChainCore) EnqueueMissingRequestIDsMsg(msg *messages.MissingRequestIDsMsgIn) {
	m.onMissingRequestIDs(msg)
}

func (m *MockedChainCore) EnqueueMissingRequestMsg(msg *messages.MissingRequestMsg) {
	m.onMissingRequest(msg)
}

func (m *MockedChainCore) EnqueueTimerTick(tick int) {
	m.onTimerTick(tick)
}

func (m *MockedChainCore) Processors() *processors.Cache {
	return m.processors
}

func (m *MockedChainCore) TriggerChainTransition(data *chain.ChainTransitionEventData) {
	m.eventStateTransition.Trigger(data)
}

func (m *MockedChainCore) OnStateTransition(f func(data *chain.ChainTransitionEventData)) {
	m.onEventStateTransition = f
}

func (m *MockedChainCore) OnRequestProcessed(f func(id iscp.RequestID)) {
	m.onEventRequestProcessed = f
}

func (m *MockedChainCore) OnGetStateReader(f func() state.OptimisticStateReader) {
	m.onGetStateReader = f
}

func (m *MockedChainCore) OnGlobalStateSync(f func() coreutil.ChainStateSync) {
	m.onGlobalStateSync = f
}

func (m *MockedChainCore) OnSendPeerMsg(fun func(netID string, msgReceiver byte, msgType byte, msgData []byte)) {
	m.onSendPeerMsg = fun
}

func (m *MockedChainCore) OnStateCandidate(fun func(state state.VirtualStateAccess, outputID *iotago.UTXOInput)) {
	m.onStateCandidate = fun
}

func (m *MockedChainCore) OnDismissChain(fun func(reason string)) {
	m.onDismissChain = fun
}

func (m *MockedChainCore) OnAliasOutput(fun func(chainOutput *iscp.AliasOutputWithID)) {
	m.onAliasOutput = fun
}

func (m *MockedChainCore) OnOffLedgerRequest(fun func(msg *messages.OffLedgerRequestMsgIn)) {
	m.onOffLedgerRequest = fun
}

func (m *MockedChainCore) OnMissingRequestIDs(fun func(msg *messages.MissingRequestIDsMsgIn)) {
	m.onMissingRequestIDs = fun
}

func (m *MockedChainCore) OnMissingRequest(fun func(msg *messages.MissingRequestMsg)) {
	m.onMissingRequest = fun
}

func (m *MockedChainCore) OnTimerTick(fun func(tick int)) {
	m.onTimerTick = fun
}

func (m *MockedChainCore) GetChainNodes() []peering.PeerStatusProvider {
	panic("not implemented MockedChainCore::GetChainNodes")
}

func (m *MockedChainCore) GetCandidateNodes() []*governance.AccessNodeInfo {
	panic("not implemented MockedChainCore::GetCandidateNodes")
}

func (m *MockedChainCore) VirtualStateAccess() state.VirtualStateAccess {
	panic("not implemented MockedChainCore::VirtualStateAccess")
}

func (m *MockedChainCore) GetAnchorOutput() *iscp.AliasOutputWithID {
	panic("not implemented MockedChainCore::GetAnchorOutput")
}
