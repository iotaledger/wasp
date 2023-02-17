// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"

	// for some reason we cannot use the name mangos, so we rename the packages
	nanomsg "go.nanomsg.org/mangos/v3"
	nanomsgsub "go.nanomsg.org/mangos/v3/protocol/sub"
	// for some reason if this import is missing things won't work
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

type ContractEvent struct {
	ChainID    string
	ContractID string
	Data       string
}

type IClientService interface {
	CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error)
	PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair, nonce uint64) (wasmtypes.ScRequestID, error)
	SubscribeEvents(msg chan ContractEvent, done chan bool) error
	WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error
}

type WasmClientService struct {
	waspClient *apiclient.APIClient
	eventPort  string
}

var _ IClientService = new(WasmClientService)

func NewWasmClientService(waspAPI, eventPort string) *WasmClientService {
	client, err := apiextensions.WaspAPIClientByHostName(waspAPI)
	if err != nil {
		panic(err.Error())
	}

	return &WasmClientService{waspClient: client, eventPort: eventPort}
}

func DefaultWasmClientService() *WasmClientService {
	return NewWasmClientService("http://localhost:19090", "127.0.0.1:15550")
}

func (sc *WasmClientService) CallViewByHname(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	res, _, err := sc.waspClient.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
		ContractHName: iscContract.String(),
		FunctionHName: iscFunction.String(),
		ChainId:       iscChainID.String(),
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

	_, err = sc.waspClient.RequestsApi.OffLedger(context.Background()).OffLedgerRequest(apiclient.OffLedgerRequest{
		ChainId: iscChainID.String(),
		Request: iotago.EncodeHex(signed.Bytes()),
	}).Execute()

	return reqID, err
}

func (sc *WasmClientService) SubscribeEvents(msg chan ContractEvent, done chan bool) error {
	socket, err := nanomsgsub.NewSocket()
	if err != nil {
		return err
	}
	err = socket.Dial("tcp://" + sc.eventPort)
	if err != nil {
		return fmt.Errorf("can't dial on sub socket %s: %w", sc.eventPort, err)
	}
	err = socket.SetOption(nanomsg.OptionSubscribe, []byte("contract"))
	if err != nil {
		return err
	}

	go func() {
		for {
			buf, err := socket.Recv()
			if err != nil {
				close(msg)
				return
			}
			if len(buf) > 0 {
				// contract tst1pqqf4qxh2w9x7rz2z4qqcvd0y8n22axsx82gqzmncvtsjqzwmhnjs438rhk | vm (contract): 89703a45: testwasmlib.test|1671671237|tst1pqqf4qxh2w9x7rz2z4qqcvd0y8n22axsx82gqzmncvtsjqzwmhnjs438rhk|Lala
				s := string(buf)
				parts := strings.Split(s, " ")
				event := ContractEvent{
					ChainID:    parts[1],
					ContractID: parts[5][:len(parts[5])-1],
					Data:       parts[6],
				}
				msg <- event
			}
		}
	}()

	go func() {
		<-done
		_ = socket.Close()
	}()

	return nil
}

func (sc *WasmClientService) WaitUntilRequestProcessed(chainID wasmtypes.ScChainID, reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	iscChainID := cvt.IscChainID(&chainID)
	iscReqID := cvt.IscRequestID(&reqID)

	_, _, err := sc.waspClient.RequestsApi.
		WaitForRequest(context.Background(), iscChainID.String(), iscReqID.String()).
		TimeoutSeconds(int32(timeout.Seconds())).
		Execute()

	return err
}
