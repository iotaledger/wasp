// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dustAllowance = 1 * isc.Million
	nftMetadata   = "NFT metadata"
)

func setupAccounts(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnLoad)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestDeposit(t *testing.T) {
	ctx := setupAccounts(t)

	const depositAmount = 1 * isc.Million
	user := ctx.NewSoloAgent()
	balanceOld := user.Balance()

	bal := ctx.Balances(user)

	f := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	f.Func.TransferBaseTokens(depositAmount).Post()
	require.NoError(t, ctx.Err)

	balanceNew := user.Balance()
	assert.Equal(t, balanceOld-depositAmount, balanceNew)

	// expected changes to L2, note that caller pays the gas fee
	bal.Chain += ctx.GasFee
	bal.Add(user, depositAmount-ctx.GasFee)
	bal.VerifyBalances(t)
}

func TestTransferAllowanceTo(t *testing.T) {
	ctx := setupAccounts(t)

	var transferAmountBaseTokens uint64 = 10_000
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()
	balanceOldUser0 := user0.Balance()
	balanceOldUser1 := user1.Balance()

	bal := ctx.Balances(user0, user1)

	f := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.OffLedger(user0))
	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Params.ForceOpenAccount().SetValue(false)
	f.Func.AllowanceBaseTokens(transferAmountBaseTokens).Post()
	require.NoError(t, ctx.Err)

	// note: transfer took place on L2, so no change on L1
	balanceNewUser0 := user0.Balance()
	balanceNewUser1 := user1.Balance()
	assert.Equal(t, balanceOldUser0, balanceNewUser0)
	assert.Equal(t, balanceOldUser1, balanceNewUser1)

	// expected changes to L2, note that caller pays the gas fee
	bal.Chain += ctx.GasFee
	bal.Add(user0, -transferAmountBaseTokens-ctx.GasFee)
	bal.Add(user1, transferAmountBaseTokens)
	bal.VerifyBalances(t)

	var mintAmount, transferAmount uint64 = 20_000, 10_000
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	scTokenID := foundry.TokenID()
	tokenID := ctx.Cvt.IscpTokenID(&scTokenID)

	balanceOldUser0 = user0.Balance(scTokenID)
	balanceOldUser1 = user1.Balance(scTokenID)
	balanceOldUser0L2 := ctx.Chain.L2NativeTokens(user0.AgentID(), tokenID)
	balanceOldUser1L2 := ctx.Chain.L2NativeTokens(user1.AgentID(), tokenID)

	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Params.ForceOpenAccount().SetValue(false)
	transfer := wasmlib.NewScTransfer()
	transfer.Set(&scTokenID, wasmtypes.NewScBigInt(transferAmount))
	f.Func.Allowance(transfer).Post()
	require.NoError(t, ctx.Err)

	// note: transfer took place on L2, so no change on L1
	balanceNewUser0 = user0.Balance(scTokenID)
	balanceNewUser1 = user1.Balance(scTokenID)
	balanceNewUser0L2 := ctx.Chain.L2NativeTokens(user0.AgentID(), tokenID)
	balanceNewUser1L2 := ctx.Chain.L2NativeTokens(user1.AgentID(), tokenID)
	assert.Equal(t, balanceOldUser0, balanceNewUser0)
	assert.Equal(t, balanceOldUser1, balanceNewUser1)
	assert.Equal(t, new(big.Int).Sub(balanceOldUser0L2, big.NewInt(int64(transferAmount))), balanceNewUser0L2)
	assert.Equal(t, new(big.Int).Add(balanceOldUser1L2, big.NewInt(int64(transferAmount))), balanceNewUser1L2)
}

func TestWithdraw(t *testing.T) {
	ctx := setupAccounts(t)
	var withdrawAmount uint64 = 10_000

	user := ctx.NewSoloAgent()
	balanceOldUser := user.Balance()

	f := coreaccounts.ScFuncs.Withdraw(ctx.OffLedger(user))
	f.Func.AllowanceBaseTokens(withdrawAmount).Post()
	require.NoError(t, ctx.Err)
	balanceNewUser := user.Balance()
	assert.Equal(t, balanceOldUser+withdrawAmount, balanceNewUser)
}

