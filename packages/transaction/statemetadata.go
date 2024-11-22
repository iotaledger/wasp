package transaction

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// StateMetadata contains the information stored in the Anchor object on L1
type StateMetadata struct {
	SchemaVersion   isc.SchemaVersion
	L1Commitment    *state.L1Commitment
	GasCoinObjectID iotago.ObjectID
	GasFeePolicy    *gas.FeePolicy
	InitParams      isc.CallArguments
	PublicURL       string
}

func NewStateMetadata(
	schemaVersion isc.SchemaVersion,
	l1Commitment *state.L1Commitment,
	gasCoinObjectID iotago.ObjectID,
	gasFeePolicy *gas.FeePolicy,
	initParams isc.CallArguments,
	publicURL string,
) *StateMetadata {
	return &StateMetadata{
		SchemaVersion:   schemaVersion,
		L1Commitment:    l1Commitment,
		GasCoinObjectID: gasCoinObjectID,
		GasFeePolicy:    gasFeePolicy,
		PublicURL:       publicURL,
		InitParams:      initParams,
	}
}

func StateMetadataFromBytes(data []byte) (*StateMetadata, error) {
	return bcs.Unmarshal[*StateMetadata](data)
}

func (s *StateMetadata) Bytes() []byte {
	return bcs.MustMarshal(s)
}

func L1CommitmentFromAnchor(anchor *isc.StateAnchor) (*state.L1Commitment, error) {
	stateMetadata, err := StateMetadataFromBytes(anchor.GetStateMetadata())
	if err != nil {
		return nil, err
	}
	return stateMetadata.L1Commitment, nil
}
