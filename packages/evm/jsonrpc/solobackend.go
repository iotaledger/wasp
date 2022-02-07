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

func (s *SoloBackend) PostOnLedgerRequest(scName, funName string, transfer *iscp.Assets, args dict.Dict) error {
	_, err := s.Chain.PostRequestSync(
		solo.NewCallParamsFromDic(scName, funName, args).WithAssets(transfer),
		s.pvtKey,
	)
	return err
}

func (s *SoloBackend) PostOffLedgerRequest(scName, funName string, transfer *iscp.Assets, args dict.Dict) error {
	_, err := s.Chain.PostRequestOffLedger(
		solo.NewCallParamsFromDic(scName, funName, args).WithAssets(transfer),
		s.pvtKey,
	)
	return err
}

func (s *SoloBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return s.Chain.CallView(scName, funName, args)
}
