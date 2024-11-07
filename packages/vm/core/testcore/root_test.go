// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/testutil/l1starter"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
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

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.GreaterOrEqual(t, len(contracts), len(corecontracts.All))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)

	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{
		Debug:           true,
		PrintStackTrace: true,
	})
	chain := env.NewChain()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := isc.NewAddressAgentID(ownerAddr)

	req := solo.NewCallParams(
		governance.FuncDelegateChainOwnership.Message(newOwnerAgentID),
	).WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)

	req = solo.NewCallParams(governance.FuncClaimChainOwnership.Message()).
		WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	_, ownerAgentID, _ = chain.GetInfo()
	require.True(t, newOwnerAgentID.Equals(ownerAgentID))
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := isc.NewAddressAgentID(ownerAddr)
	req := solo.NewCallParams(governance.FuncDelegateChainOwnership.Message(newOwnerAgentID)).
		AddBaseTokens(100_000)
	_, err := chain.PostRequestSync(req, newOwner)
	require.ErrorContains(t, err, "unauthorized")

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
}
