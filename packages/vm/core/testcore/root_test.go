// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestRootBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()
	chain.CheckChain()
}

func TestEntryPointNotFound(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	chain.CheckChain()

	req := solo.NewCallParamsEx(root.Contract.Name, "foo")
	_, err := chain.PostRequestOffLedger(req, nil)
	require.ErrorContains(t, err, "entry point not found")
}

func TestGetInfo(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	chainID, admin, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.AdminAgentID(), admin)
	require.GreaterOrEqual(t, len(contracts), len(corecontracts.All))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)

	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)
}

func TestChangeAdminAuthorized(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	chain := env.NewChain()

	newAdmin, ownerAddr := env.NewKeyPairWithFunds()
	newAdminAgentID := isc.NewAddressAgentID(ownerAddr)

	req := solo.NewCallParams(
		governance.FuncDelegateChainAdmin.Message(newAdminAgentID),
	).WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	_, admin, _ := chain.GetInfo()
	require.EqualValues(t, chain.AdminAgentID(), admin)

	req = solo.NewCallParams(governance.FuncClaimChainAdmin.Message()).
		WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err = chain.PostRequestSync(req, newAdmin)
	require.NoError(t, err)

	_, admin, _ = chain.GetInfo()
	require.True(t, newAdminAgentID.Equals(admin))
}

func TestChangeAdminUnauthorized(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	newAdmin, ownerAddr := env.NewKeyPairWithFunds()
	newAdminAgentID := isc.NewAddressAgentID(ownerAddr)
	req := solo.NewCallParams(governance.FuncDelegateChainAdmin.Message(newAdminAgentID)).
		AddBaseTokens(100_000)
	_, err := chain.PostRequestSync(req, newAdmin)
	require.ErrorContains(t, err, "unauthorized")

	_, admin, _ := chain.GetInfo()
	require.EqualValues(t, chain.AdminAgentID(), admin)
}
