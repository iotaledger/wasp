// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coregovernance"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func setupGovernance(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coregovernance.ScName, coregovernance.OnDispatch)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestRotateStateController(t *testing.T) {
	t.Skip("Chain.runRequestsNolock() hasn't been implemented yet")
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	fadd := coregovernance.ScFuncs.AddAllowedStateControllerAddress(ctx)
	fadd.Params.StateControllerAddress().SetValue(user.ScAgentID().Address())
	fadd.Func.Post()
	require.NoError(t, ctx.Err)

	frot := coregovernance.ScFuncs.RotateStateController(ctx)
	frot.Params.StateControllerAddress().SetValue(user.ScAgentID().Address())
	frot.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestAddAllowedStateControllerAddress(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	f := coregovernance.ScFuncs.AddAllowedStateControllerAddress(ctx)
	f.Params.StateControllerAddress().SetValue(user.ScAgentID().Address())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestRemoveAllowedStateControllerAddress(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	f := coregovernance.ScFuncs.RemoveAllowedStateControllerAddress(ctx)
	f.Params.StateControllerAddress().SetValue(user.ScAgentID().Address())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestClaimChainOwnership(t *testing.T) {
	t.SkipNow()
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	fdele := coregovernance.ScFuncs.DelegateChainOwnership(ctx)
	fdele.Params.ChainOwner().SetValue(user.ScAgentID())
	fdele.Func.Post()
	require.NoError(t, ctx.Err)

	fclaim := coregovernance.ScFuncs.ClaimChainOwnership(ctx)
	fclaim.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestDelegateChainOwnership(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	user := ctx.NewSoloAgent()
	f := coregovernance.ScFuncs.DelegateChainOwnership(ctx)
	f.Params.ChainOwner().SetValue(user.ScAgentID())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestSetFeePolicy(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	gfp0 := gas.DefaultFeePolicy()
	gfp0.GasPerToken = util.Ratio32{A: 1, B: 10}
	f := coregovernance.ScFuncs.SetFeePolicy(ctx)
	f.Params.FeePolicyBytes().SetValue(gfp0.Bytes())
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestAddCandidateNode(t *testing.T) {
	t.SkipNow()
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.AddCandidateNode(ctx)
	f.Params.AccessNodeInfoPubKey().SetValue(nil)
	f.Params.AccessNodeInfoCertificate().SetValue(nil)
	f.Params.AccessNodeInfoForCommittee().SetValue(false)
	f.Params.AccessNodeInfoAccessAPI().SetValue("")
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestRevokeAccessNode(t *testing.T) {
	t.SkipNow()
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.RevokeAccessNode(ctx)
	f.Params.AccessNodeInfoPubKey().SetValue(nil)
	f.Params.AccessNodeInfoCertificate().SetValue(nil)
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestChangeAccessNodes(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.ChangeAccessNodes(ctx)
	f.Params.ChangeAccessNodesActions()
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestGetChainOwner(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.GetChainOwner(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	assert.Equal(t, ctx.ChainOwnerID(), f.Results.ChainOwner().Value())
}

func TestGetFeePolicy(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.GetFeePolicy(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	fpBin := f.Results.FeePolicyBytes().Value()
	gfp, err := gas.FeePolicyFromBytes(fpBin)
	require.NoError(t, err)
	require.Equal(t, gas.DefaultGasPerToken, gfp.GasPerToken)
	require.Equal(t, uint8(0), gfp.ValidatorFeeShare) // default fee share is 0
}

func TestGetChainInfo(t *testing.T) {
	ctx := setupGovernance(t)
	require.NoError(t, ctx.Err)

	f := coregovernance.ScFuncs.GetChainInfo(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	assert.Equal(t, ctx.ChainOwnerID().String(), f.Results.ChainOwnerID().Value().String())
	gfp, err := gas.FeePolicyFromBytes(f.Results.GasFeePolicyBytes().Value())
	require.NoError(t, err)
	assert.Equal(t, ctx.Chain.GetGasFeePolicy(), gfp)
}
