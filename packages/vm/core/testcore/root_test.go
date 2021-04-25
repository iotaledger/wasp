// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/testcore/sbtests/sbtestsc"
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/stretchr/testify/require"
)

func TestRootBasic(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()
	chain.Log.Infof("\n%s\n", chain.String())
}

func TestRootRepeatInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chain.CheckChain()

	req := solo.NewCallParams(root.Interface.Name, "init")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContracts), len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	recBlob, ok := contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Interface.Hname()]
	require.True(t, ok)

	rec, err := chain.FindContract(blob.Interface.Name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recBlob), root.EncodeContractRecord(rec))
}

func TestDeployExample(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	name := "testInc"
	err := chain.DeployContract(nil, name, sbtestsc.Interface.ProgramHash)
	require.NoError(t, err)

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContracts)+1, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.OwnerFee)
	require.True(t, chain.OriginatorAgentID.Equals(rec.Creator))
	require.EqualValues(t, sbtestsc.Interface.ProgramHash, rec.ProgramHash)

	recFind, err := chain.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recFind), root.EncodeContractRecord(rec))
}

func TestDeployDouble(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	name := "testInc"
	err := chain.DeployContract(nil, name, sbtestsc.Interface.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, name, sbtestsc.Interface.ProgramHash)
	require.Error(t, err)

	chainID, ownerAgentID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
	require.EqualValues(t, len(core.AllCoreContracts)+1, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accounts.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.OwnerFee)
	require.True(t, chain.OriginatorAgentID.Equals(rec.Creator))
	require.EqualValues(t, sbtestsc.Interface.ProgramHash, rec.ProgramHash)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	req.WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)

	req = solo.NewCallParams(root.Interface.Name, root.FuncClaimChainOwnership).WithIotas(1)
	_, err = chain.PostRequestSync(req, newOwner)
	require.NoError(t, err)

	_, ownerAgentID, _ = chain.GetInfo()
	require.True(t, newOwnerAgentID.Equals(&ownerAgentID))
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")

	newOwner, ownerAddr := env.NewKeyPairWithFunds()
	newOwnerAgentID := coretypes.NewAgentID(ownerAddr, 0)
	req := solo.NewCallParams(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequestSync(req, newOwner)
	require.Error(t, err)

	_, ownerAgentID, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerAgentID)
}
