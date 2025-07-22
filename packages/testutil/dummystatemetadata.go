package testutil

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

func DummyStateMetadata(commitment *state.L1Commitment) *transaction.StateMetadata {
	return transaction.NewStateMetadata(
		0,
		commitment,
		&iotago.ObjectID{},
		gas.DefaultFeePolicy(),
		[][]byte{},
		0,
		"",
	)
}
