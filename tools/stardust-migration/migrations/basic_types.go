package migrations

import (
	"bytes"
	"fmt"
	"math"
	"math/big"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_gas "github.com/nnikolash/wasp-types-exported/packages/vm/gas"
	"github.com/samber/lo"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"
	new_util "github.com/iotaledger/wasp/packages/util"

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
		oldAddr := oldAgentID.(*old_isc.AddressAgentID).Address()
		newAdd := iotago.MustAddressFromHex(oldAddr.String())
		return isc.NewAddressAgentID(cryptolib.NewAddressFromIota(newAdd))

	case old_isc.AgentIDKindContract:
		oldAgentID := oldAgentID.(*old_isc.ContractAgentID)
		oldAgentChainID := oldAgentID.ChainID()

		hname := OldHnameToNewHname(oldAgentID.Hname())
		if !bytes.Equal(oldChainID.Bytes(), oldAgentChainID.Bytes()) {
			//NOTE: We cannot migrate cross-chain agent ID, so we "shift" it to not overlap with local chain ID
			hname++
		}

		return isc.NewContractAgentID(hname)

	case old_isc.AgentIDKindEthereumAddress:
		oldAgentID := oldAgentID.(*old_isc.EthereumAddressAgentID)
		oldAgentChainID := oldAgentID.ChainID()
		ethAddr := oldAgentID.EthAddress()
		if !oldAgentChainID.Equals(oldChainID) {
			//NOTE: We cannot migrate cross-chain agent ID, so we "shift" it to not overlap with local chain ID
			ethAddr[0]++
		}

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
		Type: OldNFTIDtoNewObjectType(),
	}
}

func OldNFTIDtoNewObjectType() coin.Type {
	return lo.Must(coin.TypeFromString("0x000000000000000000000000000000000000000000000000000000000000107a::nft::Nft"))
}

func OldNativeTokenIDtoNewCoinType(tokenID old_iotago.NativeTokenID) coin.Type {
	// TODO: Implement
	h := tokenID.ToHex()
	return lo.Must(coin.TypeFromString(fmt.Sprintf("%v::nt::NT%v", h[:66], h[66:])))
}

func OldGasPerTokenToNew(oldFeePolicy *old_gas.FeePolicy) new_util.Ratio32 {
	newGasPerToken := new_util.Ratio32{}

	if oldFeePolicy.GasPerToken.A == 1 && oldFeePolicy.GasPerToken.B == 1 {
		newGasPerToken = new_util.Ratio32{A: 1, B: 1000}
	} else if oldFeePolicy.GasPerToken.A == 100 && oldFeePolicy.GasPerToken.B == 1 {
		newGasPerToken = new_util.Ratio32{A: 1, B: 10}
	} else {
		panic(fmt.Sprintf("Unknown gas per token: %v", oldFeePolicy.GasPerToken))
	}

	return newGasPerToken
}

