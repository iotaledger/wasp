// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

type SoloAgent struct {
	Env     *solo.Solo
	KeyPair cryptolib.KeyPair
	address iotago.Address
	hname   iscp.Hname
}

func NewSoloAgent(env *solo.Solo) *SoloAgent {
	agent := &SoloAgent{Env: env}
	seed := cryptolib.NewSeed()
	agent.KeyPair = cryptolib.NewKeyPairFromSeed(seed)
	addr := cryptolib.Ed25519AddressFromPubKey(agent.KeyPair.PublicKey)
	agent.address = &addr
	return agent
}

func (a *SoloAgent) ScAddress() wasmlib.ScAddress {
	return wasmlib.NewScAddressFromBytes(iscp.BytesFromAddress(a.address))
}

func (a *SoloAgent) ScAgentID() wasmlib.ScAgentID {
	return wasmlib.NewScAgentID(a.ScAddress(), wasmlib.ScHname(a.hname))
}

func (a *SoloAgent) Balance(color ...wasmlib.ScColor) *big.Int {
	panic("TODO implement")
	// switch len(color) {
	// case 0:
	// 	return int64(a.Env.L1NativeTokenBalance(a.address, colored.IOTA))
	// case 1:
	// 	col, err := colored.ColorFromBytes(color[0].Bytes())
	// 	require.NoError(a.Env.T, err)
	// 	return int64(a.Env.L1NativeTokenBalance(a.address, col))
	// default:
	// 	require.Fail(a.Env.T, "too many color arguments")
	// 	return 0
	// }
}

func (a *SoloAgent) Mint(amount int64) (wasmlib.ScColor, error) {
	panic("TODO implement")
	// color, err := a.Env.MintTokens(a.Pair, uint64(amount))
	// return wasmlib.NewScColorFromBytes(color.Bytes()), err
}
