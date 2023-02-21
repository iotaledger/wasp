// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"errors"
	"time"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmclient/go/wasmclient"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SoloClientService struct {
	chainID  wasmtypes.ScChainID
	ctx      *SoloContext
	callback wasmclient.EventProcessor
	nonces   map[string]uint64
}

var _ wasmclient.IClientService = new(SoloClientService)

// NewSoloClientService creates a new SoloClientService
// Normally we reset the subscribers, assuming a new test.
// To prevent this when testing with multiple SoloClients,
// use the optional extra flag to indicate the extra clients.
func NewSoloClientService(ctx *SoloContext, chainID string, extra ...bool) *SoloClientService {
	s := &SoloClientService{
		chainID:  wasmtypes.ChainIDFromString(chainID),
		ctx:      ctx,
		callback: nil,
		nonces:   make(map[string]uint64),
	}
	if len(extra) != 1 || !extra[0] {
		wasmhost.EventSubscribers = nil
	}
	wasmhost.EventSubscribers = append(wasmhost.EventSubscribers, func(msg string) {
		s.Event(msg)
	})
	return s
}

func (s *SoloClientService) CallViewByHname(hContract, hFunction wasmtypes.ScHname, args []byte) ([]byte, error) {
	iscChainID := cvt.IscChainID(&s.chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return nil, err
	}
	if !iscChainID.Equals(s.ctx.Chain.ChainID) {
		return nil, errors.New("SoloClientService.CallViewByHname chain ID mismatch")
	}
	res, err := s.ctx.Chain.CallViewByHname(iscContract, iscFunction, params)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (s *SoloClientService) CurrentChainID() wasmtypes.ScChainID {
	return s.chainID
}

func (s *SoloClientService) Event(msg string) {
	s.callback(&wasmclient.ContractEvent{
		ChainID:    s.ctx.CurrentChainID().String(),
		ContractID: isc.Hn(s.ctx.scName).String(),
		Data:       msg,
	})
}

func (s *SoloClientService) PostRequest(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, args []byte, allowance *wasmlib.ScAssets, keyPair *cryptolib.KeyPair) (reqID wasmtypes.ScRequestID, err error) {
	iscChainID := cvt.IscChainID(&chainID)
	iscContract := cvt.IscHname(hContract)
	iscFunction := cvt.IscHname(hFunction)
	params, err := dict.FromBytes(args)
	if err != nil {
		return reqID, err
	}
	if !iscChainID.Equals(s.ctx.Chain.ChainID) {
		return reqID, errors.New("SoloClientService.PostRequest chain ID mismatch")
	}
	req := solo.NewCallParamsFromDictByHname(iscContract, iscFunction, params)

	key := string(keyPair.GetPublicKey().AsBytes())
	nonce := s.nonces[key]
	nonce++
	s.nonces[key] = nonce
	req.WithNonce(nonce)

	iscAllowance := cvt.IscAllowance(allowance)
	req.WithAllowance(iscAllowance)
	req.WithGasBudget(gas.MaxGasPerRequest)
	_, err = s.ctx.Chain.PostRequestOffLedger(req, keyPair)
	return reqID, err
}

func (s *SoloClientService) SubscribeEvents(callback wasmclient.EventProcessor) error {
	s.callback = callback
	return nil
}

func (s *SoloClientService) UnsubscribeEvents() {
}

func (s *SoloClientService) WaitUntilRequestProcessed(reqID wasmtypes.ScRequestID, timeout time.Duration) error {
	_ = reqID
	_ = timeout
	return nil
}
