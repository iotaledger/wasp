// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"errors"
	"fmt"
	"strings"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type WasmClientContext struct {
	chainID       wasmtypes.ScChainID
	Err           error
	eventDone     chan bool
	eventHandlers []wasmlib.IEventHandlers
	eventReceived bool
	hrp           string
	keyPair       *cryptolib.KeyPair
	nonce         uint64
	ReqID         wasmtypes.ScRequestID
	scName        string
	scHname       wasmtypes.ScHname
	svcClient     IClientService
}

var (
	_ wasmlib.ScHost            = new(WasmClientContext)
	_ wasmlib.ScFuncCallContext = new(WasmClientContext)
	_ wasmlib.ScViewCallContext = new(WasmClientContext)
)

func NewWasmClientContext(svcClient IClientService, chain string, scName string) *WasmClientContext {
	s := &WasmClientContext{}
	s.svcClient = svcClient
	s.scName = scName
	s.ServiceContractName(scName)
	hrp, _, err := iotago.ParseBech32(chain)
	if err != nil {
		s.Err = err
		return s
	}
	s.hrp = string(hrp)
	_ = wasmhost.Connect(s)
	s.chainID = wasmtypes.ChainIDFromString(chain)
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
	_ = wasmhost.Connect(s)
}

func (s *WasmClientContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	_ = hContract
	_ = wasmhost.Connect(s)
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
	return s.startEventHandlers()
}

// overrides default contract name
func (s *WasmClientContext) ServiceContractName(contractName string) {
	s.scHname = wasmtypes.HnameFromBytes(isc.Hn(contractName).Bytes())
}

func (s *WasmClientContext) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair

	// get last used nonce from accounts core contract
	iscAgent := isc.NewAgentID(keyPair.Address())
	agent := wasmtypes.AgentIDFromBytes(iscAgent.Bytes())
	ctx := NewWasmClientContext(s.svcClient, s.chainID.String(), coreaccounts.ScName)
	n := coreaccounts.ScFuncs.GetAccountNonce(ctx)
	n.Params.AgentID().SetValue(agent)
	n.Func.Call()
	s.nonce = n.Results.AccountNonce().Value()
}

func (s *WasmClientContext) Unregister(handler wasmlib.IEventHandlers) {
	for i, h := range s.eventHandlers {
		if h == handler {
			s.eventHandlers = append(s.eventHandlers[:i], s.eventHandlers[i+1:]...)
			if len(s.eventHandlers) == 0 {
				s.stopEventHandlers()
			}
			return
		}
	}
}

func (s *WasmClientContext) WaitEvent() {
	s.Err = nil
	for i := 0; i < 100; i++ {
		if s.eventReceived {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.Err = errors.New("event wait timeout")
}

func (s *WasmClientContext) WaitRequest(reqID ...wasmtypes.ScRequestID) {
	requestID := s.ReqID
	if len(reqID) == 1 {
		requestID = reqID[0]
	}
	s.Err = s.svcClient.WaitUntilRequestProcessed(s.chainID, requestID, 1*time.Minute)
}

func (s *WasmClientContext) processEvent(msg []string) {
	fmt.Printf("%s\n", strings.Join(msg, " "))

	if msg[0] != "contract" {
		// not intended for us
		return
	}

	s.eventReceived = true

	params := strings.Split(msg[6], "|")
	for i, param := range params {
		params[i] = unescape(param)
	}
	topic := params[0]
	params = params[1:]
	for _, handler := range s.eventHandlers {
		handler.CallHandler(topic, params)
	}
}

func (s *WasmClientContext) startEventHandlers() error {
	chMsg := make(chan []string, 20)
	s.eventDone = make(chan bool)
	err := s.svcClient.SubscribeEvents(chMsg, s.eventDone)
	if err != nil {
		return err
	}
	go func() {
		for {
			for msg := range chMsg {
				s.processEvent(msg)
			}
		}
	}()
	return nil
}

func (s *WasmClientContext) stopEventHandlers() {
	if len(s.eventHandlers) > 0 {
		s.eventDone <- true
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
