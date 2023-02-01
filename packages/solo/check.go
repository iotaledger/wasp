// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"bytes"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (ch *Chain) AssertL2NativeTokens(agentID isc.AgentID, nativeTokenID iotago.NativeTokenID, bal interface{}) {
	bals := ch.L2Assets(agentID)
	actualTokenBalance := bals.AmountNativeToken(nativeTokenID)
	expected := util.ToBigInt(bal)
	require.Truef(ch.Env.T,
		expected.Cmp(actualTokenBalance) == 0,
		"expected: %v, got: %v", expected.String(), actualTokenBalance.String(),
	)
}

func (ch *Chain) AssertL2BaseTokens(agentID isc.AgentID, bal uint64) {
	require.EqualValues(ch.Env.T, int(bal), int(ch.L2Assets(agentID).BaseTokens))
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {
	_, err := ch.CallView(governance.Contract.Name, governance.ViewGetChainInfo.Name)
	require.NoError(ch.Env.T, err)

	for _, c := range corecontracts.All {
		recFromState, err := ch.FindContract(c.Name)
		require.NoError(ch.Env.T, err)
		require.EqualValues(ch.Env.T, c.Name, recFromState.Name)
		require.EqualValues(ch.Env.T, c.Description, recFromState.Description)
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
	coreacc = isc.NewContractAgentID(ch.ChainID, blob.Contract.Hname())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
	coreacc = isc.NewContractAgentID(ch.ChainID, accounts.Contract.Hname())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
	require.True(ch.Env.T, ch.L2Assets(coreacc).IsEmpty())
}

func (ch *Chain) AssertL2TotalNativeTokens(nativeTokenID iotago.NativeTokenID, bal interface{}) {
	bals := ch.L2TotalAssets()
	require.True(ch.Env.T, util.ToBigInt(bal).Cmp(bals.AmountNativeToken(nativeTokenID)) == 0)
}

func (ch *Chain) AssertL2TotalBaseTokens(bal uint64) {
	baseTokens := ch.L2TotalBaseTokens()
	require.EqualValues(ch.Env.T, int(bal), int(baseTokens))
}

func (ch *Chain) AssertControlAddresses() {
	rec := ch.GetControlAddresses()
	require.True(ch.Env.T, rec.StateAddress.Equal(ch.StateControllerAddress))
	require.True(ch.Env.T, rec.GoverningAddress.Equal(ch.StateControllerAddress))
	require.EqualValues(ch.Env.T, 0, rec.SinceBlockIndex)
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

func (env *Solo) AssertL1BaseTokens(addr iotago.Address, expected uint64) {
	require.EqualValues(env.T, int(expected), int(env.L1BaseTokens(addr)))
}

func (env *Solo) AssertL1NativeTokens(addr iotago.Address, nativeTokenID iotago.NativeTokenID, expected interface{}) {
	require.True(env.T, env.L1NativeTokens(addr, nativeTokenID).Cmp(util.ToBigInt(expected)) == 0)
}

func (env *Solo) HasL1NFT(addr iotago.Address, id *iotago.NFTID) bool {
	accountNFTs := env.L1NFTs(addr)
	for outputID, nftOutput := range accountNFTs {
		nftID := nftOutput.NFTID
		if nftID.Empty() {
			nftID = iotago.NFTIDFromOutputID(outputID)
		}
		if bytes.Equal(nftID[:], id[:]) {
			return true
		}
	}
	return false
}

func (env *Solo) GetUnspentOutputs(addr iotago.Address) (iotago.OutputSet, iotago.OutputIDs) {
	return env.utxoDB.GetUnspentOutputs(addr)
}
