// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"math/big"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAccounts(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestDeposit(t *testing.T) {
	ctx := setupAccounts(t)

	user := ctx.NewSoloAgent()
	f := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestTransferAllowanceTo(t *testing.T) {
	ctx := setupAccounts(t)

	var transferAmount uint64 = 10_000
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()
	balanceOldUser0 := user0.Balance()
	balanceOldUser1 := user1.Balance()

	f := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user0))
	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Params.ForceOpenAccount().SetValue(false)
	f.Func.TransferIotas(transferAmount).Post()
	require.NoError(t, ctx.Err)

	balanceNewUser0 := user0.Balance()
	balanceNewUser1 := user1.Balance()
	assert.Equal(t, balanceOldUser0-transferAmount, balanceNewUser0)
	assert.Equal(t, balanceOldUser1+transferAmount, balanceNewUser1)

	// FIXME transfer other native tokens
}

func TestWithdraw(t *testing.T) {
	ctx := setupAccounts(t)
	var withdrawAmount uint64 = 10_000

	user := ctx.NewSoloAgent()
	balanceOldUser := user.Balance()

	f := coreaccounts.ScFuncs.Withdraw(ctx.Sign(user))
	f.Func.TransferIotas(withdrawAmount).Post()
	require.NoError(t, ctx.Err)
	balanceNewUser := user.Balance()
	assert.Equal(t, balanceOldUser-withdrawAmount, balanceNewUser)
}

func TestHarvest(t *testing.T) {}

func TestFoundryCreateNew(t *testing.T) {
	ctx := setupAccounts(t)
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	var dustAllowance uint64 = 1

	user := ctx.NewSoloAgent()
	f := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	f.Params.TokenTag().SetValue(codec.EncodeTokenTag(iotago.TokenTag{}))
	f.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), f.Results.FoundrySN().Value())

	f = coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(2001),
		MeltedTokens:  big.NewInt(2002),
		MaximumSupply: big.NewInt(2003),
	}))
	f.Params.TokenTag().SetValue(codec.EncodeTokenTag(iotago.TokenTag{}))
	f.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	assert.Equal(t, uint32(2), f.Results.FoundrySN().Value())
}

func TestFoundryDestroy(t *testing.T) {
	ctx := setupAccounts(t)
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	var dustAllowance uint64 = 1

	user := ctx.NewSoloAgent()
	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	fnew.Params.TokenTag().SetValue(codec.EncodeTokenTag(iotago.TokenTag{}))
	fnew.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	fdes := coreaccounts.ScFuncs.FoundryDestroy(ctx)
	fdes.Params.FoundrySN().SetValue(1)
	fdes.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestFoundryModifySupply(t *testing.T) {}

func TestBalance(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()
	var mintAmount uint64 = 100
	foundry, err := ctx.NewSoloFoundry(mintAmount, user)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	f := coreaccounts.ScFuncs.Balance(ctx)
	f.Params.AgentID().SetValue(user.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	balance := f.Results.Balances().GetBigInt(foundry.TokenID()).Value().Uint64()
	assert.Equal(t, mintAmount, balance)
}
func TestTotalAssets(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()

	var mintAmount0, mintAmount1 uint64 = 101, 202
	foundry0, err := ctx.NewSoloFoundry(mintAmount0, user0)
	require.NoError(t, err)
	err = foundry0.Mint(mintAmount0)
	require.NoError(t, err)
	tokenID0 := foundry0.TokenID()
	foundry1, err := ctx.NewSoloFoundry(mintAmount1, user1)
	require.NoError(t, err)
	err = foundry1.Mint(mintAmount1)
	require.NoError(t, err)
	tokenID1 := foundry1.TokenID()

	f := coreaccounts.ScFuncs.TotalAssets(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	val0 := f.Results.Assets().GetBigInt(tokenID0).Value().Uint64()
	assert.Equal(t, mintAmount0, val0)
	val1 := f.Results.Assets().GetBigInt(tokenID1).Value().Uint64()
	assert.Equal(t, mintAmount1, val1)
}

func TestAccounts(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()

	var mintAmount0, mintAmount1 uint64 = 101, 202
	foundry0, err := ctx.NewSoloFoundry(mintAmount0, user0)
	require.NoError(t, err)
	err = foundry0.Mint(mintAmount0)
	require.NoError(t, err)
	foundry1, err := ctx.NewSoloFoundry(mintAmount1, user1)
	require.NoError(t, err)
	err = foundry1.Mint(mintAmount1)
	require.NoError(t, err)

	f := coreaccounts.ScFuncs.Accounts(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	exist0 := f.Results.AllAccounts().GetBool(user0.ScAgentID()).Value()
	assert.True(t, exist0)
	exist1 := f.Results.AllAccounts().GetBool(user1.ScAgentID()).Value()
	assert.True(t, exist1)
	exist2 := f.Results.AllAccounts().GetBool(ctx.NewSoloAgent().ScAgentID()).Value()
	assert.False(t, exist2)
}

func TestGetAccountNonce(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()
	f := coreaccounts.ScFuncs.TransferAllowanceTo(ctx)
	f.Params.AgentID().SetValue(user.ScAgentID())
	f.Params.ForceOpenAccount().SetValue(true)
	f.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestGetNativeTokenIDRegistry(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()

	var mintAmount0, mintAmount1 uint64 = 101, 202
	foundry0, err := ctx.NewSoloFoundry(mintAmount0, user0)
	require.NoError(t, err)
	err = foundry0.Mint(mintAmount0)
	require.NoError(t, err)
	tokenID0 := foundry0.TokenID()
	foundry1, err := ctx.NewSoloFoundry(mintAmount1, user1)
	require.NoError(t, err)
	err = foundry1.Mint(mintAmount1)
	require.NoError(t, err)
	tokenID1 := foundry1.TokenID()

	f := coreaccounts.ScFuncs.GetNativeTokenIDRegistry(ctx)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	exist0 := f.Results.Mapping().GetBool(tokenID0).Value()
	assert.True(t, exist0)
	exist1 := f.Results.Mapping().GetBool(tokenID1).Value()
	assert.True(t, exist1)
	exist2 := f.Results.Mapping().GetBool(wasmtypes.TokenIDFromString("2C1agnXFL6r7U7wSH4MfkWcuj7tJkxes8hEPVPbb5gvApYh3thqh")).Value()
	assert.False(t, exist2)
}

func TestFoundryOutput(t *testing.T) {}

func TestAccountNFTs(t *testing.T) {}

func TestNFTData(t *testing.T) {}
