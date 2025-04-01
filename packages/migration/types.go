package migration

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/transaction"
)

// This package only exist to pass shared types from the wasp-cli to the statedb-migrator tool

type PrepareConfiguration struct {
	DKGCommitteeAddress *cryptolib.Address
	ChainOwner          *cryptolib.Address
	AnchorID            *iotago.ObjectID
	PackageID           iotago.PackageID
	L1ApiUrl            string
}

type MigrationResult struct {
	StateMetadata    *transaction.StateMetadata
	StateMetadataHex string
	StateIndex       uint32
}
