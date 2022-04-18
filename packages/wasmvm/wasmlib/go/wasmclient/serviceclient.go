// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"math"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/subscribe"
)

type IServiceClient interface {
	PostRequest(chainID *iscp.ChainID, hContract, hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) (*iscp.RequestID, error)
	SubscribeEvents(msg chan []string, done chan bool) error
	WaspClient() IWaspClient
}

type ServiceClient struct {
	waspClient IWaspClient
	eventPort  string
	nonce      uint64
}

var _ IServiceClient = new(ServiceClient)

func NewServiceClient(waspAPI, eventPort string) *ServiceClient {
	return &ServiceClient{waspClient: client.NewWaspClient(waspAPI), eventPort: eventPort}
}

func DefaultServiceClient() *ServiceClient {
	return NewServiceClient("127.0.0.1:9090", "127.0.0.1:5550")
}

func (sc *ServiceClient) PostRequest(chainID *iscp.ChainID, hContract, hFuncName iscp.Hname, params dict.Dict, allowance *iscp.Allowance, keyPair *cryptolib.KeyPair) (*iscp.RequestID, error) {
	sc.nonce++
	// TODO receive gas budget
	gas := uint64(math.MaxUint64) // maximum allowed gas
	req := iscp.NewOffLedgerRequest(chainID, hContract, hFuncName, params, sc.nonce, gas)
	req.WithTransfer(allowance)
	req.Sign(keyPair)
	err := sc.waspClient.PostOffLedgerRequest(chainID, req)
	if err != nil {
		return nil, err
	}
	id := req.ID()
	return &id, nil
}

func (sc *ServiceClient) SubscribeEvents(msg chan []string, done chan bool) error {
	return subscribe.Subscribe(sc.eventPort, msg, done, false, "")
}

func (sc *ServiceClient) WaspClient() IWaspClient {
	return sc.waspClient
}
