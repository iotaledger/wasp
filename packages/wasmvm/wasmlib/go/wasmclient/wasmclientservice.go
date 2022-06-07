// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"time"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/subscribe"
)

type IClientService interface {
	CallViewByHname(chainID *iscp.ChainID, hContract, hFunction iscp.Hname, args dict.Dict) (dict.Dict, error)
	PostRequest(chainID *iscp.ChainID, hContract, hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) (*iscp.RequestID, error)
	SubscribeEvents(msg chan []string, done chan bool) error
	WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) error
}

type WasmClientService struct {
	waspClient *client.WaspClient
	eventPort  string
	nonce      uint64
}

var _ IClientService = new(WasmClientService)

func NewWasmClientService(waspAPI, eventPort string) *WasmClientService {
	return &WasmClientService{waspClient: client.NewWaspClient(waspAPI), eventPort: eventPort}
}

func DefaultWasmClientService() *WasmClientService {
	return NewWasmClientService("127.0.0.1:9090", "127.0.0.1:5550")
}

func (sc *WasmClientService) CallViewByHname(chainID *iscp.ChainID, hContract, hFunction iscp.Hname, args dict.Dict) (dict.Dict, error) {
	return sc.waspClient.CallViewByHname(chainID, hContract, hFunction, args)
}

func (sc *WasmClientService) PostRequest(chainID *iscp.ChainID, hContract, hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) (*iscp.RequestID, error) {
	sc.nonce++
	req := iscp.NewOffLedgerRequest(chainID, hContract, hFuncName, params, sc.nonce)
	req.WithAllowance(allowance)
	signed := req.Sign(keyPair)
	err := sc.waspClient.PostOffLedgerRequest(chainID, signed)
	if err != nil {
		return nil, err
	}
	id := signed.ID()
	return &id, nil
}

func (sc *WasmClientService) SubscribeEvents(msg chan []string, done chan bool) error {
	return subscribe.Subscribe(sc.eventPort, msg, done, false, "")
}

func (sc *WasmClientService) WaitUntilRequestProcessed(chainID *iscp.ChainID, reqID iscp.RequestID, timeout time.Duration) error {
	_, err := sc.waspClient.WaitUntilRequestProcessed(chainID, reqID, timeout)
	return err
}
