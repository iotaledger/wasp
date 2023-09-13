// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/coreaccounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

const (
	sdAllowance = 1 * isc.Million
	nftMetadata = "NFT metadata"
)

func setupAccounts(t *testing.T) *wasmsolo.SoloContext {
	ctx := setup(t)
	ctx = ctx.SoloContextForCore(t, coreaccounts.ScName, coreaccounts.OnDispatch)
	require.NoError(t, ctx.Err)
	return ctx
}

func TestDeposit(t *testing.T) {
	ctx := setupAccounts(t)

	const depositAmount = 1 * isc.Million
	user := ctx.NewSoloAgent("user")
	balanceOld := user.Balance()

	bal := ctx.Balances(user)

	f := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	f.Func.TransferBaseTokens(depositAmount).Post()
	require.NoError(t, ctx.Err)

	balanceNew := user.Balance()
	require.Equal(t, balanceOld-depositAmount, balanceNew)

	// expected changes to L2, note that caller pays the gas fee
	bal.Originator += ctx.GasFee
	bal.Add(user, depositAmount-ctx.GasFee)
	bal.VerifyBalances(t)
}

func TestTransferAllowanceTo(t *testing.T) {
	ctx := setupAccounts(t)

	var transferAmountBaseTokens uint64 = 10_000
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")
	balanceOldUser0 := user0.Balance()
	balanceOldUser1 := user1.Balance()

	bal := ctx.Balances(user0, user1)

	f := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.OffLedger(user0))
	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Func.AllowanceBaseTokens(transferAmountBaseTokens).Post()
	require.NoError(t, ctx.Err)

	// note: transfer took place on L2, so no change on L1
	balanceNewUser0 := user0.Balance()
	balanceNewUser1 := user1.Balance()
	require.Equal(t, balanceOldUser0, balanceNewUser0)
	require.Equal(t, balanceOldUser1, balanceNewUser1)

	// expected changes to L2, note that caller pays the gas fee
	bal.Originator += ctx.GasFee
	bal.Add(user0, -transferAmountBaseTokens-ctx.GasFee)
	bal.Add(user1, transferAmountBaseTokens)
	bal.VerifyBalances(t)

	var mintAmount, transferAmount uint64 = 20_000, 10_000
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	scTokenID := foundry.TokenID()
	nativeTokenID := ctx.Cvt.IscTokenID(&scTokenID)

	balanceOldUser0 = user0.Balance(scTokenID)
	balanceOldUser1 = user1.Balance(scTokenID)
	balanceOldUser0L2 := ctx.Chain.L2NativeTokens(user0.AgentID(), nativeTokenID)
	balanceOldUser1L2 := ctx.Chain.L2NativeTokens(user1.AgentID(), nativeTokenID)

	f.Params.AgentID().SetValue(user1.ScAgentID())
	transfer := wasmlib.NewScTransfer()
	transfer.Set(&scTokenID, wasmtypes.NewScBigInt(transferAmount))
	f.Func.Allowance(transfer).Post()
	require.NoError(t, ctx.Err)

	// note: transfer took place on L2, so no change on L1
	balanceNewUser0 = user0.Balance(scTokenID)
	balanceNewUser1 = user1.Balance(scTokenID)
	balanceNewUser0L2 := ctx.Chain.L2NativeTokens(user0.AgentID(), nativeTokenID)
	balanceNewUser1L2 := ctx.Chain.L2NativeTokens(user1.AgentID(), nativeTokenID)
	require.Equal(t, balanceOldUser0, balanceNewUser0)
	require.Equal(t, balanceOldUser1, balanceNewUser1)
	require.Equal(t, new(big.Int).Sub(balanceOldUser0L2, big.NewInt(int64(transferAmount))), balanceNewUser0L2)
	require.Equal(t, new(big.Int).Add(balanceOldUser1L2, big.NewInt(int64(transferAmount))), balanceNewUser1L2)
}

