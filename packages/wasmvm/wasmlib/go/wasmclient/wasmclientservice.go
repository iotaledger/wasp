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
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type IClientService interface {
	CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (wasmtypes.ScRequestID, error)
	SubscribeEvents(msg chan []string, done chan bool) error
	WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error
}

type WasmClientService struct {
	cvt        wasmhost.WasmConvertor
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

func (sc *WasmClientService) CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscpChainID := sc.cvt.IscpChainID(&chainID)
	iscpContract := sc.cvt.IscpHname(hContract)
	iscpFunction := sc.cvt.IscpHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	res, err := sc.waspClient.CallViewByHname(iscpChainID, iscpContract, iscpFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (sc *WasmClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscpChainID := sc.cvt.IscpChainID(&chainID)
	iscpContract := sc.cvt.IscpHname(hContract)
	iscpFunction := sc.cvt.IscpHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	sc.nonce++
	req := iscp.NewOffLedgerRequest(iscpChainID, iscpContract, iscpFunction, params, sc.nonce)
	iscpAllowance := sc.cvt.IscpAllowance(allowance)
	req.WithAllowance(iscpAllowance)
	signed := req.Sign(keyPair)
	err = sc.waspClient.PostOffLedgerRequest(iscpChainID, signed)
	if err == nil {
		reqID = sc.cvt.ScRequestID(signed.ID())
	}
	return reqID, err
}

func (sc *WasmClientService) SubscribeEvents(msg chan []string, done chan bool) error {
	return subscribe.Subscribe(sc.eventPort, msg, done, false, "")
}

func (sc *WasmClientService) WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	iscpChainID := sc.cvt.IscpChainID(&chainID)
	iscpReqID := sc.cvt.IscpRequestID(&reqID)
	_, err := sc.waspClient.WaitUntilRequestProcessed(iscpChainID, *iscpReqID, timeout)
	return err
}
