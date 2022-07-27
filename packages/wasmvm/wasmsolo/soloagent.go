// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/stretchr/testify/require"
)

type SoloAgent struct {
	Cvt     wasmhost.WasmConvertor
	Env     *solo.Solo
	Pair    *cryptolib.KeyPair
	agentID iscp.AgentID
}

func NewSoloAgent(env *solo.Solo) *SoloAgent {
	pair, address := env.NewKeyPairWithFunds()
	return &SoloAgent{
		Env:     env,
		Pair:    pair,
		agentID: iscp.NewAgentID(address),
	}
}

func (a *SoloAgent) ScAgentID() wasmtypes.ScAgentID {
	return a.Cvt.ScAgentID(a.agentID)
}

func (a *SoloAgent) AgentID() iscp.AgentID {
	return a.agentID
}

func (a *SoloAgent) Balance(tokenID ...wasmtypes.ScTokenID) uint64 {
	address, _ := iscp.AddressFromAgentID(a.agentID)
	if address == nil {
		require.Fail(a.Env.T, "agent is not a L1 address")
	}
	switch len(tokenID) {
	case 0:
		return a.Env.L1BaseTokens(address)
	case 1:
		token := a.Cvt.IscpTokenID(&tokenID[0])
		return a.Env.L1NativeTokens(address, token).Uint64()
	default:
		require.Fail(a.Env.T, "too many tokenID arguments")
		return 0
	}
}

func (a *SoloAgent) Mint(amount uint64) (wasmtypes.ScTokenID, error) {
	token, err := a.Env.MintTokens(a.Pair, amount)
	return a.Cvt.ScTokenID(&token), err
}