func TestWithdraw(t *testing.T) {
	ctx := setupAccounts(t)
	var withdrawAmount uint64 = 10_000

	user := ctx.NewSoloAgent("user")
	balanceOldUser := user.Balance()

	f := coreaccounts.ScFuncs.Withdraw(ctx.OffLedger(user))
	f.Func.AllowanceBaseTokens(withdrawAmount).Post()
	require.NoError(t, ctx.Err)
	balanceNewUser := user.Balance()
	require.Equal(t, balanceOldUser+withdrawAmount, balanceNewUser)
}

func TestFoundryCreateNew(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")

	f := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need storage deposit allowance to keep foundry transaction not being trimmed by snapshot
	f.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	require.Equal(t, uint32(1), f.Results.FoundrySN().Value())

	f = coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	f.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(2001),
		MeltedTokens:  big.NewInt(2002),
		MaximumSupply: big.NewInt(2003),
	}))
	f.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint32(2), f.Results.FoundrySN().Value())
}

func TestFoundryDestroy(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need storage deposit allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	require.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())

	fdes := coreaccounts.ScFuncs.FoundryDestroy(ctx)
	fdes.Params.FoundrySN().SetValue(1)
	fdes.Func.Post()
	require.NoError(t, ctx.Err)
}

func TestFoundryNew(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need storage deposit allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	require.Equal(t, uint32(1), fnew.Results.FoundrySN().Value())
}

func TestFoundryModifySupply(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")

	mintAmount := wasmtypes.NewScBigInt(1000)
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)

	fmod1 := coreaccounts.ScFuncs.FoundryModifySupply(ctx.Sign(user0))
	fmod1.Params.FoundrySN().SetValue(1)
	fmod1.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod1.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)

	fmod2 := coreaccounts.ScFuncs.FoundryModifySupply(ctx.Sign(user0))
	fmod2.Params.FoundrySN().SetValue(foundry.SN())
	fmod2.Params.SupplyDeltaAbs().SetValue(wasmtypes.BigIntFromString("10"))
	fmod2.Params.DestroyTokens().SetValue(true)
	tokenID := foundry.TokenID()
	allowance := wasmlib.ScTransferFromTokens(&tokenID, wasmtypes.NewScBigInt(10))
	fmod2.Func.Allowance(allowance)
	fmod2.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
}

func TestAccountFoundries(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

	mintAmount := wasmtypes.NewScBigInt(1000)
	var foundries0 []*wasmsolo.SoloFoundry
	for i := 0; i < 10; i++ {
		foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
		require.NoError(t, err)
		foundries0 = append(foundries0, foundry)
		err = foundry.Mint(mintAmount)
		require.NoError(t, err)
	}

	var foundries1 []*wasmsolo.SoloFoundry
	for i := 0; i < 3; i++ {
		foundry, err := ctx.NewSoloFoundry(mintAmount, user1)
		require.NoError(t, err)
		foundries1 = append(foundries1, foundry)
		err = foundry.Mint(mintAmount)
		require.NoError(t, err)
	}

	f := coreaccounts.ScFuncs.AccountFoundries(ctx)
	f.Params.AgentID().SetValue(user0.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	for _, foundry := range foundries0 {
		require.True(t, f.Results.Foundries().GetBool(foundry.SN()).Value())
	}

	f.Params.AgentID().SetValue(user1.ScAgentID())
	f.Func.Call()
	require.NoError(t, ctx.Err)
	for _, foundry := range foundries1 {
		require.True(t, f.Results.Foundries().GetBool(foundry.SN()).Value())
	}
}

func TestAccountNFTAmount(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")
	userAddr, _ := isc.AddressFromAgentID(user.AgentID())
	nftNum := 7
	nftMetadataSlice := make([][]byte, nftNum)
	nftIDs := make([]wasmtypes.ScNftID, nftNum)
	for i := range nftMetadataSlice {
		randNftMetadata := make([]byte, 4)
		rand.Read(randNftMetadata)
		nftMetadataSlice[i] = randNftMetadata
		nftIDs[i] = ctx.MintNFT(user, randNftMetadata)
		require.NoError(t, ctx.Err)
		require.True(t, ctx.Chain.Env.HasL1NFT(userAddr, ctx.Cvt.IscNFTID(&nftIDs[i])))
	}

	fd := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	nftL2Num := 3
	for i := 0; i < nftL2Num; i++ {
		transfer := wasmlib.ScTransferFromNFT(&nftIDs[i])
		fd.Func.Transfer(transfer).Post()
		require.NoError(t, ctx.Err)
		require.True(t, ctx.Chain.HasL2NFT(user.AgentID(), ctx.Cvt.IscNFTID(&nftIDs[i])))
	}

	v := coreaccounts.ScFuncs.AccountNFTAmount(ctx)
	v.Params.AgentID().SetValue(user.ScAgentID())
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, uint32(nftL2Num), v.Results.Amount().Value())
}

