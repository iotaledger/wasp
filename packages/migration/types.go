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
	AssetsBagID         *iotago.ObjectID
	GasCoinID           *iotago.ObjectID
	AnchorID            *iotago.ObjectID
	PackageID           iotago.PackageID
}

type MigrationResult struct {
	StateMetadata    *transaction.StateMetadata
	StateMetadataHex string
	StateIndex       uint32
}
