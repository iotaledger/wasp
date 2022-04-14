// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/subscribe"
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

type Service struct {
	chainID       *iscp.ChainID
	cvt           wasmhost.WasmConvertor
	Err           error
	eventHandlers []IEventHandler
	keyPair       *cryptolib.KeyPair
	Req           Request
	scHname       iscp.Hname
	waspClient    *client.WaspClient
}

func (s *Service) Init(svcClient *ServiceClient, chainID *wasmtypes.ScChainID, scHname uint32) (err error) {
	s.waspClient = svcClient.waspClient
	s.scHname = iscp.Hname(scHname)
	s.chainID, err = iscp.ChainIDFromBytes(chainID.Bytes())
	if err != nil {
		return err
	}
	return s.startEventHandlers(svcClient.eventPort)
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
	req := iscp.NewOffLedgerRequest(s.chainID, s.scHname, hFuncName, params, uint64(time.Now().UnixNano()))
	req.WithTransfer(allowance)
	req.Sign(keyPair)
	s.Err = s.waspClient.PostOffLedgerRequest(s.chainID, req)
	if s.Err != nil {
		s.Req.err = s.Err
		return s.Req
	}
	s.Req.err = nil
	id := req.ID()
	s.Req.id = &id
	return s.Req
}

func (s *Service) Register(handler IEventHandler) {
	for _, h := range s.eventHandlers {
		if h == handler {
			return
		}
	}
	s.eventHandlers = append(s.eventHandlers, handler)
}

// overrides default contract name
func (s *Service) ServiceContractName(contractName string) {
	s.scHname = iscp.Hn(contractName)
}

func (s *Service) SignRequests(keyPair *cryptolib.KeyPair) {
	s.keyPair = keyPair
}

func (s *Service) Unegister(handler IEventHandler) {
	for i, h := range s.eventHandlers {
		if h == handler {
			s.eventHandlers = append(s.eventHandlers[:i], s.eventHandlers[i+1:]...)
			return
		}
	}
}

func (s *Service) WaitRequest(req Request) error {
	return s.waspClient.WaitUntilRequestProcessed(s.chainID, *req.id, 1*time.Minute)
}

func (s *Service) startEventHandlers(eventPort string) error {
	chMsg := make(chan []string, 20)
	chDone := make(chan bool)
	err := subscribe.Subscribe(eventPort, chMsg, chDone, true, "")
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
