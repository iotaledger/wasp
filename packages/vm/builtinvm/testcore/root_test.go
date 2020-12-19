// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/examples/inccounter"
	"github.com/iotaledger/wasp/packages/vm/solo"
	"github.com/stretchr/testify/require"
)

func TestRootBasic(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	chain.CheckBase()
	chain.Infof("\n%s\n", chain.String())
}

func TestRootRepeatInit(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	chain.CheckBase()

	req := solo.NewCall(root.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}

func TestGetInfo(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	chainID, chainOwnerID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 4, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	recBlob, ok := contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, err := chain.FindContract(blob.Interface.Name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recBlob), root.EncodeContractRecord(rec))
}

func TestDeployExample(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	name := "testInc"
	err := chain.DeployContract(nil, name, inccounter.ProgramHash)
	require.NoError(t, err)

	chainID, chainOwnerID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 5, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.OwnerFee)
	require.EqualValues(t, chain.OriginatorAgentID, rec.Creator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)

	recFind, err := chain.FindContract(name)
	require.NoError(t, err)
	require.EqualValues(t, root.EncodeContractRecord(recFind), root.EncodeContractRecord(rec))
}

func TestDeployDouble(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	name := "testInc"
	err := chain.DeployContract(nil, name, inccounter.ProgramHash)
	require.NoError(t, err)

	err = chain.DeployContract(nil, name, inccounter.ProgramHash)
	require.Error(t, err)

	chainID, chainOwnerID, contracts := chain.GetInfo()

	require.EqualValues(t, chain.ChainID, chainID)
	require.EqualValues(t, chain.OriginatorAgentID, chainOwnerID)
	require.EqualValues(t, 5, len(contracts))

	_, ok := contracts[root.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[blob.Interface.Hname()]
	require.True(t, ok)
	_, ok = contracts[accountsc.Interface.Hname()]
	require.True(t, ok)

	rec, ok := contracts[coretypes.Hn(name)]
	require.True(t, ok)

	require.EqualValues(t, name, rec.Name)
	require.EqualValues(t, "N/A", rec.Description)
	require.EqualValues(t, 0, rec.OwnerFee)
	require.EqualValues(t, chain.OriginatorAgentID, rec.Creator)
	require.EqualValues(t, inccounter.ProgramHash, rec.ProgramHash)
}

func TestChangeOwnerAuthorized(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	_, ownerBack, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerBack)

	req = solo.NewCall(root.Interface.Name, root.FuncClaimChainOwnership)
	_, err = chain.PostRequest(req, newOwner)
	require.NoError(t, err)

	_, ownerBack, _ = chain.GetInfo()
	require.EqualValues(t, newOwnerAgentID, ownerBack)
}

func TestChangeOwnerUnauthorized(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	defer chain.WaitEmptyBacklog()

	newOwner := glb.NewSignatureSchemeWithFunds()
	newOwnerAgentID := coretypes.NewAgentIDFromAddress(newOwner.Address())
	req := solo.NewCall(root.Interface.Name, root.FuncDelegateChainOwnership, root.ParamChainOwner, newOwnerAgentID)
	_, err := chain.PostRequest(req, newOwner)
	require.Error(t, err)

	_, ownerBack, _ := chain.GetInfo()
	require.EqualValues(t, chain.OriginatorAgentID, ownerBack)
}
