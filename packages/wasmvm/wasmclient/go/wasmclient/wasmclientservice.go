// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

// for some reason we cannot use the import name mangos, so we rename those packages
// for some other reason if the third mamgos import is missing things won't work
import (
	"context"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/publisher/publisherws"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ContractEvent struct {
	ChainID    string
	ContractID string
	Data       string
}

type EventProcessor func(event *ContractEvent)

type IClientService interface {
	CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	CurrentChainID() wasmtypes.ScChainID
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (wasmtypes.ScRequestID, error)
	SubscribeEvents(callback EventProcessor) error
	UnsubscribeEvents()
	WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error
}

// WasmClientService TODO should be linked to a chain and holds the nonces for signers for that chain
type WasmClientService struct {
	callback   EventProcessor
	chainID    wasmtypes.ScChainID
	eventDone  chan bool
	nonceLock  sync.Mutex
	nonces     map[string]uint64
	waspClient *apiclient.APIClient
	webSocket  string
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

func (sc *WasmClientService) CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}

	res, _, err := sc.waspClient.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
		ChainId:       cvt.IscChainID(&sc.chainID).String(),
		ContractHName: cvt.IscHname(hContract).String(),
		FunctionHName: cvt.IscHname(hFunction).String(),
		Arguments:     apiextensions.JSONDictToAPIJSONDict(params.JSONDict()),
	}).Execute()
	if err != nil {
		return nil, err
	}

	decodedParams, err := apiextensions.APIJsonDictToDict(*res)
	if err != nil {
		return nil, err
	}

	return decodedParams.Bytes(), nil
}

func (sc *WasmClientService) CurrentChainID() wasmtypes.ScChainID {
	return sc.chainID
}

func (sc *WasmClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}

	nonce, err := sc.cachedNonce(keyPair)
	if err != nil {
		return reqID, err
	}

	req := isc.NewOffLedgerRequest(iscChainID, iscContract, iscFunction, params, nonce)
	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	signed := req.Sign(keyPair)
	reqID = cvt.ScRequestID(signed.ID())

	_, err = sc.waspClient.RequestsApi.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		ChainId: iscChainID.String(),
		Request: iotago.EncodeHex(signed.Bytes()),
	}).Execute()
	return reqID, err
}

func (sc *WasmClientService) SubscribeEvents(callback EventProcessor) error {
	// TODO multiple callbacks, see TS version
	sc.callback = callback

	ctx := context.Background()
	ws, _, err := websocket.Dial(ctx, sc.webSocket, nil)
	if err != nil {
		return err
	}
	err = eventSubscribe(ctx, ws, "chains")
	if err != nil {
		return err
	}
	err = eventSubscribe(ctx, ws, publisher.ISCEventKindSmartContract)
	if err != nil {
		return err
	}

	sc.eventDone = make(chan bool)
	go func() {
		<-sc.eventDone
		_ = ws.Close(websocket.StatusNormalClosure, "intentional close")
	}()

	go sc.eventLoop(ctx, ws)

	return nil
}

func (sc *WasmClientService) UnsubscribeEvents() {
	sc.eventDone <- true
}

func (sc *WasmClientService) WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	iscChainID := cvt.IscChainID(&sc.chainID)
	iscReqID := cvt.IscRequestID(&reqID)

	_, _, err := sc.waspClient.RequestsApi.
		WaitForRequest(context.Background(), iscChainID.String(), iscReqID.String()).
		TimeoutSeconds(int32(timeout.Seconds())).
		Execute()

	return err
}

func (sc *WasmClientService) cachedNonce(keyPair *cryptolib.KeyPair) (uint64, error) {
	sc.nonceLock.Lock()
	defer sc.nonceLock.Unlock()

	key := string(keyPair.GetPublicKey().AsBytes())
	nonce, ok := sc.nonces[key]
	if !ok {
		// note that even while getting the current nonce we keep the lock active
		// that way prevent other potential contenders to do the same in parallel
		iscAgent := isc.NewAgentID(keyPair.Address())
		agent := wasmtypes.AgentIDFromBytes(iscAgent.Bytes())
		ctx := NewWasmClientContext(sc, coreaccounts.ScName)
		n := coreaccounts.ScFuncs.GetAccountNonce(ctx)
		n.Params.AgentID().SetValue(agent)
		n.Func.Call()
		if ctx.Err != nil {
			return 0, ctx.Err
		}
		nonce = n.Results.AccountNonce().Value()
	}
	nonce++
	sc.nonces[key] = nonce
	return nonce, nil
}

func (sc *WasmClientService) eventLoop(ctx context.Context, ws *websocket.Conn) {
	for {
		evt := publisher.ISCEvent{}
		err := wsjson.Read(ctx, ws, &evt)
		if err != nil {
			sc.callback = nil
			return
		}
		items := evt.Content.([]interface{})
		for _, item := range items {
			parts := strings.Split(item.(string), ": ")
			event := ContractEvent{
				ChainID:    evt.ChainID,
				ContractID: parts[0],
				Data:       parts[1],
			}
			sc.callback(&event)
		}
	}
}

func eventSubscribe(ctx context.Context, ws *websocket.Conn, topic string) error {
	msg := publisherws.SubscriptionCommand{
		Command: publisherws.CommandSubscribe,
		Topic:   topic,
	}
	err := wsjson.Write(ctx, ws, msg)
	if err != nil {
		return err
	}
	return wsjson.Read(ctx, ws, &msg)
}
