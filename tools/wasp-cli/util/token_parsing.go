package util

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

const BaseTokenStr = "base"

func TokenIDFromString(s string) []byte {
	ret, err := hex.DecodeString(s)
	log.Check(err)
	return ret
}

func ParseFungibleTokens(args []string) *iscp.FungibleTokens {
	tokens := iscp.NewEmptyAssets()
	for _, tr := range args {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatalf("fungible tokens syntax: <token-id>:<amount> <token-id:amount>... -- Example: base:100")
		}
		// In the past we would indicate base tokens as 'IOTA:nnn'
		// Now we can simply use ':nnn', but let's keep it
		// backward compatible for now and allow both
		if strings.ToLower(parts[0]) == BaseTokenStr {
			parts[0] = ""
		}
		tokenIDBytes := TokenIDFromString(parts[0])

		amount, ok := new(big.Int).SetString(parts[1], 10)
		if !ok {
			log.Fatalf("error parsing token amount")
		}

		if iscp.IsBaseToken(tokenIDBytes) {
			tokens.AddBaseTokens(amount.Uint64())
			continue
		}

		tokenID, err := iscp.NativeTokenIDFromBytes(tokenIDBytes)
		log.Check(err)

		tokens.AddNativeTokens(tokenID, amount)
	}
	return tokens
}