func TestHarvest(t *testing.T) {
	ctx := setupAccounts(t)
	var transferAmount, mintAmount uint64 = 10_000, 20_000
	var minimumBaseTokensOnCommonAccount uint64 = 3000

	user := ctx.NewSoloAgent()
	creatorAgentID := ctx.Creator().AgentID()
	commonAccount := ctx.Chain.CommonAccount()
	commonAccountBal0 := ctx.Chain.L2Assets(commonAccount)
	foundry, err := ctx.NewSoloFoundry(mintAmount, user)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	tokenID := foundry.TokenID()

	fTransfer0 := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user))
	fTransfer0.Params.AgentID().SetValue(ctx.Cvt.ScAgentID(commonAccount))
	fTransfer0.Func.AllowanceBaseTokens(transferAmount).Post()
	require.NoError(t, ctx.Err)
	fTransfer1 := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user))
	fTransfer1.Params.AgentID().SetValue(ctx.Cvt.ScAgentID(commonAccount))
	transfer := wasmlib.NewScTransfer()
	transfer.Set(&tokenID, wasmtypes.BigIntFromString(fmt.Sprint(transferAmount)))
	fTransfer1.Func.Allowance(transfer).Post()
	creatorBal0 := ctx.Chain.L2Assets(creatorAgentID)
	commonAccountBal1 := ctx.Chain.L2Assets(commonAccount)
	// create foundry, mint token, transfer BaseTokens and transfer token each charge GasFee, so there 4*GasFee in common account
	assert.Equal(t, commonAccountBal0.BaseTokens+transferAmount+ctx.GasFee*4, commonAccountBal1.BaseTokens)

	f := coreaccounts.ScFuncs.Harvest(ctx.Sign(ctx.Creator()))
	f.Func.Post()
	require.NoError(t, ctx.Err)
	commonAccountBal2 := ctx.Chain.L2Assets(commonAccount)
	creatorBal1 := ctx.Chain.L2Assets(creatorAgentID)
	assert.Equal(t, minimumBaseTokensOnCommonAccount+ctx.GasFee, commonAccountBal2.BaseTokens)
	assert.Equal(t, creatorBal0.BaseTokens+(commonAccountBal1.BaseTokens-commonAccountBal2.BaseTokens)+isc.Million, creatorBal1.BaseTokens)
	assert.Equal(t, big.NewInt(int64(transferAmount)), creatorBal1.Tokens[0].Amount)
}

func TestFoundryCreateNew(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()

	f := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	f.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), f.Results.FoundrySN().Value())

	f = coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(2001),
		MeltedTokens:  big.NewInt(2002),
		MaximumSupply: big.NewInt(2003),
	}))
	f.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	assert.Equal(t, uint32(2), f.Results.FoundrySN().Value())
}

func TestFoundryDestroy(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	fdes := coreaccounts.ScFuncs.FoundryDestroy(ctx)
	fdes.Params.FoundrySN().SetValue(1)
	fdes.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestFoundryNew(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	assert.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())
}

func TestFoundryModifySupply(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()

	mintAmount := wasmtypes.NewScBigInt(1000)
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)

	fmod1 := coreaccounts.ScFuncs.FoundryModifySupply(ctx.Sign(user0))
	fmod1.Params.FoundrySN().SetValue(1)
	fmod1.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod1.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)

	fmod2 := coreaccounts.ScFuncs.FoundryModifySupply(ctx.Sign(user0))
	fmod2.Params.FoundrySN().SetValue(foundry.SN())
	fmod2.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod2.Params.DestroyTokens().SetValue(true)
	tokenID := foundry.TokenID()
	allowance := wasmlib.NewScTransferTokens(&tokenID, wasmtypes.NewScBigInt(10))
	fmod2.Func.Allowance(allowance)
	fmod2.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
}

