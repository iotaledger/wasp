// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type IClientService interface {
	CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	CurrentChainID() wasmtypes.ScChainID
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (wasmtypes.ScRequestID, error)
	SubscribeEvents(eventHandler *WasmClientEvents) error
	UnsubscribeEvents(eventsID uint32)
	WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error
}

type WasmClientService struct {
	chainID       wasmtypes.ScChainID
	eventDone     chan bool
	eventHandlers []*WasmClientEvents
	nonceLock     sync.Mutex
	nonces        map[string]uint64
	waspClient    *apiclient.APIClient
	webSocket     string
}

var _ IClientService = new(WasmClientService)

func NewWasmClientService(waspAPI string, chainID string) *WasmClientService {
	err := SetSandboxWrappers(chainID)
	if err != nil {
		panic(err)
	}
	client, err := apiextensions.WaspAPIClientByHostName(waspAPI)
	if err != nil {
		panic(err)
	}
	return &WasmClientService{
		chainID:    wasmtypes.ChainIDFromString(chainID),
		nonces:     make(map[string]uint64),
		waspClient: client,
		webSocket:  strings.Replace(waspAPI, "http:", "ws:", 1) + "/ws",
	}
}

func (svc *WasmClientService) CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}

	res, _, err := svc.waspClient.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
		ChainId:       cvt.IscChainID(&svc.chainID).String(),
		ContractHName: cvt.IscHname(hContract).String(),
		FunctionHName: cvt.IscHname(hFunction).String(),
		Arguments:     apiextensions.JSONDictToAPIJSONDict(params.JSONDict()),
	}).Execute()
	if err != nil {
		return nil, apiError(err)
	}

	decodedParams, err := apiextensions.APIJsonDictToDict(*res)
	if err != nil {
		return nil, err
	}

	return decodedParams.Bytes(), nil
}

func (svc *WasmClientService) CurrentChainID() wasmtypes.ScChainID {
	return svc.chainID
}

func (svc *WasmClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}

	nonce, err := svc.cachedNonce(keyPair)
	if err != nil {
		return reqID, err
	}

	req := isc.NewOffLedgerRequest(iscChainID, iscContract, iscFunction, params, nonce)
	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	signed := req.Sign(keyPair)
	reqID = cvt.ScRequestID(signed.ID())

	_, err = svc.waspClient.RequestsApi.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		ChainId: iscChainID.String(),
		Request: iotago.EncodeHex(signed.Bytes()),
	}).Execute()
	return reqID, apiError(err)
}

func (svc *WasmClientService) SubscribeEvents(eventHandler *WasmClientEvents) error {
	svc.eventHandlers = append(svc.eventHandlers, eventHandler)
	if len(svc.eventHandlers) != 1 {
		return nil
	}

	svc.eventDone = make(chan bool)
	return startEventLoop(svc.webSocket, svc.eventDone, &svc.eventHandlers)
}

func (svc *WasmClientService) UnsubscribeEvents(eventsID uint32) {
	svc.eventHandlers = RemoveHandler(svc.eventHandlers, eventsID)
	if len(svc.eventHandlers) == 0 {
		// stop event loop
		svc.eventDone <- true
		svc.eventDone = nil
	}
}

func (svc *WasmClientService) WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	iscChainID := cvt.IscChainID(&svc.chainID)
	iscReqID := cvt.IscRequestID(&reqID)

	_, _, err := svc.waspClient.RequestsApi.
		WaitForRequest(context.Background(), iscChainID.String(), iscReqID.String()).
		TimeoutSeconds(int32(timeout.Seconds())).
		Execute()

	return apiError(err)
}

func apiError(err error) error {
	if err != nil {
		reqErr, ok := err.(*apiclient.GenericOpenAPIError)
		if ok {
			err = errors.New(reqErr.Error() + ": " + string(reqErr.Body()))
		}
	}
	return err
}

func (svc *WasmClientService) cachedNonce(keyPair *cryptolib.KeyPair) (uint64, error) {
	svc.nonceLock.Lock()
	defer svc.nonceLock.Unlock()

	key := string(keyPair.GetPublicKey().AsBytes())
	nonce, ok := svc.nonces[key]
	if !ok {
		// note that even while getting the current nonce we keep the lock active
		// that way prevent other potential contenders to do the same in parallel
		iscAgent := isc.NewAgentID(keyPair.Address())
		agent := wasmtypes.AgentIDFromBytes(iscAgent.Bytes())
		ctx := NewWasmClientContext(svc, coreaccounts.ScName)
		n := coreaccounts.ScFuncs.GetAccountNonce(ctx)
		n.Params.AgentID().SetValue(agent)
		n.Func.Call()
		if ctx.Err != nil {
			return 0, ctx.Err
		}
		nonce = n.Results.AccountNonce().Value()
	}
	nonce++
	svc.nonces[key] = nonce
	return nonce, nil
}
