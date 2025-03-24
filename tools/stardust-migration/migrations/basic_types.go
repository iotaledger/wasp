package migrations

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	"github.com/samber/lo"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

var newSchema isc.SchemaVersion = allmigrations.SchemaVersionMigratedRebased

func OldHnameToNewHname(oldHname old_isc.Hname) isc.Hname {
	return isc.Hname(oldHname)
}

func OldAgentIDToCryptoLibAddress(oldAgentID old_isc.AgentID) *cryptolib.Address {
	if oldAgentID == nil {
		return nil
	}

	switch oldAgentID.Kind() {
	case old_isc.AgentIDKindAddress:
		oldAddr := oldAgentID.(*old_isc.AddressAgentID).Address()
		return lo.Must(cryptolib.NewAddressFromHexString(oldAddr.String()))
	default:
		panic("OldAgentIDToCryptoLibAddress: Unsupported address type!")
	}
}

func OldIotaGoAddressToCryptoLibAddress(address old_iotago.Address) *cryptolib.Address {
	return lo.Must(cryptolib.NewAddressFromHexString(address.String()))
}

func OldAgentIDtoNewAgentID(oldAgentID old_isc.AgentID, oldChainID old_isc.ChainID, newChainID isc.ChainID) isc.AgentID {
	switch oldAgentID.Kind() {
	case old_isc.AgentIDKindAddress:
		// TODO: I think we need to remove the first byte from the byte array
		// https://docs.iota.org/developer/stardust/addresses#bech32-to-hex-conversion
		// There you can see Bech32->Bytes->Remove first byte-> == new address
		oldAddr := oldAgentID.(*old_isc.AddressAgentID).Address()
		newAdd := iotago.MustAddressFromHex(oldAddr.String())
		return isc.NewAddressAgentID(cryptolib.NewAddressFromIota(newAdd))

	case old_isc.AgentIDKindContract:
		oldAgentID := oldAgentID.(*old_isc.ContractAgentID)
		oldAgentChainID := oldAgentID.ChainID()

		if !bytes.Equal(oldChainID.Bytes(), oldAgentChainID.Bytes()) {
			//panic(fmt.Sprintf("Found cross-chain agent ID: %s", oldAgentID.ChainID().AsAddress().Bech32(old_iotago.PrefixMainnet)))
		}
		hname := OldHnameToNewHname(oldAgentID.Hname())
		return isc.NewContractAgentID(newChainID, hname)

	case old_isc.AgentIDKindEthereumAddress:
		oldAgentID := oldAgentID.(*old_isc.EthereumAddressAgentID)
		oldAgentChainID := oldAgentID.ChainID()
		if !oldAgentChainID.Equals(oldChainID) {
			//panic(fmt.Sprintf("Found cross-chain agent ID: %s", oldAgentID))
		}
		ethAddr := oldAgentID.EthAddress()
		return isc.NewEthereumAddressAgentID(newChainID, ethAddr)

	case old_isc.AgentIDIsNil:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", oldAgentID))

	case old_isc.AgentIDKindNil:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDKindNil: %v", oldAgentID))

	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v = %v", oldAgentID.Kind(), oldAgentID))
	}
}

func OldNFTIDtoNewObjectID(nftID old_iotago.NFTID) iotago.ObjectID {
	//panic("TODO: Not implemented")
	return iotago.ObjectID{}
}

func OldNFTIDtoNewObjectRecord(nftID old_iotago.NFTID) *accounts.ObjectRecord {
	//panic("TODO: Not implemented")
	return &accounts.ObjectRecord{
		ID:  iotago.ObjectID{},
		BCS: []byte{1, 2, 3, 4, 5},
	}
}

// // Creates converter from old account key to new account key.
// func OldAccountKeyToNewAccountKey(oldChainID old_isc.ChainID, newChainID isc.ChainID) func(oldAccountKey old_kv.Key) (newAccKey kv.Key) {
// 	return func(oldAccountKey old_kv.Key) (newAccKey kv.Key) {
// 		oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChainID))
// 		newAgentID := OldAgentIDtoNewAgentID(oldAgentID, newChainID)
// 		return accounts.AccountKey(newAgentID, newChainID)
// 	}
// }