func TestAccountNFTAmountInCollection(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")
	collectionOwnerAddr, _ := isc.AddressFromAgentID(user.AgentID())
	collectionOwner := user.Pair
	err := ctx.Chain.DepositBaseTokensToL2(ctx.Chain.Env.L1BaseTokens(collectionOwnerAddr)/2, collectionOwner)
	require.NoError(t, err)

	_, ethAddr := ctx.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ctx.Chain.ID(), ethAddr)

	collectionMetadata := isc.NewIRC27NFTMetadata(
		"text/html",
		"https://my-awesome-nft-project.com",
		"a string that is longer than 32 bytes",
	)

	collection, collectionInfo, err := ctx.Chain.Env.MintNFTL1(collectionOwner, collectionOwnerAddr, collectionMetadata.Bytes())
	require.NoError(t, err)

	nftMetadatas := []*isc.IRC27NFTMetadata{
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/1.json",
			"nft1",
		),
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/2.json",
			"nft2",
		),
	}
	nftNum := len(nftMetadatas)
	allNFTs, _, err := ctx.Chain.Env.MintNFTsL1(collectionOwner, collectionOwnerAddr, &collectionInfo.OutputID,
		lo.Map(nftMetadatas, func(item *isc.IRC27NFTMetadata, index int) []byte {
			return item.Bytes()
		}),
	)
	require.NoError(t, err)

	require.Len(t, allNFTs, nftNum+1)
	for _, nft := range allNFTs {
		require.True(t, ctx.Chain.Env.HasL1NFT(collectionOwnerAddr, &nft.ID))
	}

	// deposit all nfts on L2
	nfts := func() []*isc.NFT {
		var nfts []*isc.NFT
		for _, nft := range allNFTs {
			if nft.ID == collection.ID {
				// the collection NFT in the owner's account
				ctx.Chain.MustDepositNFT(nft, isc.NewAgentID(collectionOwnerAddr), collectionOwner)
			} else {
				// others in ethAgentID's account
				ctx.Chain.MustDepositNFT(nft, ethAgentID, collectionOwner)
				nfts = append(nfts, nft)
			}
		}
		return nfts
	}()
	require.Len(t, nfts, nftNum)

	f := coreaccounts.ScFuncs.AccountNFTAmountInCollection(ctx)
	f.Params.AgentID().SetValue(ctx.Cvt.ScAgentID(ethAgentID))
	f.Params.Collection().SetValue(ctx.Cvt.ScNftID(&collection.ID))
	f.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint32(nftNum), f.Results.Amount().Value())
}

