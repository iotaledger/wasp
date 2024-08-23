package transaction

import (
	"io"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// StateMetadata contains the information stored in the Anchor object on L1
type StateMetadata struct {
	SchemaVersion isc.SchemaVersion
	L1Commitment  *state.L1Commitment
	GasFeePolicy  *gas.FeePolicy
	InitParams    isc.CallArguments
	PublicURL     string
}

func NewStateMetadata(
	schemaVersion isc.SchemaVersion,
	l1Commitment *state.L1Commitment,
	gasFeePolicy *gas.FeePolicy,
	initParams isc.CallArguments,
	publicURL string,
) *StateMetadata {
	return &StateMetadata{
		SchemaVersion: schemaVersion,
		L1Commitment:  l1Commitment,
		GasFeePolicy:  gasFeePolicy,
		PublicURL:     publicURL,
		InitParams:    initParams,
	}
}

func StateMetadataFromBytes(data []byte) (*StateMetadata, error) {
	return rwutil.ReadFromBytes(data, new(StateMetadata))
}

func (s *StateMetadata) Bytes() []byte {
	return rwutil.WriteToBytes(s)
}

func (s *StateMetadata) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	s.SchemaVersion = isc.SchemaVersion(rr.ReadUint32())
	s.L1Commitment = new(state.L1Commitment)
	rr.Read(s.L1Commitment)
	s.GasFeePolicy = new(gas.FeePolicy)
	rr.Read(s.GasFeePolicy)
	s.InitParams = make(isc.CallArguments, 0)
	rr.Read(&s.InitParams)
	s.PublicURL = rr.ReadString()
	return rr.Err
}

func (s *StateMetadata) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteUint32(uint32(s.SchemaVersion))
	ww.Write(s.L1Commitment)
	ww.Write(s.GasFeePolicy)
	ww.Write(s.InitParams)
	ww.WriteString(s.PublicURL)
	return ww.Err
}

func L1CommitmentFromAnchor(anchor *iscmove.Anchor) (*state.L1Commitment, error) {
	stateMetadata, err := StateMetadataFromBytes(anchor.StateMetadata)
	if err != nil {
		return nil, err
	}
	return stateMetadata.L1Commitment, nil
}
