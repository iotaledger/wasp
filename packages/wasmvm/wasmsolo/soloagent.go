// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmhost"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type SoloAgent struct {
	Cvt     wasmhost.WasmConvertor
	Env     *solo.Solo
	Pair    *cryptolib.KeyPair
	agentID isc.AgentID
}

func NewSoloAgent(env *solo.Solo) *SoloAgent {
	pair, address := env.NewKeyPairWithFunds()
	return &SoloAgent{
		Env:     env,
		Pair:    pair,
		agentID: isc.NewAgentID(address),
	}
}

func (a *SoloAgent) ScAgentID() wasmtypes.ScAgentID {
	return a.Cvt.ScAgentID(a.agentID)
}

func (a *SoloAgent) AgentID() isc.AgentID {
	return a.agentID
}

// The optional nativeTokenID parameter can be used to retrieve the balance for the specific token.
// When nativeTokenID is omitted, the base tokens balance is assumed.
func (a *SoloAgent) Balance(nativeTokenID ...wasmtypes.ScTokenID) uint64 {
	address, _ := isc.AddressFromAgentID(a.agentID)
	if address == nil {
		require.Fail(a.Env.T, "agent is not a L1 address")
	}
	switch len(nativeTokenID) {
	case 0:
		return a.Env.L1BaseTokens(address)
	case 1:
		token := a.Cvt.IscTokenID(&nativeTokenID[0])
		return a.Env.L1NativeTokens(address, token).Uint64()
	default:
		require.Fail(a.Env.T, "too many nativeTokenID arguments")
		return 0
	}
}
