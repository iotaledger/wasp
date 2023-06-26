package transaction_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func TestStateMetadataSerialization(t *testing.T) {
	s := transaction.NewStateMetadata(
		state.PseudoRandL1Commitment(),
		&gas.FeePolicy{
			GasPerToken: util.Ratio32{
				A: 1,
				B: 2,
			},
			EVMGasRatio: util.Ratio32{
				A: 3,
				B: 4,
			},
			ValidatorFeeShare: 5,
		},
		6,
		"https://iota.org",
	)
	rwutil.BytesTest(t, s, transaction.StateMetadataFromBytes)
}
