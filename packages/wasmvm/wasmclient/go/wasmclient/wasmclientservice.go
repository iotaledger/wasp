// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient/iscclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type IClientService interface {
	CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	CurrentChainID() wasmtypes.ScChainID
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *iscclient.Keypair) (wasmtypes.ScRequestID, error)
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

func NewWasmClientService(waspAPI string) *WasmClientService {
	client, err := apiextensions.WaspAPIClientByHostName(waspAPI)
	if err != nil {
		panic(err)
	}
	return &WasmClientService{
		nonces:     make(map[string]uint64),
		waspClient: client,
		webSocket:  strings.Replace(waspAPI, "http:", "ws:", 1) + "/v1/ws",
	}
}

func (svc *WasmClientService) CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}

	res, _, err := svc.waspClient.ChainsApi.CallView(context.Background(), svc.chainID.String()).
		ContractCallViewRequest(apiclient.ContractCallViewRequest{
			ContractHName: hContract.String(),
			FunctionHName: hFunction.String(),
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

func (svc *WasmClientService) IsHealthy() bool {
	_, err := svc.waspClient.DefaultApi.
		GetHealth(context.Background()).Execute()
	return err == nil
}

func (svc *WasmClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *iscclient.Keypair) (reqID wasmtypes.ScRequestID, err error) {
	nonce, err := svc.cachedNonce(keyPair)
	if err != nil {
		return reqID, err
	}

	req, err := iscclient.NewOffLedgerRequest(chainID, hContract, hFunction, args, allowance, nonce)
	if err != nil {
		return reqID, err
	}

	req.Sign(keyPair)

	_, err = svc.waspClient.RequestsApi.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		ChainId: chainID.String(),
		Request: wasmtypes.HexEncode(req.Bytes()),
	}).Execute()
	return req.ID(), apiError(err)
}

func (svc *WasmClientService) SetCurrentChainID(chainID string) error {
	err := iscclient.SetSandboxWrappers(chainID)
	if err != nil {
		return err
	}

	svc.chainID = wasmtypes.ChainIDFromString(chainID)
	return nil
}

func (svc *WasmClientService) SetDefaultChainID() error {
	chains, _, err := svc.waspClient.ChainsApi.
		GetChains(context.Background()).Execute()
	if err != nil {
		return apiError(err)
	}
	if len(chains) != 1 {
		return errors.New("expected a single chain for default chain ID")
	}
	chainID := chains[0].ChainID
	fmt.Printf("default chain ID: %s\n", chainID)
	return svc.SetCurrentChainID(chainID)
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
	_, _, err := svc.waspClient.ChainsApi.
		WaitForRequest(context.Background(), svc.chainID.String(), reqID.String()).
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

func (svc *WasmClientService) cachedNonce(keyPair *iscclient.Keypair) (uint64, error) {
	svc.nonceLock.Lock()
	defer svc.nonceLock.Unlock()

	key := string(keyPair.GetPublicKey())
	nonce, ok := svc.nonces[key]
	if ok {
		svc.nonces[key] = nonce + 1
		return nonce, nil
	}

	agent := wasmtypes.ScAgentIDFromAddress(keyPair.Address())
	ctx := NewWasmClientContext(svc, coreaccounts.ScName)
	n := coreaccounts.ScFuncs.GetAccountNonce(ctx)
	n.Params.AgentID().SetValue(agent)
	n.Func.Call()
	if ctx.Err != nil {
		return 0, ctx.Err
	}
	nonce = n.Results.AccountNonce().Value()
	svc.nonces[key] = nonce + 1
	return nonce, nil
}
