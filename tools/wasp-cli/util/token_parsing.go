package util

import (
	"math/big"
	"strings"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const BaseTokenStr = "base"

func TokenIDFromString(s string) []byte {
	ret, err := iotago.DecodeHex(s)
	if err != nil {
		log.Fatalf("Invalid token id: %s", s)
	}
	return ret
}

func ParseNFTIDs(args []string) iotago.NFTIDs {
	tokens := iotago.NFTIDs{}

	for _, tr := range args {
		nftId, err := iotago.ParseNFTAddressFromHexString(tr)

		if err != nil {
			continue
		}

		tokens = append(tokens, nftId.NFTID())
	}

	return tokens
}

func ParseAssetArgs(args []string) *isc.Assets {
	tokens := isc.NewEmptyAssets()
	for _, tr := range args {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatal("assets syntax: <token-id>:<amount>, [<token-id>:amount ...], " +
				"<nft>:nftID, [<nft>:nftID ...] -- Example: base:100, " +
				"nft:0xb7d0b62cafbae6c2c60215acce712b4d714c44fca211f65b9f6e51818cb59dcc")
		}

		leftSide := strings.ToLower(parts[0])

		// First check if the supplied argument is of id "nft"
		if leftSide == "nft" {
			rightSide := strings.ToLower(parts[1])
			nft, err := iotago.ParseNFTAddressFromHexString(strings.TrimSpace(rightSide))

			log.Check(err)
			tokens.AddNFTs(nft.NFTID())
			continue
		}

		// Otherwise handle native tokens

		// In the past we would indicate base tokens as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		tokenIDBytes := isc.BaseTokenID

		if leftSide != BaseTokenStr {
			tokenIDBytes = TokenIDFromString(strings.TrimSpace(parts[0]))
		}

		amount, ok := new(big.Int).SetString(parts[1], 10)
		if !ok {
			log.Fatal("error parsing token amount")
		}

		if isc.IsBaseToken(tokenIDBytes) {
			tokens.AddBaseTokens(amount.Uint64())
			continue
		}

		nativeTokenID, err := isc.NativeTokenIDFromBytes(tokenIDBytes)
		log.Check(err)

		tokens.AddNativeTokens(nativeTokenID, amount)
	}
	return tokens
}
