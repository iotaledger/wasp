package coin_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestCoinCodec(t *testing.T) {
	bcs.TestCodec(t, coin.Value(123))
	bcs.TestCodec(t, coin.Type("0xa1::a::A"))
}
