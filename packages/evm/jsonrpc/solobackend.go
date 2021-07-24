// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp/color"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
)

type SoloBackend struct {
	Env    *solo.Solo
	Chain  *solo.Chain
	signer *ed25519.KeyPair
}

var _ ChainBackend = &SoloBackend{}

func NewSoloBackend(env *solo.Solo, chain *solo.Chain, signer *ed25519.KeyPair) *SoloBackend {
	return &SoloBackend{env, chain, signer}
}

func (s *SoloBackend) Signer() *ed25519.KeyPair {
	return s.signer
}

func (s *SoloBackend) PostOnLedgerRequest(scName, funName string, transfer color.Balances, args dict.Dict) error {
	_, err := s.Chain.PostRequestSync(
		solo.NewCallParamsFromDic(scName, funName, args).WithTransfers(transfer),
		s.signer,
	)
	return err
}

func (s *SoloBackend) PostOffLedgerRequest(scName, funName string, transfer color.Balances, args dict.Dict) error {
	_, err := s.Chain.PostRequestOffLedger(
		solo.NewCallParamsFromDic(scName, funName, args).WithTransfers(transfer),
		s.signer,
	)
	return err
}

func (s *SoloBackend) CallView(scName, funName string, args dict.Dict) (dict.Dict, error) {
	return s.Chain.CallView(scName, funName, args)
}
