package util

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/iotaledger/iota.go/v3/bech32"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func TestTokenParsing(t *testing.T) {
	fakeCoin, err := coin.TypeFromString("0x3::threeota::threeota")
	require.NoError(t, err)

	coinStr := []string{
		"base|1000",
		fmt.Sprintf("%v|1000", coin.BaseTokenType),
		fmt.Sprintf("%v|1074", fakeCoin),
	}

	assets := ParseFungibleTokens(coinStr)

	require.Equal(t, assets.BaseTokens().Uint64(), uint64(2000))
	require.Equal(t, assets.CoinBalance(fakeCoin).Uint64(), uint64(1074))
}

func TestSeed(t *testing.T) {
	see := "iotaprivkey1qranfhgmyxs5qazfse4xx8p6l4c5m974j6nfcfl690autxedgalh6w4y74r"
	k, b, e := bech32.Decode(see)
	log.Check(e)
	fmt.Printf("%v %s", k, hexutil.Encode(b))
}
