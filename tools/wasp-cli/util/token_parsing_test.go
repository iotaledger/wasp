package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/coin"
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
