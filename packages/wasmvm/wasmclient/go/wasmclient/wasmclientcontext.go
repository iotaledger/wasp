// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type WasmClientContext struct {
	Err       error
	keyPair   *cryptolib.KeyPair
	ReqID     wasmtypes.ScRequestID
	scName    string
	scHname   wasmtypes.ScHname
	svcClient IClientService
}

var (
	_ wasmlib.ScFuncCallContext = new(WasmClientContext)
	_ wasmlib.ScViewCallContext = new(WasmClientContext)
)

// NewWasmClientContext uses IClientService instead of WasmClientService
// because this could also be a SoloClientService
func NewWasmClientContext(svcClient IClientService, scName string) *WasmClientContext {
	s := &WasmClientContext{
		svcClient: svcClient,
		scName:    scName,
	}
	s.ServiceContractName(scName)
	return s
}

func (s *WasmClientContext) CurrentChainID() wasmtypes.ScChainID {
	return s.svcClient.CurrentChainID()
}

func (s *WasmClientContext) CurrentKeyPair() *cryptolib.KeyPair {
	return s.keyPair
}

func (s *WasmClientContext) CurrentSvcClient() IClientService {
	return s.svcClient
}

func (s *WasmClientContext) InitFuncCallContext() {
}

func (s *WasmClientContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	_ = hContract
	return s.scHname
}

// Register the event handler. So the corresponding incoming events will be handled by this event handler
func (s *WasmClientContext) Register(handler wasmlib.IEventHandlers) {
	s.Err = s.svcClient.SubscribeEvents(&WasmClientEvents{
		chainID:    s.svcClient.CurrentChainID(),
		contractID: s.scHname,
		handler:    handler,
	})
}

func (s *WasmClientContext) ServiceContractName(contractName string) {
	s.scHname = wasmtypes.HnameFromBytes(isc.Hn(contractName).Bytes())
}

func (s *WasmClientContext) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair
}

func (s *WasmClientContext) Unregister(eventsID uint32) {
	s.svcClient.UnsubscribeEvents(eventsID)
}

func (s *WasmClientContext) WaitRequest(reqID ...wasmtypes.ScRequestID) {
	requestID := s.ReqID
	if len(reqID) == 1 {
		requestID = reqID[0]
	}
	s.Err = s.svcClient.WaitUntilRequestProcessed(requestID, 60*time.Second)
}
