// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
)

type SoloBackend struct {
	Env    *solo.Solo
	Chain  *solo.Chain
	pvtKey *cryptolib.KeyPair
}

var _ ChainBackend = &SoloBackend{}

func NewSoloBackend(env *solo.Solo, chain *solo.Chain, signer *cryptolib.KeyPair) *SoloBackend {
	return &SoloBackend{env, chain, signer}
}

func (s *SoloBackend) Signer() *cryptolib.KeyPair {
	return s.pvtKey
}

func (s *SoloBackend) gasFeeAssets(gas, feeAmount uint64, err error) (uint64, *iscp.FungibleTokens, error) {
	if err != nil {
		return 0, nil, err
	}
	gp := s.Chain.GetGasFeePolicy()
	return gas, iscp.NewFungibleTokensForGasFee(gp, feeAmount), nil
}

func (s *SoloBackend) EstimateGasOnLedger(scName, funName string, transfer *iscp.FungibleTokens, args dict.Dict) (uint64, *iscp.FungibleTokens, error) {
	return s.gasFeeAssets(s.Chain.EstimateGasOnLedger(
		solo.NewCallParamsFromDic(scName, funName, args).WithAssets(transfer),
		s.pvtKey,
		true,
	))
}

func (s *SoloBackend) PostOnLedgerRequest(scName, funName string, transfer *iscp.FungibleTokens, args dict.Dict, gasBudget uint64) error {
	_, err := s.Chain.PostRequestSync(
		solo.NewCallParamsFromDic(scName, funName, args).WithAssets(transfer).WithGasBudget(gasBudget),
		s.pvtKey,
	)
	return err
}

func (s *SoloBackend) EstimateGasOffLedger(scName, funName string, args dict.Dict) (uint64, *iscp.FungibleTokens, error) {
	return s.gasFeeAssets(s.Chain.EstimateGasOffLedger(
		solo.NewCallParamsFromDic(scName, funName, args),
		s.pvtKey,
		true,
	))
}

func (s *SoloBackend) PostOffLedgerRequest(scName, funName string, args dict.Dict, gasBudget uint64) error {
	_, err := s.Chain.PostRequestOffLedger(
		solo.NewCallParamsFromDic(scName, funName, args).WithGasBudget(gasBudget),
		s.pvtKey,
	)
	return err
}

func (s *SoloBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return s.Chain.CallView(scName, funName, args)
}