// Taken from here: https://explorer.iota.org/mainnet/addr/iota1pzt3mstq6khgc3tl0mwuzk3eqddkryqnpdxmk4nr25re2466uxwm28qqxu5?tab=NativeTokens
var knownCoinInfos = map[string]parameters.IotaCoinInfo{
	"0x08b4db42a2f13195b19c0f8b16e1928b297f677d7678f3121acb6524f8e8f8b21e0100000000": {
		Name:        "TUDITI",
		Symbol:      "TUD",
		Decimals:    1,
		TotalSupply: 1234,
	},
	"0x08971dc160d5ae8c457f7eddc15a39035b6190130b4dbb5663550795575ae19db50100000000": {
		Name:        "Nicole Coin",
		Symbol:      "NICO",
		Decimals:    18,
		TotalSupply: math.MaxUint64 - 1,
	},
	"0x08c7c5eac5ac0ffb8dc0f1f555d8150e45f220547a7ec02c43d3dc43c95a2888580200000000": {
		Name:        "SDV",
		Symbol:      "SDV",
		Decimals:    18,
		Description: "SDV token",
		IconURL:     "https://sdvconsulting.it/wp-content/uploads/2023/10/Logo-chiaro-quadrato-1.jpg",
		TotalSupply: 100000,
	},
	"0x08fc7819ec252ade5d3e51beee9b28a156073d9c4a1d30ba2e3f6720036abdb6430100000000": {
		Name:        "House Token",
		Symbol:      "HOUSE",
		Decimals:    6,
		Description: "The House token represents everyone that is building and supporting the IOTA eco system.We have built the first quality NFT houses on soonaverse (IOTA).\n\nBenefits holding  $house tokens:\n\n🏠Metaverse game\n🏠Staking your iotahouse nft's to earn passive income\n🏠Buying real estate in the future with your $HOUSE tokens. \n\n\n\"The global real estate market size was valued at USD 3.69 trillion in 2021 and is expected to expand at a compound annual growth rate (CAGR) of 5.2% from 2022 to 2030.\" You will be able to buy with crypto  real estate and is happening already. We want to be the first feeless  leader where you can buy with your $House tokens  real estate.\n\n\n\n",
		IconURL:     "ipfs://bafkreif5leoegqkdu4564fpg3vm6dhfwhfl4ohx2qhso6y7wdi754rcrhy",
		TotalSupply: 1000000000000000,
	},
	"0x083c55be0f034673cef16a7553f42a7928d998ccc1e970968ea0965608de2c6a440100000000": {
		Name:        "Raguguys",
		Symbol:      "RDG",
		Decimals:    0,
		IconURL:     "https://gurutom.github.io/raguguys/logo512.png",
		TotalSupply: 10000000000,
	},
	"0x084fbbb31eb49ae1c2416fdd97d675d94b280cbd88ebf0c13c93a4f6f01cba5fdf0100000000": {
		Name:        "Ape token",
		Symbol:      "APE",
		Decimals:    6,
		Description: "The $Ape token is created to  give everyone passive income by staking it or our NFT collection\nBuying our $Ape token will have the following benefits:\n✅$Ape tokens (will  give you a free upgraded NFT to play in the metaverse/game) \n✅Unlocking the metaverse/game\n✅Play to earn game. Earning $ape tokens by  playing game competition with a score board.\n✅Being an OG hodler on our telegram/discord\n✅Representing womenDAO (i.e. educating more women to the crypto and NFT market , since we've done a research that crypto/NFT world  is male dominated. \n\nTo make sure we get the best trust from the community and being fair,we recommend everyone to check our up-to-date roadmap on our website or our Twitter page.\n\nWe hope we can get the best support from our lovely $ape community and everyone that likes our vision. We've a strong team and dedication to bring this project beyond the moon.\n\nAnd like always, it's not always about the product , but how much utility  you'll bring in the long-term to your community.",
		IconURL:     "ipfs://bafkreihtdy5o5ck6owan6iw26dbootexfloih6ecs3oydxeteiabfgtdju",
		TotalSupply: 100000000000000,
	},
}

func OldNativeTokenIDtoNewCoinInfo(tokenID old_iotago.NativeTokenID) parameters.IotaCoinInfo {
	coinInfo, known := knownCoinInfos[tokenID.ToHex()]
	if !known {
		//panic(fmt.Sprintf("Unknown token ID: %v", tokenID.ToHex()))
		return parameters.IotaCoinInfo{
			CoinType: coin.BaseTokenType,
		}
	}

	if coinInfo.CoinType.String() == "" {
		coinInfo.CoinType = OldNativeTokenIDtoNewCoinType(tokenID)
	}

	return coinInfo
}

func OldNativeTokenBalanceToNewCoinValue(oldNativeTokenAmount *big.Int) coin.Value {
	if !oldNativeTokenAmount.IsUint64() {
		fmt.Println(fmt.Errorf("\n** ERROR old native token amount cannot be represented as uint64: balance = %v", oldNativeTokenAmount))
		return coin.Value(math.MaxUint64 - 1)
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
		nftObjType := OldNFTIDtoNewObjectType()
		newAssets.Objects.Add(isc.IotaObject{
			ID:   nftObjID,
			Type: nftObjType,
		})
	}

	return newAssets
}
