// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
)

func TestRootBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	chain.CheckChain()
	chain.Log().Infof("\n%s\n", chain.String())
}

func TestRootRepeatInit(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()

	chain.CheckChain()

	req := solo.NewCallParams(root.Contract.Name, "init")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain()

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(corecontracts.All), len(contracts))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)
	recBlob, ok := contracts[blob.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)

	rec, err := chain.FindContract(blob.Contract.Name)
	require.NoError(t, err)
	require.EqualValues(t, recBlob.Bytes(), rec.Bytes())
}

func TestDeployExample(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10_000, nil)
	require.NoError(t, err)

	name := "testInc"
	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	chainID, ownerAgentID, contracts := ch.GetInfo()

	require.EqualValues(t, ch.ChainID, chainID)
	require.EqualValues(t, ch.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(corecontracts.All)+1, len(contracts))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)

	rec, ok := contracts[isc.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, sbtestsc.Contract.ProgramHash, rec.ProgramHash)

	recFind, err := ch.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, recFind.Bytes(), rec.Bytes())
}

func TestDeployDouble(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain()

	err := ch.DepositBaseTokensToL2(10_000, nil)
	require.NoError(t, err)

	name := "testInc"
	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.Error(t, err)

	chainID, ownerAgentID, contracts := ch.GetInfo()

	require.EqualValues(t, ch.ChainID, chainID)
	require.EqualValues(t, ch.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(corecontracts.All)+1, len(contracts))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)

	rec, ok := contracts[isc.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, sbtestsc.Contract.ProgramHash, rec.ProgramHash)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{Debug: true, PrintStackTrace: true})
	chain := env.NewChain()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := isc.NewAgentID(ownerAddr)

	req := solo.NewCallParams(
		governance.Contract.Name, governance.FuncDelegateChainOwnership.Name,
		governance.ParamChainOwner, newOwnerAgentID,
	).WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncClaimChainOwnership.Name).
		WithGasBudget(100_000).
		AddBaseTokens(100_000)

	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	_, ownerAgentID, _ = chain.GetInfo()
	require.True(t, newOwnerAgentID.Equals(ownerAgentID))
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustStorageDeposit: true})
	chain := env.NewChain()

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := isc.NewAgentID(ownerAddr)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name, governance.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequestSync(req, newOwner)
	require.Error(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
}
