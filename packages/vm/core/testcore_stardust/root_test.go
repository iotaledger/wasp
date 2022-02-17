// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/core/testcore_stardust/sbtests/sbtestsc"
	"github.com/stretchr/testify/require"
)

func TestRootBasic(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()
	chain.Log().Infof("\n%s\n", chain.String())
}

func TestRootRepeatInit(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	req := solo.NewCallParams(root.Contract.Name, "init")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContractsByHash), len(contracts))

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
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain(nil, "chain1")

	err := ch.DepositIotasToL2(10_000, nil)
	require.NoError(t, err)

	name := "testInc"
	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	chainID, ownerAgentID, contracts := ch.GetInfo()

	require.EqualValues(t, ch.ChainID, chainID)
	require.EqualValues(t, ch.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contracts))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)

	rec, ok := contracts[iscp.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.True(t, ch.OriginatorAgentID.Equals(rec.Creator))
	require.EqualValues(t, sbtestsc.Contract.ProgramHash, rec.ProgramHash)

	recFind, err := ch.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, recFind.Bytes(), rec.Bytes())
}

func TestDeployDouble(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true}).
		WithNativeContract(sbtestsc.Processor)
	ch := env.NewChain(nil, "chain1")

	err := ch.DepositIotasToL2(10_000, nil)
	require.NoError(t, err)

	name := "testInc"
	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.NoError(t, err)

	err = ch.DeployContract(nil, name, sbtestsc.Contract.ProgramHash)
	require.Error(t, err)

	chainID, ownerAgentID, contracts := ch.GetInfo()

	require.EqualValues(t, ch.ChainID, chainID)
	require.EqualValues(t, ch.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contracts))

	_, ok := contracts[root.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Contract.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Contract.Hname()]
	require.True(t, ok)

	rec, ok := contracts[iscp.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.True(t, ch.OriginatorAgentID.Equals(rec.Creator))
	require.EqualValues(t, sbtestsc.Contract.ProgramHash, rec.ProgramHash)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	env := solo.New(t)
	chain := env.NewChain(nil, "chain1")

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(ownerAddr, 0)

	req := solo.NewCallParams(
		governance.Contract.Name, governance.FuncDelegateChainOwnership.Name,
		string(governance.ParamChainOwner), newOwnerAgentID,
	).WithGasBudget(100_000).
		AddAssetsIotas(100_000)

	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)

	req = solo.NewCallParams(governance.Contract.Name, governance.FuncClaimChainOwnership.Name).
		WithGasBudget(100_000).
		AddAssetsIotas(100_000)

	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	_, ownerAgentID, _ = chain.GetInfo()
	require.True(t, newOwnerAgentID.Equals(ownerAgentID))
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true})
	chain := env.NewChain(nil, "chain1")

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := iscp.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(governance.Contract.Name, governance.FuncDelegateChainOwnership.Name, string(governance.ParamChainOwner), newOwnerAgentID)
	_, err := chain.PostRequestSync(req, newOwner)
	require.Error(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
}
