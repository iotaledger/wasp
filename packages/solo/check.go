// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func (ch *Chain) AssertL2AccountNativeToken(agentID *iscp.AgentID, tokenID *iotago.NativeTokenID, bal interface{}) {
	bals := ch.L2AccountAssets(agentID)
	require.True(ch.Env.T, util.ToBigInt(bal).Cmp(bals.AmountNativeToken(tokenID)) == 0)
}

func (ch *Chain) AssertL2AccountIotas(agentID *iscp.AgentID, bal uint64) {
	require.Equal(ch.Env.T, int(bal), int(ch.L2AccountAssets(agentID).Iotas))
}

// CheckChain checks fundamental integrity of the chain
func (ch *Chain) CheckChain() {
	_, err := ch.CallView(governance.Contract.Name, governance.FuncGetChainInfo.Name)
	require.NoError(ch.Env.T, err)

	for _, rec := range core.AllCoreContractsByHash {
		recFromState, err := ch.FindContract(rec.Contract.Name)
		require.NoError(ch.Env.T, err)
		require.EqualValues(ch.Env.T, rec.Contract.Name, recFromState.Name)
		require.EqualValues(ch.Env.T, rec.Contract.Description, recFromState.Description)
		require.EqualValues(ch.Env.T, rec.Contract.ProgramHash, recFromState.ProgramHash)
		require.True(ch.Env.T, recFromState.Creator.IsNil())
	}
	ch.CheckAccountLedger()
}

// CheckAccountLedger check integrity of the on-chain ledger.
// Sum of all accounts must be equal to total assets
func (ch *Chain) CheckAccountLedger() {
	total := ch.L2TotalAssetsInAccounts()
	accs := ch.L2Accounts()
	sum := iscp.NewEmptyAssets()
	for i := range accs {
		acc := accs[i]
		sum.Add(ch.L2AccountAssets(acc))
	}
	require.True(ch.Env.T, total.Equals(sum))
	coreacc := iscp.NewAgentID(ch.ChainID.AsAddress(), root.Contract.Hname())
	require.True(ch.Env.T, ch.L2AccountAssets(coreacc).IsEmpty())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), blob.Contract.Hname())
	require.True(ch.Env.T, ch.L2AccountAssets(coreacc).IsEmpty())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), accounts.Contract.Hname())
	require.True(ch.Env.T, ch.L2AccountAssets(coreacc).IsEmpty())
	require.True(ch.Env.T, ch.L2AccountAssets(coreacc).IsEmpty())
}

func (ch *Chain) AssertL2TotalNativeTokens(tokenID *iotago.NativeTokenID, bal interface{}) {
	bals := ch.L2TotalAssetsInAccounts()
	require.True(ch.Env.T, util.ToBigInt(bal).Cmp(bals.AmountNativeToken(tokenID)) == 0)
}

func (ch *Chain) AssertL2TotalIotas(bal uint64) {
	iotas := ch.L2TotalIotasInAccounts()
	require.EqualValues(ch.Env.T, int(bal), int(iotas))
}

func (ch *Chain) AssertControlAddresses() {
	rec := ch.GetControlAddresses()
	require.True(ch.Env.T, rec.StateAddress.Equal(ch.StateControllerAddress))
	require.True(ch.Env.T, rec.GoverningAddress.Equal(ch.StateControllerAddress))
	require.EqualValues(ch.Env.T, 0, rec.SinceBlockIndex)
}

func (env *Solo) AssertL1AddressIotas(addr iotago.Address, expected uint64) {
	require.EqualValues(env.T, int(expected), int(env.L1IotaBalance(addr)))
}

func (env *Solo) AssertL1NativeTokens(addr iotago.Address, tokenID *iotago.NativeTokenID, expected interface{}) {
	require.True(env.T, env.L1NativeTokenBalance(addr, tokenID).Cmp(util.ToBigInt(expected)) == 0)
}
