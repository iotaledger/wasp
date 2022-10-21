package util

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const BaseTokenStr = "base"

func TokenIDFromString(s string) []byte {
	ret, err := hexutil.Decode(s)
	if err != nil {
		log.Fatalf("Invalid token id: %s", s)
	}
	return ret
}

func ParseFungibleTokens(args []string) *isc.FungibleTokens {
	tokens := isc.NewEmptyFungibleTokens()
	for _, tr := range args {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatalf("fungible tokens syntax: <token-id>:<amount> <token-id:amount>... -- Example: base:100")
		}
		// In the past we would indicate base tokens as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		tokenIDBytes := isc.BaseTokenID
		if strings.ToLower(parts[0]) != BaseTokenStr {
			tokenIDBytes = TokenIDFromString(parts[0])
		}

		amount, ok := new(big.Int).SetString(parts[1], 10)
		if !ok {
			log.Fatalf("error parsing token amount")
		}

		if isc.IsBaseToken(tokenIDBytes) {
			tokens.AddBaseTokens(amount.Uint64())
			continue
		}

		tokenID, err := isc.NativeTokenIDFromBytes(tokenIDBytes)
		log.Check(err)

		tokens.AddNativeTokens(tokenID, amount)
	}
	return tokens
}
