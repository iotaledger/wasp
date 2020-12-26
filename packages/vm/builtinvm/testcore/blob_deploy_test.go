// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testcore

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"testing"

	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/blob"
	"github.com/stretchr/testify/require"
)

func TestBlobRepeatInit(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	req := solo.NewCall(blob.Interface.Name, "init")
	_, err := chain.PostRequest(req, nil)
	require.Error(t, err)
}

func TestBlobUpload(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	binary := []byte("supposed to be wasm")
	hwasm, err := chain.UploadWasm(nil, binary)
	require.NoError(t, err)

	binBack, err := chain.GetWasmBinary(hwasm)
	require.NoError(t, err)

	require.EqualValues(t, binary, binBack)
}

func TestBlobUploadTwice(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
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

const wasmFile = "../../../../tools/cluster/tests/wasptest_new/wasm/inccounter_bg.wasm"

func TestDeploy(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	hwasm, err := chain.UploadWasmFromFile(nil, wasmFile)
	require.NoError(t, err)

	err = chain.DeployContract(nil, "testInccounter", hwasm)
	require.NoError(t, err)
}

func TestDeployWasm(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, "testInccounter", wasmFile)
	require.NoError(t, err)
}

func TestDeployRubbish(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	name := "testInccounter"
	err := chain.DeployWasmContract(nil, name, "blob_deploy_test.go")
	require.Error(t, err)

	_, err = chain.FindContract(name)
	require.Error(t, err)
}

func TestListBlobs(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	err := chain.DeployWasmContract(nil, "testInccounter", wasmFile)
	require.NoError(t, err)

	ret, err := chain.PostRequest(solo.NewCall(blob.Interface.Name, blob.FuncListBlobs), nil)
	require.NoError(t, err)
	require.EqualValues(t, 1, len(ret))
}

func TestDeployNotAuthorized(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	user1 := glb.NewSignatureSchemeWithFunds()
	err := chain.DeployWasmContract(user1, "testInccounter", wasmFile)
	require.Error(t, err)
}

func TestDeployGrant(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	user1 := glb.NewSignatureSchemeWithFunds()
	user1AgentID := coretypes.NewAgentIDFromAddress(user1.Address())

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, user1AgentID,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testInccounter", wasmFile)
	require.NoError(t, err)

	_, contacts := chain.GetInfo()
	require.EqualValues(t, 5, len(contacts))

	err = chain.DeployWasmContract(user1, "testInccounter2", wasmFile)
	require.NoError(t, err)

	_, contacts = chain.GetInfo()
	require.EqualValues(t, 6, len(contacts))
}

func TestRevokeDeploy(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	user1 := glb.NewSignatureSchemeWithFunds()
	user1AgentID := coretypes.NewAgentIDFromAddress(user1.Address())

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, user1AgentID,
	)
	_, err := chain.PostRequest(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testInccounter", wasmFile)
	require.NoError(t, err)

	_, contacts := chain.GetInfo()
	require.EqualValues(t, 5, len(contacts))

	req = solo.NewCall(root.Interface.Name, root.FuncRevokeDeploy,
		root.ParamDeployer, user1AgentID,
	)
	_, err = chain.PostRequest(req, nil)
	require.NoError(t, err)

	err = chain.DeployWasmContract(user1, "testInccounter2", wasmFile)
	require.Error(t, err)

	_, contacts = chain.GetInfo()
	require.EqualValues(t, 5, len(contacts))
}

func TestDeployGrantFail(t *testing.T) {
	glb := solo.New(t, false, false)
	chain := glb.NewChain(nil, "chain1")
	user1 := glb.NewSignatureSchemeWithFunds()
	user1AgentID := coretypes.NewAgentIDFromAddress(user1.Address())

	req := solo.NewCall(root.Interface.Name, root.FuncGrantDeploy,
		root.ParamDeployer, user1AgentID,
	)
	_, err := chain.PostRequest(req, user1)
	require.Error(t, err)

	err = chain.DeployWasmContract(user1, "testInccounter", wasmFile)
	require.Error(t, err)
}
