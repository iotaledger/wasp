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
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type WasmClientContext struct {
	chainID       wasmtypes.ScChainID
	Err           error
	eventHandlers []wasmlib.IEventHandlers
	keyPair       *cryptolib.KeyPair
	nonce         uint64
	ReqID         wasmtypes.ScRequestID
	scName        string
	scHname       wasmtypes.ScHname
	svcClient     IClientService
}

var (
	_ wasmlib.ScFuncCallContext = new(WasmClientContext)
	_ wasmlib.ScViewCallContext = new(WasmClientContext)
)

func NewWasmClientContext(svcClient IClientService, chainID string, scName string) *WasmClientContext {
	s := &WasmClientContext{
		svcClient: svcClient,
		scName:    scName,
		Err:       SetSandboxWrappers(chainID),
	}
	s.ServiceContractName(scName)
	if s.Err == nil {
		// only do this when SetSandboxWrappers() was successful
		s.chainID = wasmtypes.ChainIDFromString(chainID)
	}
	return s
}

func (s *WasmClientContext) CurrentChainID() wasmtypes.ScChainID {
	return s.chainID
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

func (s *WasmClientContext) Register(handler wasmlib.IEventHandlers) error {
	for _, h := range s.eventHandlers {
		if h == handler {
			return nil
		}
	}
	s.eventHandlers = append(s.eventHandlers, handler)
	if len(s.eventHandlers) > 1 {
		return nil
	}
	return s.svcClient.SubscribeEvents(s.processEvent)
}

func (s *WasmClientContext) ServiceContractName(contractName string) {
	s.scHname = wasmtypes.HnameFromBytes(isc.Hn(contractName).Bytes())
}

func (s *WasmClientContext) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair

	// TODO not here
	// get last used nonce from accounts core contract
	iscAgent := isc.NewAgentID(keyPair.Address())
	agent := wasmtypes.AgentIDFromBytes(iscAgent.Bytes())
	ctx := NewWasmClientContext(s.svcClient, s.chainID.String(), coreaccounts.ScName)
	n := coreaccounts.ScFuncs.GetAccountNonce(ctx)
	n.Params.AgentID().SetValue(agent)
	n.Func.Call()
	s.Err = ctx.Err
	if s.Err == nil {
		s.nonce = n.Results.AccountNonce().Value()
	}
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
	s.Err = s.svcClient.WaitUntilRequestProcessed(s.chainID, requestID, 60*time.Second)
}

func (s *WasmClientContext) processEvent(msg *ContractEvent) {
	fmt.Printf("%s %s %s\n", msg.ChainID, msg.ContractID, msg.Data)

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