func TestBalance(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()

	mintAmount := wasmtypes.NewScBigInt(1000)
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	tokenID := foundry.TokenID()

	f := coreaccounts.ScFuncs.Balance(ctx)
	f.Params.AgentID().SetValue(user0.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	balance := f.Results.Balances().GetBigInt(foundry.TokenID()).Value()
	assert.Equal(t, mintAmount, balance)

	transferTokenAmount := wasmtypes.NewScBigInt(9)
	ftrans := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user0))
	ftrans.Params.AgentID().SetValue(user1.ScAgentID())
	transfer := wasmlib.NewScTransfer()
	transfer.Set(&tokenID, transferTokenAmount)
	ftrans.Func.Allowance(transfer).Post()
	require.NoError(t, ctx.Err)

	f.Params.AgentID().SetValue(user0.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	balance = f.Results.Balances().GetBigInt(foundry.TokenID()).Value()
	assert.Equal(t, mintAmount.Sub(transferTokenAmount), balance)
}

func TestTotalAssets(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent()
	user1 := ctx.NewSoloAgent()

	mintAmount0, mintAmount1 := wasmtypes.NewScBigInt(101), wasmtypes.NewScBigInt(202)
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
	val0 := f.Results.Assets().GetBigInt(tokenID0).Value()
	assert.Equal(t, mintAmount0, val0)
	val1 := f.Results.Assets().GetBigInt(tokenID1).Value()
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
	ftrans.Func.TransferBaseTokens(1000).Post()
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
	notExistTokenID := wasmtypes.TokenIDFromString("0x08f824508968d585ede1d154d34ba0d966ee03c928670fb85bd72e2924f67137890100000000")
	exist2 := f.Results.Mapping().GetBool(notExistTokenID).Value()
	assert.False(t, exist2)
}

func TestFoundryOutput(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need dust allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(dustAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	serialNum := uint32(1)
	assert.Equal(t, serialNum, fnew.Results.FoundrySN().Value())

	f := coreaccounts.ScFuncs.FoundryOutput(ctx)
	f.Params.FoundrySN().SetValue(1)
	f.Func.Call()
	require.NoError(t, ctx.Err)
	b := f.Results.FoundryOutputBin().Value()
	outFoundry := &iotago.FoundryOutput{}
	_, err := outFoundry.Deserialize(b, serializer.DeSeriModeNoValidation, nil)
	require.NoError(t, err)
	soloFoundry, err := ctx.Chain.GetFoundryOutput(serialNum)
	require.NoError(t, err)
	assert.Equal(t, soloFoundry, outFoundry)
}

func TestAccountNFTs(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()
	nftID := ctx.MintNFT(user, []byte(nftMetadata))
	userAddr, _ := isc.AddressFromAgentID(user.AgentID())

	require.True(t, ctx.Chain.Env.HasL1NFT(userAddr, ctx.Cvt.IscpNFTID(&nftID)))

	fd := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	transfer := wasmlib.NewScTransferNFT(&nftID)
	fd.Func.Transfer(transfer).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.Chain.HasL2NFT(user.AgentID(), ctx.Cvt.IscpNFTID(&nftID)))

	v := coreaccounts.ScFuncs.AccountNFTs(ctx)
	v.Params.AgentID().SetValue(user.ScAgentID())
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1, v.Results.NftIDs().Length())
	require.EqualValues(t, nftID, v.Results.NftIDs().GetNftID(0).Value())
}

func TestNFTData(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent()
	nftID := ctx.MintNFT(user, []byte(nftMetadata))
	userAddr, _ := isc.AddressFromAgentID(user.AgentID())

	iscpNFTID := ctx.Cvt.IscpNFTID(&nftID)
	require.True(t, ctx.Chain.Env.HasL1NFT(userAddr, iscpNFTID))

	fd := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	transfer := wasmlib.NewScTransferNFT(&nftID)
	fd.Func.Transfer(transfer).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.Chain.HasL2NFT(user.AgentID(), iscpNFTID))

	v := coreaccounts.ScFuncs.NftData(ctx)
	v.Params.NftID().SetValue(nftID)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	nftData, err := isc.NFTFromBytes(v.Results.NftData().Value())
	require.NoError(t, err)
	require.EqualValues(t, *iscpNFTID, nftData.ID)
	require.EqualValues(t, userAddr, nftData.Issuer)
	require.EqualValues(t, []byte(nftMetadata), nftData.Metadata)
}