// func OldBaseTokensFullDecimalBalanceToNewBalance(oldTokensBalance *big.Int) (balance coin.Value, remainder *big.Int) {
// 	// TODO: what is the conversion rate?
// 	var newBalance uint64
// 	newBalance, remainder = old_util.EthereumDecimalsToBaseTokenDecimals(oldTokensBalance, oldBaseTokenDecimals)
// 	return coin.Value(newBalance), remainder
// }

// Base token balance storage format:
//   New IOTA rebased (uint64 IOTA + big.Int Wei remainder):
//     Acc1 = 4 IOTA
//     Rem1 = 0.5 IOTA
//     Acc2 = 5 IOTA
//     Rem2 = 0.5 IOTA
//     Total = 10 IOTA
//   Old IOTA with new schema (big.Int):
//     Acc1 = 4.5 IOTA
//     Acc2 = 5.5 IOTA
//     Total = 10 IOTA
//   Old IOTA with old schema (uint64 + added zeros in evm)
//     Acc1 = 4.000 IOTA
//     Acc2 = 5.000 IOTA
//     Total = 9 IOTA

// const (
// 	// TODO: what is the correct value?
// 	oldBaseTokenDecimals uint32 = 6
// )

// func DecodeOldTokens(b []byte) uint64 {
// 	amount := old_codec.MustDecodeBigIntAbs(b, big.NewInt(0))
// 	// TODO: This is incorrect for native tokens and for base tokens of old schema
// 	convertedAmount, remainder := old_util.EthereumDecimalsToBaseTokenDecimals(amount, oldBaseTokenDecimals)
// 	_ = remainder

// 	return convertedAmount
// }

func OldNativeTokenIDtoNewCoinType(tokenID old_iotago.NativeTokenID) coin.Type {
	// TODO: Implement
	// Temporary implementaion needed to fix base tokens migration, becase right now native token balances override base token balances
	return lo.Must(coin.TypeFromString(fmt.Sprintf("%v::nt::NT", tokenID.ToHex()[:64])))
}

func OldNativeTokenIDtoNewCoinInfo(tokenID old_iotago.NativeTokenID) parameters.IotaCoinInfo {
	//panic("TODO: Not implemented")
	return parameters.IotaCoinInfo{
		CoinType:    coin.BaseTokenType,
		Decimals:    6,
		Name:        "DUMMY",
		Symbol:      "DUMMY",
		Description: "DUMMY",
		IconURL:     "DUMMY",
		TotalSupply: 123456,
	}
}

func OldNativeTokenBalanceToNewCoinValue(oldNativeTokenAmount *big.Int) coin.Value {
	// TODO: There is no cinversion rate, right?

	if !oldNativeTokenAmount.IsUint64() {
		fmt.Println(fmt.Errorf("old native token amount cannot be represented as uint64: balance = %v", oldNativeTokenAmount))
	}

	u := uint64(18446744073709551615)
	return coin.Value(u)
}

func convertBaseTokens(oldBalanceFullDecimals *big.Int) *big.Int {
	//panic("TODO: do we need to apply a conversion rate because of iota's 6 to 9 decimals change?")
	return big.NewInt(0).Mul(oldBalanceFullDecimals, big.NewInt(1_000))
}

func ConvertOldCoinDecimalsToNew(from uint64) coin.Value {
	return coin.Value(from * 1000) // Stardust 6 / Rebased 9 decimals
}

func OldAssetsToNewAssets(oldAssets *old_isc.Assets) *isc.Assets {
	// TODO: conversion rate?
	newBaseTokensBalance := ConvertOldCoinDecimalsToNew(oldAssets.BaseTokens)
	newAssets := isc.NewAssets(newBaseTokensBalance)

	for _, oldToken := range oldAssets.NativeTokens {
		newCoinType := OldNativeTokenIDtoNewCoinType(oldToken.ID)
		newBalance := OldNativeTokenBalanceToNewCoinValue(oldToken.Amount)
		newAssets.Coins.Add(newCoinType, newBalance)
	}

	for _, nftID := range oldAssets.NFTs {
		nftObjID := OldNFTIDtoNewObjectID(nftID)
		newAssets.Objects.Add(nftObjID)
	}

	return newAssets
}

func IsValidOldAccountKeyBytesLen(n int) bool {
	if n > old_isc.ChainIDLength {
		return true
	}

	return n == 4 || n == common.AddressLength
}
