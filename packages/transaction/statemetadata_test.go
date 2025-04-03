package transaction_test

import (
	"testing"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state/statetest"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/gas"
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
}