func TestAccountNFTsInCollection(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")
	collectionOwnerAddr, _ := isc.AddressFromAgentID(user.AgentID())
	collectionOwner := user.Pair
	err := ctx.Chain.DepositBaseTokensToL2(ctx.Chain.Env.L1BaseTokens(collectionOwnerAddr)/2, collectionOwner)
	require.NoError(t, err)

	_, ethAddr := ctx.Chain.NewEthereumAccountWithL2Funds()
	ethAgentID := isc.NewEthereumAddressAgentID(ctx.Chain.ID(), ethAddr)

	collectionMetadata := isc.NewIRC27NFTMetadata(
		"text/html",
		"https://my-awesome-nft-project.com",
		"a string that is longer than 32 bytes",
	)

	collection, collectionInfo, err := ctx.Chain.Env.MintNFTL1(collectionOwner, collectionOwnerAddr, collectionMetadata.Bytes())
	require.NoError(t, err)

	nftMetadatas := []*isc.IRC27NFTMetadata{
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/1.json",
			"nft1",
		),
		isc.NewIRC27NFTMetadata(
			"application/json",
			"https://my-awesome-nft-project.com/2.json",
			"nft2",
		),
	}
	nftNum := len(nftMetadatas)
	allNFTs, _, err := ctx.Chain.Env.MintNFTsL1(collectionOwner, collectionOwnerAddr, &collectionInfo.OutputID,
		lo.Map(nftMetadatas, func(item *isc.IRC27NFTMetadata, index int) []byte {
			return item.Bytes()
		}),
	)
	require.NoError(t, err)

	require.Len(t, allNFTs, nftNum+1)
	for _, nft := range allNFTs {
		require.True(t, ctx.Chain.Env.HasL1NFT(collectionOwnerAddr, &nft.ID))
	}

	// deposit all nfts on L2
	nfts := func() []*isc.NFT {
		var nfts []*isc.NFT
		for _, nft := range allNFTs {
			if nft.ID == collection.ID {
				// the collection NFT in the owner's account
				ctx.Chain.MustDepositNFT(nft, isc.NewAgentID(collectionOwnerAddr), collectionOwner)
			} else {
				// others in ethAgentID's account
				ctx.Chain.MustDepositNFT(nft, ethAgentID, collectionOwner)
				nfts = append(nfts, nft)
			}
		}
		return nfts
	}()
	require.Len(t, nfts, nftNum)

	f := coreaccounts.ScFuncs.AccountNFTsInCollection(ctx)
	f.Params.AgentID().SetValue(ctx.Cvt.ScAgentID(ethAgentID))
	f.Params.Collection().SetValue(ctx.Cvt.ScNftID(&collection.ID))
	f.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint32(nftNum), f.Results.NftIDs().Length())
	for i := 0; i < nftNum; i++ {
		nftID := f.Results.NftIDs().GetNftID(uint32(i)).Value()
		require.True(t, ifContainNFTID(nfts, nftID))
	}
}

func ifContainNFTID(nfts []*isc.NFT, nftID wasmtypes.ScNftID) bool {
	ret := false
	stringRet := false
	byteRet := false
	for _, nft := range nfts {
		if nft.ID.String() == nftID.String() {
			stringRet = true
		}
		if bytes.Equal(nft.ID[:], nftID.Bytes()) {
			byteRet = true
		}
		ret = stringRet && byteRet
	}
	if !ret {
		fmt.Println("not exist nftID: ", nftID)
	}
	return ret
}

func TestBalance(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

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
	require.Equal(t, mintAmount, balance)

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
	require.Equal(t, mintAmount.Sub(transferTokenAmount), balance)
}

