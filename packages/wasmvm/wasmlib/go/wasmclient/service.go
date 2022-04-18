// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/mr-tron/base58"
)

type ArgMap dict.Dict

func (m ArgMap) Get(key string) []byte {
	return m[kv.Key(key)]
}

func (m ArgMap) Set(key string, value []byte) {
	m[kv.Key(key)] = value
}

type ResMap dict.Dict

func (m ResMap) Get(key string) []byte {
	return m[kv.Key(key)]
}

type IEventHandler interface {
	CallHandler(topic string, params []string)
}

type IWaspClient interface {
	CallView(chainID *iscp.ChainID, hContract iscp.Hname, functionName string, args dict.Dict, optimisticReadTimeout ...time.Duration) (dict.Dict, error)
	CallViewByHname(chainID *iscp.ChainID, hContract, hFunction iscp.Hname, args dict.Dict, optimisticReadTimeout ...time.Duration) (dict.Dict, error)
	PostOffLedgerRequest(chainID *iscp.ChainID, req *iscp.OffLedgerRequestData) error
	WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) (*iscp.Receipt, error)
}

type Service struct {
	chainID       *iscp.ChainID
	cvt           wasmhost.WasmConvertor
	Err           error
	eventDone     chan bool
	eventHandlers []IEventHandler
	keyPair       *cryptolib.KeyPair
	Req           Request
	scName        string
	scHname       iscp.Hname
	svcClient     IServiceClient
	waspClient    IWaspClient
}

func NewService(svcClient IServiceClient, chainID *wasmtypes.ScChainID, scName string) *Service {
	s := &Service{}
	s.svcClient = svcClient
	s.waspClient = svcClient.WaspClient()
	s.scName = scName
	s.scHname = iscp.Hn(scName)
	s.chainID, s.Err = iscp.ChainIDFromBytes(chainID.Bytes())
	return s
}

func (s *Service) CallView(viewName string, args ArgMap) (ResMap, error) {
	res, err := s.waspClient.CallView(s.chainID, s.scHname, viewName, dict.Dict(args))
	if err != nil {
		return nil, err
	}
	return ResMap(res), nil
}

func (s *Service) ChainID() wasmtypes.ScChainID {
	return s.cvt.ScChainID(s.chainID)
}

func (s *Service) InitFuncCallContext() {
	_ = wasmhost.Connect(s)
}

func (s *Service) InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname {
	_ = wasmhost.Connect(s)
	return wasmtypes.ScHname(s.scHname)
}

func (s *Service) postRequestOffLedger(hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) Request {
	s.Req.id, s.Req.err = s.svcClient.PostRequest(s.chainID, s.scHname, hFuncName, params, allowance, keyPair)
	s.Err = s.Req.err
	return s.Req
}

func (s *Service) Register(handler IEventHandler) error {
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
func (s *Service) ServiceContractName(contractName string) {
	s.scHname = iscp.Hn(contractName)
}

func (s *Service) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair
}

func (s *Service) Unregister(handler IEventHandler) {
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

func (s *Service) WaitRequest(reqID ...*iscp.RequestID) error {
	id := s.Req.id
	if len(reqID) == 1 {
		id = reqID[0]
	}
	if id == nil {
		return nil
	}
	_, err := s.waspClient.WaitUntilRequestProcessed(s.chainID, *id, 1*time.Minute)
	// TODO check receipt ? - not sure if that is intended here
	return err
}

func (s *Service) startEventHandlers() error {
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
					msg := strings.Split(msgSplit[3], "|")
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

func (s *Service) stopEventHandlers() {
	if len(s.eventHandlers) > 0 {
		s.eventDone <- true
	}
}

/////////////////////////////////////////////////////////////////

func Base58Decode(s string) []byte {
	res, err := base58.Decode(s)
	if err != nil {
		panic("invalid base58 encoding")
	}
	return res
}

func Base58Encode(b []byte) string {
	return base58.Encode(b)
}
