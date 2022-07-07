// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type IEventHandler interface {
	CallHandler(topic string, params []string)
}

type WasmClientContext struct {
	chainID       wasmtypes.ScChainID
	Err           error
	eventDone     chan bool
	eventHandlers []IEventHandler
	keyPair       *cryptolib.KeyPair
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

func NewWasmClientContext(svcClient IClientService, chainID wasmtypes.ScChainID, scName string) *WasmClientContext {
	s := &WasmClientContext{}
	s.svcClient = svcClient
	s.scName = scName
	s.scHname = wasmtypes.NewScHname(scName)
	s.chainID = chainID
	return s
}

func (s *WasmClientContext) CurrentChainID() wasmtypes.ScChainID {
	return s.chainID
}

func (s *WasmClientContext) InitFuncCallContext() {
	_ = wasmhost.Connect(s)
}

func (s *WasmClientContext) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	_ = wasmhost.Connect(s)
	return s.scHname
}

func (s *WasmClientContext) Register(handler IEventHandler) error {
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
	s.scHname = wasmtypes.NewScHname(contractName)
}

func (s *WasmClientContext) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair
}

func (s *WasmClientContext) Unregister(handler IEventHandler) {
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

func (s *WasmClientContext) WaitRequest(reqID ...wasmtypes.ScRequestID) error {
	requestID := s.ReqID
	if len(reqID) == 1 {
		requestID = reqID[0]
	}
	return s.svcClient.WaitUntilRequestProcessed(s.chainID, requestID, 1*time.Minute)
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
			for msgSplit := range chMsg {
				event := strings.Join(msgSplit, " ")
				fmt.Printf("%s\n", event)
				if msgSplit[0] == "vmmsg" {
					msgEscaped := strings.Split(msgSplit[3], "|")
					msg := make([]string, len(msgEscaped))
					for i, m := range msgEscaped {
						msg[i] = strings.ReplaceAll(m, "\\/", "|")
						msg[i] = strings.ReplaceAll(m, "\\\\", "\\")
					}
					topic := msg[0]
					params := msg[1:]
					for _, handler := range s.eventHandlers {
						handler.CallHandler(topic, params)
					}
				}
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
