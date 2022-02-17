// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/erc721/go/erc721"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	ctx := setup(t)
	require.NoError(t, ctx.ContractExists(erc721.ScName))
}

func TestMint(t *testing.T) {
	ctx := setup(t)
	owner := ctx.NewSoloAgent()
	tokenID := wasmtypes.HashFromBytes(owner.ScAgentID().Bytes()[:32])
	mint(ctx, owner, tokenID)
	require.NoError(t, ctx.Err)
}

func TestApprove(t *testing.T) {
	ctx := setup(t)
	owner := ctx.NewSoloAgent()
	tokenID := wasmtypes.HashFromBytes(owner.ScAgentID().Bytes()[:32])
	mint(ctx, owner, tokenID)
	require.NoError(t, ctx.Err)

	require.Nil(t, getApproved(t, ctx, tokenID))

	approve(ctx, owner, owner, tokenID)
	require.Error(t, ctx.Err)

	approved := getApproved(t, ctx, tokenID)
	require.Nil(t, approved)

	friend1 := ctx.NewSoloAgent()
	approve(ctx, owner, friend1, tokenID)
	require.NoError(t, ctx.Err)

	approved = getApproved(t, ctx, tokenID)
	require.NotNil(t, approved)
	require.EqualValues(t, *approved, friend1.ScAgentID())

	approve(ctx, owner, nil, tokenID)
	require.NoError(t, ctx.Err)

	approved = getApproved(t, ctx, tokenID)
	require.Nil(t, approved)
}

func TestApproveAll(t *testing.T) {
	ctx := setup(t)
	owner := ctx.NewSoloAgent()
	tokenID := wasmtypes.HashFromBytes(owner.ScAgentID().Bytes()[:32])
	mint(ctx, owner, tokenID)
	require.NoError(t, ctx.Err)

	friend1 := ctx.NewSoloAgent()
	require.False(t, isApprovedForAll(t, ctx, owner, friend1))

	friend2 := ctx.NewSoloAgent()
	require.False(t, isApprovedForAll(t, ctx, owner, friend2))

	// approve friend1
	setApprovalForAll(ctx, owner, friend1, true)
	require.NoError(t, ctx.Err)

	require.True(t, isApprovedForAll(t, ctx, owner, friend1))
	require.False(t, isApprovedForAll(t, ctx, owner, friend2))

	// approve friend2
	setApprovalForAll(ctx, owner, friend2, true)
	require.NoError(t, ctx.Err)

	require.True(t, isApprovedForAll(t, ctx, owner, friend1))
	require.True(t, isApprovedForAll(t, ctx, owner, friend2))

	// unapprove friend1
	setApprovalForAll(ctx, owner, friend1, false)
	require.NoError(t, ctx.Err)

	require.False(t, isApprovedForAll(t, ctx, owner, friend1))
	require.True(t, isApprovedForAll(t, ctx, owner, friend2))

	// unapprove friend2
	setApprovalForAll(ctx, owner, friend2, false)
	require.NoError(t, ctx.Err)

	require.False(t, isApprovedForAll(t, ctx, owner, friend1))
	require.False(t, isApprovedForAll(t, ctx, owner, friend2))
}

func TestTransferFrom(t *testing.T) {
	ctx := setup(t)
	owner := ctx.NewSoloAgent()
	tokenID := wasmtypes.HashFromBytes(owner.ScAgentID().Bytes()[:32])
	mint(ctx, owner, tokenID)
	require.NoError(t, ctx.Err)

	// verify current ownership
	currentOwner := ownerOf(t, ctx, tokenID)
	require.EqualValues(t, owner.ScAgentID(), currentOwner)

	// no one approved for token
	require.Nil(t, getApproved(t, ctx, tokenID))

	friend1 := ctx.NewSoloAgent()

	// try to transfer without approval, should fail
	transferFrom(ctx, friend1, owner, friend1, tokenID)
	require.Error(t, ctx.Err)

	// verify current ownership has not changed
	currentOwner = ownerOf(t, ctx, tokenID)
	require.EqualValues(t, owner.ScAgentID(), currentOwner)

	// have owner himself transfer token, should succeed
	transferFrom(ctx, owner, owner, friend1, tokenID)
	require.NoError(t, ctx.Err)

	// verify new owner
	currentOwner = ownerOf(t, ctx, tokenID)
	require.EqualValues(t, friend1.ScAgentID(), currentOwner)

	// have previous try to transfer token back, should fail
	transferFrom(ctx, owner, friend1, owner, tokenID)
	require.Error(t, ctx.Err)

	// verify new owner is still owner
	currentOwner = ownerOf(t, ctx, tokenID)
	require.EqualValues(t, friend1.ScAgentID(), currentOwner)

	// have new owner transfer token back, should succeed
	transferFrom(ctx, friend1, friend1, owner, tokenID)
	require.NoError(t, ctx.Err)

	// verify ownership has returned to first owner
	currentOwner = ownerOf(t, ctx, tokenID)
	require.EqualValues(t, owner.ScAgentID(), currentOwner)
}

