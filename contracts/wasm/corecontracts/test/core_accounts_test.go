// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"math/big"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
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

	bal := ctx.Balances(user0, user1)

	f := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.OffLedger(user0))
	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Params.ForceOpenAccount().SetValue(false)
	f.Func.AllowanceIotas(transferAmount).Post()
	require.NoError(t, ctx.Err)

	// note: transfer took place on L2, so no change on L1
	balanceNewUser0 := user0.Balance()
	balanceNewUser1 := user1.Balance()
	assert.Equal(t, balanceOldUser0, balanceNewUser0)
	assert.Equal(t, balanceOldUser1, balanceNewUser1)

	// expected changes to L2, note that caller pays the gas fee
	bal.Chain += ctx.GasFee
	bal.Add(user0, -transferAmount-ctx.GasFee)
	bal.Add(user1, transferAmount)
	bal.VerifyBalances(t)

	// FIXME transfer other native tokens
}

func TestWithdraw(t *testing.T) {
	ctx := setupAccounts(t)
	var withdrawAmount uint64 = 10_000

	user := ctx.NewSoloAgent()
	balanceOldUser := user.Balance()

	f := coreaccounts.ScFuncs.Withdraw(ctx.OffLedger(user))
	f.Func.AllowanceIotas(withdrawAmount).Post()
	require.NoError(t, ctx.Err)
	balanceNewUser := user.Balance()
	assert.Equal(t, balanceOldUser+withdrawAmount, balanceNewUser)
}

func TestHarvest(t *testing.T) {
	t.SkipNow()
	ctx := setupAccounts(t)
	var withdrawAmount uint64 = 10_000

	f := coreaccounts.ScFuncs.Harvest(ctx.Sign(ctx.Creator()))
	f.Func.TransferIotas(withdrawAmount).Post()
	require.NoError(t, ctx.Err)
}

func TestFoundryCreateNew(t *testing.T) {
	t.SkipNow()
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
	f.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	assert.Equal(t, uint32(2), f.Results.FoundrySN().Value())
}

func TestFoundryDestroy(t *testing.T) {
	t.SkipNow()
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
	fnew.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	fdes := coreaccounts.ScFuncs.FoundryDestroy(ctx)
	fdes.Params.FoundrySN().SetValue(1)
	fdes.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestFoundryModifySupply(t *testing.T) {
	t.SkipNow()
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
	fnew.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	fmod1 := coreaccounts.ScFuncs.FoundryModifySupply(ctx)
	fmod1.Params.FoundrySN().SetValue(1)
	fmod1.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod1.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)

	fmod2 := coreaccounts.ScFuncs.FoundryModifySupply(ctx)
	fmod2.Params.FoundrySN().SetValue(1)
	fmod2.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod2.Params.DestroyTokens().SetValue(true)
	fmod2.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
}

func TestBalance(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	// user1 := ctx.NewSoloAgent()

	var mintAmount uint64 = 100
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	f := coreaccounts.ScFuncs.Balance(ctx)
	f.Params.AgentID().SetValue(user0.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	balance := f.Results.Balances().GetBigInt(foundry.TokenID()).Value().Uint64()
	assert.Equal(t, mintAmount, balance)

	// FIXME complete this test
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
	user0 := ctx.NewSoloAgent()

	fnon := coreaccounts.ScFuncs.GetAccountNonce(ctx)
	fnon.Params.AgentID().SetValue(user0.ScAgentID())
	fnon.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint64(0), fnon.Results.AccountNonce().Value())

	ftrans := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.OffLedger(user0))
	ftrans.Params.AgentID().SetValue(user0.ScAgentID())
	ftrans.Params.ForceOpenAccount().SetValue(false)
	ftrans.Func.TransferIotas(1000).Post()
	require.NoError(t, ctx.Err)

	fnon.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint64(1), fnon.Results.AccountNonce().Value())
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
	notExistTokenID := wasmtypes.TokenIDFromString("2C1agnXFL6r7U7wSH4MfkWcuj7tJkxes8hEPVPbb5gvApYh3thqh")
	exist2 := f.Results.Mapping().GetBool(notExistTokenID).Value()
	assert.False(t, exist2)
}

func TestFoundryOutput(t *testing.T) {
	ctx := setupAccounts(t)
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	var dustAllowance uint64 = 1 * iscp.Mi

	user := ctx.NewSoloAgent()
	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	fnew.Func.TransferIotas(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	f := coreaccounts.ScFuncs.FoundryOutput(ctx)
	f.Params.FoundrySN().SetValue(1)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	b := f.Results.FoundryOutputBin().Value()
	fmt.Println("b: ", b)
}

func TestAccountNFTs(t *testing.T) {}

func TestNFTData(t *testing.T) {}
