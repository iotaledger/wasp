// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type WasmClientContext struct {
	Err           error
	eventHandlers []wasmlib.IEventHandlers
	keyPair       *cryptolib.KeyPair
	ReqID         wasmtypes.ScRequestID
	scName        string
	scHname       wasmtypes.ScHname
	svcClient     IClientService
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
	s.Err = nil
	for _, h := range s.eventHandlers {
		if h == handler {
			return
		}
	}
	s.eventHandlers = append(s.eventHandlers, handler)
	if len(s.eventHandlers) > 1 {
		return
	}
	s.Err = s.svcClient.SubscribeEvents(s.processEvent)
}

func (s *WasmClientContext) ServiceContractName(contractName string) {
	s.scHname = wasmtypes.HnameFromBytes(isc.Hn(contractName).Bytes())
}

func (s *WasmClientContext) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair
}

func (s *WasmClientContext) Unregister(handler wasmlib.IEventHandlers) {
	for i, h := range s.eventHandlers {
		if h == handler {
			s.eventHandlers = append(s.eventHandlers[:i], s.eventHandlers[i+1:]...)
			if len(s.eventHandlers) == 0 {
				s.svcClient.UnsubscribeEvents()
			}
			return
		}
	}
}

func (s *WasmClientContext) WaitRequest(reqID ...wasmtypes.ScRequestID) {
	requestID := s.ReqID
	if len(reqID) == 1 {
		requestID = reqID[0]
	}
	s.Err = s.svcClient.WaitUntilRequestProcessed(requestID, 60*time.Second)
}

func (s *WasmClientContext) processEvent(msg *ContractEvent) {
	if msg.ContractID != s.scHname ||
		msg.ChainID != s.svcClient.CurrentChainID() {
		return
	}
	fmt.Printf("%s %s %s\n", msg.ChainID.String(), msg.ContractID.String(), msg.Data)

	params := strings.Split(msg.Data, "|")
	for i, param := range params {
		params[i] = unescape(param)
	}
	topic := params[0]
	params = params[1:]
	for _, handler := range s.eventHandlers {
		handler.CallHandler(topic, params)
	}
}

func unescape(param string) string {
	i := strings.IndexByte(param, '~')
	if i < 0 {
		// no escape detected, return original string
		return param
	}

	switch param[i+1] {
	case '~': // escaped escape character
		return param[:i] + "~" + unescape(param[i+2:])
	case '/': // escaped vertical bar
		return param[:i] + "|" + unescape(param[i+2:])
	case '_': // escaped space
		return param[:i] + " " + unescape(param[i+2:])
	default:
		panic("invalid event encoding")
	}
}
