// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type tHelper interface {
	Helper()
}

func (ch *Chain) AssertL2Coins(agentID isc.AgentID, coinType coin.Type, expected coin.Value) {
	if h, ok := ch.Env.T.(tHelper); ok {
		h.Helper()
	}
	bals := ch.L2Assets(agentID)
	actualTokenBalance := bals.Coins.Get(coinType)
	require.Equal(ch.Env.T, expected, actualTokenBalance)
}

func (ch *Chain) AssertL2BaseTokens(agentID isc.AgentID, bal coin.Value) {
	if h, ok := ch.Env.T.(tHelper); ok {
		h.Helper()
	}
	require.Equal(ch.Env.T, bal, ch.L2Assets(agentID).BaseTokens)
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {
	_, err := ch.CallView(governance.ViewGetChainInfo.Message())
	require.NoError(ch.Env.T, err)

	for _, c := range corecontracts.All {
		recFromState, err := ch.FindContract(c.Name)
		require.NoError(ch.Env.T, err)
		require.EqualValues(ch.Env.T, c.Name, recFromState.Name)
		require.EqualValues(ch.Env.T, c.ProgramHash, recFromState.ProgramHash)
	}
	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total ftokens
func (ch *Chain) CheckAccountLedger() {
	total := ch.L2TotalAssets()
	accs := ch.L2Accounts()
	sum := isc.NewEmptyAssets()
	for i := range accs {
		acc := accs[i]
		sum.Add(ch.L2Assets(acc))
	}
	require.True(ch.Env.T, total.Equals(sum))
	coreacc := isc.NewContractAgentID(ch.ChainID, root.Contract.Hname())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
	coreacc = isc.NewContractAgentID(ch.ChainID, accounts.Contract.Hname())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
}

func (ch *Chain) AssertL2TotalCoins(coinType coin.Type, bal coin.Value) {
	if h, ok := ch.Env.T.(tHelper); ok {
		h.Helper()
	}
	bals := ch.L2TotalAssets()
	require.Equal(ch.Env.T, bal, bals.Coins.Get(coinType))
}

func (ch *Chain) AssertL2TotalBaseTokens(bal uint64) {
	if h, ok := ch.Env.T.(tHelper); ok {
		h.Helper()
	}
	baseTokens := ch.L2TotalBaseTokens()
	require.EqualValues(ch.Env.T, int(bal), int(baseTokens))
}

func (ch *Chain) AssertControlAddresses() {
	if h, ok := ch.Env.T.(tHelper); ok {
		h.Helper()
	}
	rec := ch.GetControlAddresses()
	require.True(ch.Env.T, rec.StateAddress.Equals(ch.OriginatorAddress))
	require.True(ch.Env.T, rec.GoverningAddress.Equals(ch.OriginatorAddress))
	require.EqualValues(ch.Env.T, ch.LatestBlock().StateIndex(), rec.SinceBlockIndex)
}

func (ch *Chain) HasL2NFT(agentID isc.AgentID, nftID *iotago.NFTID) bool {
	accNFTIDs := ch.L2NFTs(agentID)
	for _, id := range accNFTIDs {
		if bytes.Equal(id[:], nftID[:]) {
			return true
		}
	}
	return false
}

func (env *Solo) AssertL1BaseTokens(addr *cryptolib.Address, expected coin.Value) {
	if h, ok := env.T.(tHelper); ok {
		h.Helper()
	}
	require.EqualValues(env.T, expected, env.L1BaseTokens(addr))
}

func (env *Solo) AssertL1Coins(addr *cryptolib.Address, coinType coin.Type, expected coin.Value) {
	if h, ok := env.T.(tHelper); ok {
		h.Helper()
	}
	require.True(env.T, env.L1CoinBalance(addr, coinType) == expected)
}

func (env *Solo) HasL1NFT(addr *cryptolib.Address, id sui.ObjectID) bool {
	accountNFTs := env.L1NFTs(addr)
	return lo.Contains(accountNFTs, id)
}