func setup(t *testing.T) *wasmsolo.SoloContext {
	init := erc721.ScFuncs.Init(nil)
	init.Params.Name().SetValue("My Valuable NFT")
	init.Params.Symbol().SetValue("MVNFT")
	ctx := wasmsolo.NewSoloContext(t, erc721.ScName, erc721.OnLoad, init.Func)
	require.NoError(t, ctx.Err)
	return ctx
}

func approve(ctx *wasmsolo.SoloContext, owner, approved *wasmsolo.SoloAgent, tokenID wasmtypes.ScHash) {
	f := erc721.ScFuncs.Approve(ctx.Sign(owner))
	if approved != nil {
		f.Params.Approved().SetValue(approved.ScAgentID())
	}
	f.Params.TokenID().SetValue(tokenID)
	f.Func.TransferIotas(1).Post()
}

func getApproved(t *testing.T, ctx *wasmsolo.SoloContext, tokenID wasmtypes.ScHash) *wasmtypes.ScAgentID {
	v := erc721.ScFuncs.GetApproved(ctx)
	v.Params.TokenID().SetValue(tokenID)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	approved := v.Results.Approved()
	if !approved.Exists() {
		return nil
	}
	ret := approved.Value()
	return &ret
}

func isApprovedForAll(t *testing.T, ctx *wasmsolo.SoloContext, owner, friend *wasmsolo.SoloAgent) bool {
	v := erc721.ScFuncs.IsApprovedForAll(ctx)
	v.Params.Owner().SetValue(owner.ScAgentID())
	v.Params.Operator().SetValue(friend.ScAgentID())
	v.Func.Call()
	require.NoError(t, ctx.Err)
	return v.Results.Approval().Value()
}

func mint(ctx *wasmsolo.SoloContext, owner *wasmsolo.SoloAgent, tokenID wasmtypes.ScHash) {
	f := erc721.ScFuncs.Mint(ctx.Sign(owner))
	f.Params.TokenID().SetValue(tokenID)
	f.Func.TransferIotas(1).Post()
}

func ownerOf(t *testing.T, ctx *wasmsolo.SoloContext, tokenID wasmtypes.ScHash) wasmtypes.ScAgentID {
	v := erc721.ScFuncs.OwnerOf(ctx)
	v.Params.TokenID().SetValue(tokenID)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	return v.Results.Owner().Value()
}

func setApprovalForAll(ctx *wasmsolo.SoloContext, owner, operator *wasmsolo.SoloAgent, approval bool) {
	f := erc721.ScFuncs.SetApprovalForAll(ctx.Sign(owner))
	f.Params.Operator().SetValue(operator.ScAgentID())
	f.Params.Approval().SetValue(approval)
	f.Func.TransferIotas(1).Post()
}

func transferFrom(ctx *wasmsolo.SoloContext, sender, from, to *wasmsolo.SoloAgent, tokenID wasmtypes.ScHash) {
	f := erc721.ScFuncs.TransferFrom(ctx.Sign(sender))
	f.Params.From().SetValue(from.ScAgentID())
	f.Params.To().SetValue(to.ScAgentID())
	f.Params.TokenID().SetValue(tokenID)
	f.Func.TransferIotas(1).Post()
}
