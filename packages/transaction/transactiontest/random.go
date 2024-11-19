package transactiontest

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

func RandomStateMetadata() *transaction.StateMetadata {
	return transaction.NewStateMetadata(
		allmigrations.LatestSchemaVersion,
		state.NewPseudoRandL1Commitment(),
		gas.DefaultFeePolicy(),
		isc.NewCallArguments([]byte{1, 2, 3}),
		"https://iota.org",
	)
}
