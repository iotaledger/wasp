package accounts_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func TestCoinRecCodec(t *testing.T) {
	bcs.TestCodec(t, &accounts.CoinRecord{
		ID:     *sui.RandomAddress(),
		Amount: 123,
	})

	bcs.TestCodecAsymmetric(t, &accounts.CoinRecord{
		ID:       *sui.RandomAddress(),
		CoinType: "IOTA",
		Amount:   123,
	})

}
