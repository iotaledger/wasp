package migration

import (
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/transaction"
)

// This package only exist to pass shared types from the wasp-cli to the statedb-migrator tool

type PrepareConfiguration struct {
	DKGCommitteeAddress *cryptolib.Address
	ChainOwner          *cryptolib.Address
	StateMetadata       *transaction.StateMetadata
	Anchor              *iscmove.AnchorWithRef
}