func TestBalanceBaseToken(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

	fbal := coreaccounts.ScFuncs.BalanceBaseToken(ctx)
	fbal.Params.AgentID().SetValue(user0.ScAgentID())
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user0Balance0 := fbal.Results.Balance().Value()

	fbal.Params.AgentID().SetValue(user1.ScAgentID())
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user1Balance0 := fbal.Results.Balance().Value()

	transferAmt := uint64(9)
	ftrans := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user0))
	ftrans.Params.AgentID().SetValue(user1.ScAgentID())
	transfer := wasmlib.ScTransferFromBaseTokens(transferAmt)
	ftrans.Func.Allowance(transfer).Post()
	require.NoError(t, ctx.Err)
	gasFee := ctx.GasFee

	fbal.Params.AgentID().SetValue(user0.ScAgentID())
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user0Balance1 := fbal.Results.Balance().Value()
	fbal.Params.AgentID().SetValue(user1.ScAgentID())
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user1Balance1 := fbal.Results.Balance().Value()
	require.Equal(t, user0Balance0+ctx.StorageDeposit-transferAmt-gasFee, user0Balance1)
	require.Equal(t, user1Balance0+transferAmt, user1Balance1)
}

func TestBalanceNativeToken(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

	mintAmount := wasmtypes.NewScBigInt(1000)
	foundry, err := ctx.NewSoloFoundry(mintAmount, user0)
	require.NoError(t, err)
	err = foundry.Mint(mintAmount)
	require.NoError(t, err)
	tokenID := foundry.TokenID()

	fbal := coreaccounts.ScFuncs.BalanceNativeToken(ctx)
	fbal.Params.AgentID().SetValue(user0.ScAgentID())
	fbal.Params.TokenID().SetValue(tokenID)
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user0Balance0 := fbal.Results.Tokens().Value().Uint64()

	fbal.Params.AgentID().SetValue(user1.ScAgentID())
	fbal.Params.TokenID().SetValue(tokenID)
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user1Balance0 := fbal.Results.Tokens().Value().Uint64()

	transferAmt := wasmtypes.NewScBigInt(9)
	ftrans := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.Sign(user0))
	ftrans.Params.AgentID().SetValue(user1.ScAgentID())
	transfer := wasmlib.NewScTransfer()
	transfer.Set(&tokenID, transferAmt)
	ftrans.Func.Allowance(transfer).Post()
	require.NoError(t, ctx.Err)

	fbal.Params.AgentID().SetValue(user0.ScAgentID())
	fbal.Params.TokenID().SetValue(tokenID)
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user0Balance1 := fbal.Results.Tokens().Value().Uint64()
	fbal.Params.AgentID().SetValue(user1.ScAgentID())
	fbal.Params.TokenID().SetValue(tokenID)
	fbal.Func.Call()
	require.NoError(t, ctx.Err)
	user1Balance1 := fbal.Results.Tokens().Value().Uint64()
	require.Equal(t, user0Balance0-transferAmt.Uint64(), user0Balance1)
	require.Equal(t, user1Balance0+transferAmt.Uint64(), user1Balance1)
}

func TestTotalAssets(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

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
	require.Equal(t, mintAmount0, val0)
	val1 := f.Results.Assets().GetBigInt(tokenID1).Value()
	require.Equal(t, mintAmount1, val1)
}

func TestAccounts(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

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
	allAccounts := f.Results.AllAccounts()
	require.True(t, allAccounts.GetBool(user0.ScAgentID()).Value())
	require.True(t, allAccounts.GetBool(user1.ScAgentID()).Value())
	require.False(t, allAccounts.GetBool(ctx.NewSoloAgent("dummy").ScAgentID()).Value())
}

func TestGetAccountNonce(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")

	fnon := coreaccounts.ScFuncs.GetAccountNonce(ctx)
	fnon.Params.AgentID().SetValue(user0.ScAgentID())
	fnon.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint64(0), fnon.Results.AccountNonce().Value())

	ftrans := coreaccounts.ScFuncs.TransferAllowanceTo(ctx.OffLedger(user0))
	ftrans.Params.AgentID().SetValue(user0.ScAgentID())
	ftrans.Func.TransferBaseTokens(1000).Post()
	require.NoError(t, ctx.Err)

	fnon.Func.Call()
	require.NoError(t, ctx.Err)
	require.Equal(t, uint64(1), fnon.Results.AccountNonce().Value())
}

