// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SoloAgent struct {
	Env     *solo.Solo
	Pair    *cryptolib.KeyPair
	address iotago.Address
	hname   iscp.Hname
}

func NewSoloAgent(env *solo.Solo) *SoloAgent {
	agent := &SoloAgent{Env: env}
	agent.Pair, agent.address = agent.Env.NewKeyPairWithFunds()
	return agent
}

func (a *SoloAgent) ScAddress() wasmtypes.ScAddress {
	return wasmhost.WasmConvertor{}.ScAddress(a.address)
}

func (a *SoloAgent) ScAgentID() wasmtypes.ScAgentID {
	return wasmtypes.NewScAgentID(a.ScAddress(), wasmtypes.ScHname(a.hname))
}

func (a *SoloAgent) Balance(color ...wasmtypes.ScColor) int64 {
	panic("fixme")
	//switch len(color) {
	//case 0:
	//	return int64(a.Env.GetAddressBalance(a.address, colored.IOTA))
	//case 1:
	//	col, err := colored.ColorFromBytes(color[0].Bytes())
	//	require.NoError(a.Env.T, err)
	//	return int64(a.Env.GetAddressBalance(a.address, col))
	//default:
	//	require.Fail(a.Env.T, "too many color arguments")
	//	return 0
	//}
}

func (a *SoloAgent) Mint(amount int64) (wasmtypes.ScColor, error) {
	color, err := a.Env.MintTokens(a.Pair, uint64(amount))
	return wasmhost.WasmConvertor{}.ScColor(color), err
}
