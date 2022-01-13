// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"fmt"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/subscribe"
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
	eventHandlers []IEventHandler
	keyPair       *ed25519.KeyPair
	scHname       iscp.Hname
	waspClient    *client.WaspClient
}

func (s *Service) Init(svcClient *ServiceClient, chainID string, scHname uint32) (err error) {
	s.waspClient = svcClient.waspClient
	s.scHname = iscp.Hname(scHname)
	s.chainID, err = iscp.ChainIDFromString(chainID)
	if err != nil {
		return err
	}
	return s.startEventHandlers(svcClient.eventPort)
}

func (s *Service) AsClientFunc() ClientFunc {
	return ClientFunc{svc: s}
}

func (s *Service) AsClientView() ClientView {
	return ClientView{svc: s}
}

func (s *Service) CallView(viewName string, args ArgMap) (ResMap, error) {
	res, err := s.waspClient.CallView(s.chainID, s.scHname, viewName, dict.Dict(args))
	if err != nil {
		return nil, err
	}
	return ResMap(res), nil
}

func (s *Service) PostRequest(hFuncName uint32, args ArgMap, transfer *Transfer, keyPair *ed25519.KeyPair, onLedger bool) Request {
	bal, err := makeBalances(transfer)
	if err != nil {
		return Request{err: err}
	}
	reqArgs := requestargs.New()
	if args != nil {
		reqArgs.AddEncodeSimpleMany(dict.Dict(args))
	}

	if onLedger {
		return s.postRequestOnLedger(hFuncName, reqArgs, bal, keyPair)
	}

	req := request.NewOffLedger(s.chainID, s.scHname, iscp.Hname(hFuncName), reqArgs)
	req.WithTransfer(bal)
	req.Sign(keyPair)
	err = s.waspClient.PostOffLedgerRequest(s.chainID, req)
	if err != nil {
		return Request{err: err}
	}
	id := req.ID()
	return Request{id: &id}
}

func (s *Service) postRequestOnLedger(hFuncName uint32, args requestargs.RequestArgs, bal colored.Balances, pair *ed25519.KeyPair) Request {
	// TODO implement
	return Request{}
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

func (s *Service) SignRequests(keyPair *ed25519.KeyPair) {
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

func makeBalances(transfer *Transfer) (colored.Balances, error) {
	cb := colored.NewBalances()
	if transfer != nil {
		for color, amount := range transfer.xfer {
			c, err := colored.ColorFromBase58EncodedString(color)
			if err != nil {
				return nil, err
			}
			cb.Set(c, amount)
		}
	}
	return cb, nil
}
