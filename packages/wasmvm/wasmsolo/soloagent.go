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
	"github.com/stretchr/testify/require"
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

func (a *SoloAgent) AgentID() *iscp.AgentID {
	return iscp.NewAgentID(a.address, a.hname)
}

func (a *SoloAgent) Balance(color ...wasmtypes.ScColor) uint64 {
	switch len(color) {
	case 0:
		return a.Env.L1Iotas(a.address)
	case 1:
		token := wasmhost.WasmConvertor{}.IscpColor(&color[0])
		return a.Env.L1NativeTokens(a.address, token).Uint64()
	default:
		require.Fail(a.Env.T, "too many color arguments")
		return 0
	}
}

func (a *SoloAgent) Mint(amount uint64) (wasmtypes.ScColor, error) {
	token, err := a.Env.MintTokens(a.Pair, amount)
	return wasmhost.WasmConvertor{}.ScColor(&token), err
}
