// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/subscribe"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type IClientService interface {
	CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair, nonce uint64) (wasmtypes.ScRequestID, error)
	SubscribeEvents(msg chan []string, done chan bool) error
	WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error
}

type WasmClientService struct {
	cvt        wasmhost.WasmConvertor
	waspClient *clients.WaspClient
	eventPort  string
}

var _ IClientService = new(WasmClientService)

func NewWasmClientService(waspAPI, eventPort string) *WasmClientService {
	return &WasmClientService{waspClient: clients.NewWaspClient(waspAPI), eventPort: eventPort}
}

func DefaultWasmClientService() *WasmClientService {
	return NewWasmClientService("127.0.0.1:19090", "127.0.0.1:15550")
}

func (sc *WasmClientService) CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	res, err := sc.waspClient.CallViewByHname(iscChainID, iscContract, iscFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (sc *WasmClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair, nonce uint64) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	req := isc.NewOffLedgerRequest(iscChainID, iscContract, iscFunction, params, nonce)
	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	signed := req.Sign(keyPair)
	reqID = cvt.ScRequestID(signed.ID())
	err = sc.waspClient.PostOffLedgerRequest(iscChainID, signed)
	return reqID, err
}

func (sc *WasmClientService) SubscribeEvents(msg chan []string, done chan bool) error {
	return subscribe.Subscribe(sc.eventPort, msg, done, false, "contract")
}

func (sc *WasmClientService) WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	iscChainID := cvt.IscChainID(&chainID)
	iscReqID := cvt.IscRequestID(&reqID)
	_, err := sc.waspClient.WaitUntilRequestProcessed(iscChainID, iscReqID, timeout)
	return err
}
