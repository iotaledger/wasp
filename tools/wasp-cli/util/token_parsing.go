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
	// TODO: "send-funds" and "chain deposit" pass tokens in different way.
	// When there are multiple tokens, one command separates with comma, one doesn't
	return strings.Split(strings.Join(args, ""), ",")
}

func ParseFungibleTokens(args []string) *isc.Assets {
	tokens := isc.NewEmptyAssets()

	for _, tr := range args {
		parts := strings.Split(tr, "|")
		if len(parts) != 2 {
			log.Fatal("fungible tokens syntax: <token-id1>|<amount1>, <token-id2>|<amount2>... -- Example: base|100")
		}

		amount, err := coin.ValueFromString(parts[1])
		if err != nil {
			log.Fatalf("error parsing token amount: %v", err)
		}

		// In the past we would indicate base tokens as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		if strings.ToLower(parts[0]) == BaseTokenStr || coin.BaseTokenType.MatchesStringType(parts[0]) {
			tokens.AddBaseTokens(amount)
		} else {
			coinID, err := coin.TypeFromString(parts[0])
			log.Check(err)
			tokens.AddCoin(coinID, amount)
		}
	}

	return tokens
}
