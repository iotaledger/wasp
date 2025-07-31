package transaction_test

import (
	"testing"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/util"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func TestStateMetadataSerialization(t *testing.T) {
	s := transaction.NewStateMetadata(
		1,
		statetest.NewRandL1Commitment(),
		&iotago.ObjectID{},
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
		isc.NewCallArguments([]byte{1, 2, 3}),
		0,
		"https://iota.org",
	)
	bcs.TestCodec(t, s)

	s.L1Commitment = statetest.TestL1Commitment
	bcs.TestCodecAndHash(t, s, "0dd16b4478ba")
}
