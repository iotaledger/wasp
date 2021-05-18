// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/stretchr/testify/require"
)

func TestBlobRepeatInit(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	req := solo.NewCallParams(blob.Interface.Name, "init")
	_, err := chain.PostRequestSync(req, nil)
	require.Error(t, err)
}

func TestBlobUpload(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	binary := []byte("supposed to be wasm")
	hwasm, err := chain.UploadWasm(nil, binary)
	require.NoError(t, err)

	binBack, err := chain.GetWasmBinary(hwasm)
	require.NoError(t, err)

	require.EqualValues(t, binary, binBack)
}

func TestBlobUploadTwice(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	binary := []byte("supposed to be wasm")
	hwasm1, err := chain.UploadWasm(nil, binary)
	require.NoError(t, err)

	hwasm2, err := chain.UploadWasm(nil, binary)
	require.NoError(t, err)

	require.EqualValues(t, hwasm1, hwasm2)

	binBack, err := chain.GetWasmBinary(hwasm1)
	require.NoError(t, err)

	require.EqualValues(t, binary, binBack)
}

var wasmFile = "sbtests/sbtestsc/testcore_bg.wasm"

func TestDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testCore", hwasm)
	require.NoError(t, err)
}

func TestDeployWasm(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, "testCore", wasmFile)
	require.NoError(t, err)
}

func TestDeployRubbish(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	name := "testCore"
	_, err := chain.FindContract(name)
	require.Error(t, err)
	err = chain.DeployWasmContract(nil, name, "blob_deploy_test.go")
	require.Error(t, err)

	_, err = chain.FindContract(name)
	require.Error(t, err)
}

func TestListBlobs(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, "testCore", wasmFile)
	require.NoError(t, err)

	ret, err := chain.CallView(blob.Interface.Name, blob.FuncListBlobs)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(ret))
}

func TestDeployNotAuthorized(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	user1, _ := env.NewKeyPairWithFunds()
	err := chain.DeployWasmContract(user1, "testCore", wasmFile)
	require.Error(t, err)
}

func TestDeployGrant(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	user1, addr1 := env.NewKeyPairWithFunds()
	user1AgentID := coretypes.NewAgentID(addr1, 0)

	req := solo.NewCallParams(root.Interface.Name, root.FuncGrantDeployPermission,
		root.ParamDeployer, user1AgentID,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testCore", wasmFile)
	require.NoError(t, err)

	_, _, contacts := chain.GetInfo()
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contacts))

	err = chain.DeployWasmContract(user1, "testInccounter2", wasmFile)
	require.NoError(t, err)

	_, _, contacts = chain.GetInfo()
	require.EqualValues(t, len(core.AllCoreContractsByHash)+2, len(contacts))
}

func TestRevokeDeploy(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	user1, addr1 := env.NewKeyPairWithFunds()
	user1AgentID := coretypes.NewAgentID(addr1, 0)

	req := solo.NewCallParams(root.Interface.Name, root.FuncGrantDeployPermission,
		root.ParamDeployer, user1AgentID,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testCore", wasmFile)
	require.NoError(t, err)

	_, _, contacts := chain.GetInfo()
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contacts))

	req = solo.NewCallParams(root.Interface.Name, root.FuncRevokeDeployPermission,
		root.ParamDeployer, user1AgentID,
	).WithIotas(1)
	_, err = chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testInccounter2", wasmFile)
	require.Error(t, err)

	_, _, contacts = chain.GetInfo()
	require.EqualValues(t, len(core.AllCoreContractsByHash)+1, len(contacts))
}

func TestDeployGrantFail(t *testing.T) {
	env := solo.New(t, false, false)
	chain := env.NewChain(nil, "chain1")
	user1, addr1 := env.NewKeyPairWithFunds()
	user1AgentID := coretypes.NewAgentID(addr1, 0)

	req := solo.NewCallParams(root.Interface.Name, root.FuncGrantDeployPermission,
		root.ParamDeployer, user1AgentID,
	).WithIotas(1)
	_, err := chain.PostRequestSync(req, user1)
	require.Error(t, err)

	err = chain.DeployWasmContract(user1, "testCore", wasmFile)
	require.Error(t, err)
}
