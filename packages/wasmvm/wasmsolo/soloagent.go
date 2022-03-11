// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	cryptolib "github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/stretchr/testify/require"
)

type SoloAgent struct {
	Cvt     wasmhost.WasmConvertor
	Env     *solo.Solo
	Pair    *cryptolib.KeyPair
	address ledgerstate.Address
	hname   iscp.Hname
}

func NewSoloAgent(env *solo.Solo) *SoloAgent {
	agent := &SoloAgent{Env: env}
	agent.Pair, agent.address = agent.Env.NewKeyPairWithFunds()
	return agent
}

func (a *SoloAgent) ScAddress() wasmtypes.ScAddress {
	return a.Cvt.ScAddress(a.address)
}

func (a *SoloAgent) ScAgentID() wasmtypes.ScAgentID {
	return wasmtypes.NewScAgentID(a.ScAddress(), wasmtypes.ScHname(a.hname))
}

func (a *SoloAgent) Balance(color ...wasmtypes.ScColor) uint64 {
	switch len(color) {
	case 0:
		return a.Env.GetAddressBalance(a.address, colored.IOTA)
	case 1:
		col := a.Cvt.IscpColor(&color[0])
		return a.Env.GetAddressBalance(a.address, col)
	default:
		require.Fail(a.Env.T, "too many color arguments")
		return 0
	}
}

func (a *SoloAgent) Mint(amount uint64) (wasmtypes.ScColor, error) {
	color, err := a.Env.MintTokens(a.Pair, amount)
	return a.Cvt.ScColor(&color), err
}
