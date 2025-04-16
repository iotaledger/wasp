package migrations

import (
	"bytes"
	"fmt"
	"math"
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

func OldAgentIDtoNewAgentID(oldAgentID old_isc.AgentID, oldChainID old_isc.ChainID) isc.AgentID {
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
			// TODO: Handle cross-chain agent ID
			//panic(fmt.Sprintf("Found cross-chain agent ID: %s", oldAgentID.ChainID().AsAddress().Bech32(old_iotago.PrefixMainnet)))
		}
		hname := OldHnameToNewHname(oldAgentID.Hname())
		return isc.NewContractAgentID(hname)

	case old_isc.AgentIDKindEthereumAddress:
		oldAgentID := oldAgentID.(*old_isc.EthereumAddressAgentID)
		oldAgentChainID := oldAgentID.ChainID()
		if !oldAgentChainID.Equals(oldChainID) {
			// TODO: Handle cross-chain agent ID
			//panic(fmt.Sprintf("Found cross-chain agent ID: %s", oldAgentID))
		}
		ethAddr := oldAgentID.EthAddress()
		return isc.NewEthereumAddressAgentID(ethAddr)

	case old_isc.AgentIDIsNil:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDIsNil: %v", oldAgentID))

	case old_isc.AgentIDKindNil:
		panic(fmt.Sprintf("Found agent ID with kind = AgentIDKindNil: %v", oldAgentID))

	default:
		panic(fmt.Sprintf("Unknown agent ID kind: %v = %v", oldAgentID.Kind(), oldAgentID))
	}
}

func OldNFTIDtoNewObjectID(nftID old_iotago.NFTID) iotago.ObjectID {
	return iotago.ObjectID(nftID[:])
}

func OldNFTIDtoNewObjectRecord(nftID old_iotago.NFTID) *isc.IotaObject {
	return &isc.IotaObject{
		ID:   OldNFTIDtoNewObjectID(nftID),
		Type: OldNFTIDtoNewObjectType(nftID),
	}
}

func OldNFTIDtoNewObjectType(nftID old_iotago.NFTID) coin.Type {
	// TODO: Implement
	return lo.Must(coin.TypeFromString(fmt.Sprintf("%v::nft::NFT", nftID.ToHex())))
}

func OldNativeTokenIDtoNewCoinType(tokenID old_iotago.NativeTokenID) coin.Type {
	// TODO: Implement
	h := tokenID.ToHex()
	return lo.Must(coin.TypeFromString(fmt.Sprintf("%v::nt::NT%v", h[:66], h[66:])))
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
	if !oldNativeTokenAmount.IsUint64() {
		fmt.Println(fmt.Errorf("\n** ERROR old native token amount cannot be represented as uint64: balance = %v", oldNativeTokenAmount))
		return coin.Value(math.MaxUint64)
	}

	return coin.Value(oldNativeTokenAmount.Uint64())
}

func ConvertOldCoinDecimalsToNew(from uint64) coin.Value {
	return coin.Value(from * 1000) // Stardust 6 / Rebased 9 decimals
}

func OldAssetsToNewAssets(oldAssets *old_isc.Assets) *isc.Assets {
	newBaseTokensBalance := ConvertOldCoinDecimalsToNew(oldAssets.BaseTokens)
	newAssets := isc.NewAssets(newBaseTokensBalance)

	for _, oldToken := range oldAssets.NativeTokens {
		newCoinType := OldNativeTokenIDtoNewCoinType(oldToken.ID)
		newBalance := OldNativeTokenBalanceToNewCoinValue(oldToken.Amount)
		newAssets.Coins.Add(newCoinType, newBalance)
	}

	for _, nftID := range oldAssets.NFTs {
		nftObjID := OldNFTIDtoNewObjectID(nftID)
		nftObjType := OldNFTIDtoNewObjectType(nftID)
		newAssets.Objects.Add(isc.IotaObject{
			ID:   nftObjID,
			Type: nftObjType,
		})
	}

	return newAssets
}

func IsValidOldAccountKeyBytesLen(n int) bool {
	return n == isc.HnameLength || n == common.AddressLength || n == iotago.AddressLen
}
