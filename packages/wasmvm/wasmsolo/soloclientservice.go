// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"errors"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SoloClientService struct {
	chainID       wasmtypes.ScChainID
	ctx           *SoloContext
	eventHandlers []*wasmclient.WasmClientEvents
	nonces        map[string]uint64
}

var _ wasmclient.IClientService = new(SoloClientService)

// NewSoloClientService creates a new SoloClientService
// Normally we reset the subscribers, assuming a new test.
// To prevent this when testing with multiple SoloClients,
// use the optional extra flag to indicate the extra clients.
func NewSoloClientService(ctx *SoloContext, chainID string, extra ...bool) *SoloClientService {
	s := &SoloClientService{
		chainID: wasmtypes.ChainIDFromString(chainID),
		ctx:     ctx,
		nonces:  make(map[string]uint64),
	}
	if len(extra) != 1 || !extra[0] {
		wasmhost.EventSubscribers = nil
	}
	wasmhost.EventSubscribers = append(wasmhost.EventSubscribers, func(msg string) {
		s.Event(msg)
	})
	return s
}

func (svc *SoloClientService) CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscChainID := cvt.IscChainID(&svc.chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	if !iscChainID.Equals(svc.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.CallViewByHname chain ID mismatch")
	}
	res, err := svc.ctx.Chain.CallViewByHname(iscContract, iscFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (svc *SoloClientService) CurrentChainID() wasmtypes.ScChainID {
	return svc.chainID
}

func (svc *SoloClientService) Event(msg string) {
	event := wasmclient.ContractEvent{
		ChainID:    svc.ctx.CurrentChainID(),
		ContractID: wasmtypes.NewScHname(svc.ctx.scName),
		Data:       msg,
	}
	for _, h := range svc.eventHandlers {
		h.ProcessEvent(&event)
	}
}

func (svc *SoloClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	if !iscChainID.Equals(svc.ctx.Chain.ChainID) {
		return reqID, errors.New("SoloClientService.PostRequest chain ID mismatch")
	}
	req := solo.NewCallParamsFromDictByHname(iscContract, iscFunction, params)

	key := string(keyPair.GetPublicKey().AsBytes())
	nonce := svc.nonces[key]
	nonce++
	svc.nonces[key] = nonce
	req.WithNonce(nonce)

	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	req.WithMaxAffordableGasBudget()
	_, err = svc.ctx.Chain.PostRequestOffLedger(req, keyPair)
	return reqID, err
}

func (svc *SoloClientService) SubscribeEvents(eventHandler *wasmclient.WasmClientEvents) error {
	svc.eventHandlers = append(svc.eventHandlers, eventHandler)
	return nil
}

func (svc *SoloClientService) UnsubscribeEvents(eventsID uint32) {
	svc.eventHandlers = wasmclient.RemoveHandler(svc.eventHandlers, eventsID)
}

func (svc *SoloClientService) WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	_ = reqID
	_ = timeout
	return nil
}
