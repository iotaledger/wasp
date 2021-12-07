// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package solo

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

// AssertAddressBalance asserts the UTXODB address balance of specific color in the address
func (env *Solo) AssertAddressBalance(addr iotago.Address, assetID []byte, expected *big.Int) {
	require.Zero(env.T, expected.Cmp(env.GetAddressBalance(addr, assetID)))
}

func (env *Solo) AssertAddressIotas(addr iotago.Address, expected uint64) {
	expectedBig := big.NewInt(int64(expected))
	env.AssertAddressBalance(addr, iscp.IotaAssetID, expectedBig)
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
	total := ch.GetTotalAssets()
	accs := ch.GetAccounts()
	sum := iscp.NewEmptyAssets()
	for i := range accs {
		acc := accs[i]
		sum.Add(ch.GetAccountBalance(acc))
	}
	require.True(ch.Env.T, total.Equals(sum))
	coreacc := iscp.NewAgentID(ch.ChainID.AsAddress(), root.Contract.Hname())
	require.True(ch.Env.T, ch.GetAccountBalance(coreacc).IsEmpty())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), blob.Contract.Hname())
	require.True(ch.Env.T, ch.GetAccountBalance(coreacc).IsEmpty())
	coreacc = iscp.NewAgentID(ch.ChainID.AsAddress(), accounts.Contract.Hname())
	require.True(ch.Env.T, ch.GetAccountBalance(coreacc).IsEmpty())
	require.True(ch.Env.T, ch.GetAccountBalance(coreacc).IsEmpty())
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertAccountBalance(agentID *iscp.AgentID, assetID []byte, bal *big.Int) {
	bals := ch.GetAccountBalance(agentID)
	require.Zero(ch.Env.T, bal.Cmp(bals.AmountOf(assetID)))
}

func (ch *Chain) AssertIotas(agentID *iscp.AgentID, bal uint64) {
	require.Equal(ch.Env.T, bal, ch.GetAccountBalance(agentID).Iotas)
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertOwnersBalance(assetID []byte, bal *big.Int) {
	bals := ch.GetCommonAccountBalance()
	require.Zero(ch.Env.T, bal.Cmp(bals.AmountOf(assetID)))
}

func (ch *Chain) AssertCommonAccountIotas(bal uint64) {
	require.EqualValues(ch.Env.T, int(bal), int(ch.GetCommonAccountIotas()))
}

// AssertAccountBalance asserts the on-chain account balance controlled by agentID for specific color
func (ch *Chain) AssertTotalAssets(assetID []byte, bal *big.Int) {
	bals := ch.GetTotalAssets()
	require.Zero(ch.Env.T, bal.Cmp(bals.AmountOf(assetID)))
}

func (ch *Chain) AssertTotalIotas(bal uint64) {
	iotas := ch.GetTotalIotas()
	require.EqualValues(ch.Env.T, int(bal), int(iotas))
}

func (ch *Chain) CheckControlAddresses() {
	rec := ch.GetControlAddresses()
	require.True(ch.Env.T, rec.StateAddress.Equal(ch.StateControllerAddress))
	require.True(ch.Env.T, rec.GoverningAddress.Equal(ch.StateControllerAddress))
	require.EqualValues(ch.Env.T, 0, rec.SinceBlockIndex)
}
