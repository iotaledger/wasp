// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"os"
	"testing"

	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreblob"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreroot"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func setupRoot(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreroot.ScName, coreroot.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestDeployContract(t *testing.T) {
	ctxr := setupRoot(t)

	ctxb := ctxr.SoloContextForCore(t, coreblob.ScName, coreblob.OnLoad)
	require.NoError(t, ctxb.Err)
	fblob := coreblob.ScFuncs.StoreBlob(ctxb.OffLedger(ctxb.NewSoloAgent()))
	wasm, err := os.ReadFile("./corecontracts_bg.wasm")
	require.NoError(t, err)
	fblob.Params.ProgBinary().SetValue(wasm)
	fblob.Params.VmType().SetValue("wasmtime")
	fblob.Func.Post()
	require.NoError(t, ctxb.Err)

	fdeploy := coreroot.ScFuncs.DeployContract(ctxr)
	fdeploy.Params.ProgramHash().SetValue(fblob.Results.Hash().Value())
	fdeploy.Params.Name().SetValue("test_name")
	fdeploy.Params.Description().SetValue("this is desc")
	fdeploy.Func.Post()
	require.NoError(t, ctxr.Err)
}

func TestGrantDeployPermission(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	f := coreroot.ScFuncs.GrantDeployPermission(ctx)
	f.Params.Deployer().SetValue(user.ScAgentID())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestRevokeDeployPermission(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	f := coreroot.ScFuncs.RevokeDeployPermission(ctx)
	f.Params.Deployer().SetValue(user.ScAgentID())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestRequireDeployPermissions(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)

	f := coreroot.ScFuncs.RequireDeployPermissions(ctx)
	f.Params.DeployPermissionsEnabled().SetValue(true)
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestFindContract(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)

	f := coreroot.ScFuncs.FindContract(ctx)
	f.Params.Hname().SetValue(coreroot.HScName)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, []byte{0xff}, f.Results.ContractFound().Value())
	require.NotNil(t, f.Results.ContractRecData().Value())
	rbin := f.Results.ContractRecData().Value()
	record, err := root.ContractRecordFromBytes(rbin)
	require.NoError(t, err)
	require.Equal(t, root.Contract.ProgramHash, record.ProgramHash)
	require.Equal(t, coreroot.ScName, record.Name)
	require.Equal(t, coreroot.ScDescription, record.Description)
}

func TestGetContractRecords(t *testing.T) {
	ctx := setupRoot(t)
	require.NoError(t, ctx.Err)

	f := coreroot.ScFuncs.GetContractRecords(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	rbin := f.Results.ContractRegistry().GetBytes(coreroot.HScName).Value()
	record, err := root.ContractRecordFromBytes(rbin)
	require.NoError(t, err)
	require.Equal(t, root.Contract.ProgramHash, record.ProgramHash)
	require.Equal(t, coreroot.ScName, record.Name)
	require.Equal(t, coreroot.ScDescription, record.Description)
}
