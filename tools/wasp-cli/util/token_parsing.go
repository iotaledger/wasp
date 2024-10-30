package util

import (
	"strings"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const BaseTokenStr = "base"

func TokenIDFromString(s string) []byte {
	ret, err := cryptolib.DecodeHex(s)
	if err != nil {
		log.Fatalf("Invalid token id: %s", s)
	}
	return ret
}

func ArgsToFungibleTokensStr(args []string) []string {
	return strings.Split(strings.Join(args, ""), ",")
}

func ParseFungibleTokens(args []string) *isc.Assets {

	tokens := isc.NewEmptyAssets()
	for _, tr := range args {
		parts := strings.Split(tr, "|")
		if len(parts) != 2 {
			log.Fatal("fungible tokens syntax: <token-id>|<amount>, <token-id|amount>... -- Example: base|100")
		}
		// In the past we would indicate base tokens as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		if strings.ToLower(parts[0]) != BaseTokenStr {
			log.Fatalf("invalid token id: %s", parts[0])
		}

		amount, err := coin.ValueFromString(parts[1])
		if err != nil {
			log.Fatalf("error parsing token amount: %v", err)
		}

		tokens.AddBaseTokens(amount)

		/*
			nativeTokenID, err := isc.NativeTokenIDFromBytes(tokenIDBytes)
			log.Check(err)

			tokens.AddNativeTokens(nativeTokenID, amount)*/
	}
	return tokens
}