func TestGetNativeTokenIDRegistry(t *testing.T) {
	ctx := setupAccounts(t)
	user0 := ctx.NewSoloAgent("user0")
	user1 := ctx.NewSoloAgent("user1")

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
	require.True(t, f.Results.Mapping().GetBool(tokenID0).Value())
	require.True(t, f.Results.Mapping().GetBool(tokenID1).Value())
	notExistTokenID := wasmtypes.TokenIDFromString("0x08f824508968d585ede1d154d34ba0d966ee03c928670fb85bd72e2924f67137890100000000")
	require.False(t, f.Results.Mapping().GetBool(notExistTokenID).Value())
}

func TestFoundryOutput(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")

	fnew := coreaccounts.ScFuncs.FoundryCreateNew(ctx.Sign(user))
	fnew.Params.TokenScheme().SetValue(codec.EncodeTokenScheme(&iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(1001),
		MeltedTokens:  big.NewInt(1002),
		MaximumSupply: big.NewInt(1003),
	}))
	// we need storage deposit allowance to keep foundry transaction not being trimmed by snapshot
	fnew.Func.TransferBaseTokens(sdAllowance).Post()
	require.NoError(t, ctx.Err)
	// Foundry Serial Number start from 1 and has increment 1 each func call
	serialNum := uint32(1)
	require.Equal(t, serialNum, fnew.Results.FoundrySN().Value())

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
	require.Equal(t, soloFoundry, outFoundry)
}

func TestAccountNFTs(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")
	nftID := ctx.MintNFT(user, []byte(nftMetadata))
	require.NoError(t, ctx.Err)
	userAddr, _ := isc.AddressFromAgentID(user.AgentID())

	require.True(t, ctx.Chain.Env.HasL1NFT(userAddr, ctx.Cvt.IscNFTID(&nftID)))

	fd := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	transfer := wasmlib.ScTransferFromNFT(&nftID)
	fd.Func.Transfer(transfer).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.Chain.HasL2NFT(user.AgentID(), ctx.Cvt.IscNFTID(&nftID)))

	v := coreaccounts.ScFuncs.AccountNFTs(ctx)
	v.Params.AgentID().SetValue(user.ScAgentID())
	v.Func.Call()
	require.NoError(t, ctx.Err)
	require.EqualValues(t, 1, v.Results.NftIDs().Length())
	require.EqualValues(t, nftID, v.Results.NftIDs().GetNftID(0).Value())
}

func TestNFTData(t *testing.T) {
	ctx := setupAccounts(t)
	user := ctx.NewSoloAgent("user")
	nftID := ctx.MintNFT(user, []byte(nftMetadata))
	require.NoError(t, ctx.Err)
	userAddr, _ := isc.AddressFromAgentID(user.AgentID())

	iscNFTID := ctx.Cvt.IscNFTID(&nftID)
	require.True(t, ctx.Chain.Env.HasL1NFT(userAddr, iscNFTID))

	fd := coreaccounts.ScFuncs.Deposit(ctx.Sign(user))
	transfer := wasmlib.ScTransferFromNFT(&nftID)
	fd.Func.Transfer(transfer).Post()
	require.NoError(t, ctx.Err)

	require.True(t, ctx.Chain.HasL2NFT(user.AgentID(), iscNFTID))

	v := coreaccounts.ScFuncs.NftData(ctx)
	v.Params.NftID().SetValue(nftID)
	v.Func.Call()
	require.NoError(t, ctx.Err)
	nftData, err := isc.NFTFromBytes(v.Results.NftData().Value())
	require.NoError(t, err)
	require.EqualValues(t, *iscNFTID, nftData.ID)
	require.EqualValues(t, userAddr, nftData.Issuer)
	require.EqualValues(t, []byte(nftMetadata), nftData.Metadata)
}
