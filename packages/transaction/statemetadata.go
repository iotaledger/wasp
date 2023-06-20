package transaction

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	// L1Commitment calculation has changed from version 0 to version 1.
	// The structure is actually the same, but the L1 commitment in V0
	// refers to an empty state, and in V1 refers to the first initialized
	// state.
	StateMetadataSupportedVersion byte = 1
)

type StateMetadata struct {
	Version       byte
	L1Commitment  *state.L1Commitment
	GasFeePolicy  *gas.FeePolicy
	SchemaVersion uint32
	PublicURL     string
}

func NewStateMetadata(
	l1Commitment *state.L1Commitment,
	gasFeePolicy *gas.FeePolicy,
	schemaVersion uint32,
	publicURL string,
) *StateMetadata {
	return &StateMetadata{
		Version:       StateMetadataSupportedVersion,
		L1Commitment:  l1Commitment,
		GasFeePolicy:  gasFeePolicy,
		SchemaVersion: schemaVersion,
		PublicURL:     publicURL,
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
	s.Version = rr.ReadByte()
	if s.Version > StateMetadataSupportedVersion && rr.Err == nil {
		return fmt.Errorf("unsupported state metadata version: %d", s.Version)
	}
	s.SchemaVersion = rr.ReadUint32()
	s.L1Commitment = new(state.L1Commitment)
	rr.Read(s.L1Commitment)
	s.GasFeePolicy = new(gas.FeePolicy)
	rr.Read(s.GasFeePolicy)
	s.PublicURL = rr.ReadString()
	return rr.Err
}

func (s *StateMetadata) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteByte(StateMetadataSupportedVersion)
	ww.WriteUint32(s.SchemaVersion)
	ww.Write(s.L1Commitment)
	ww.Write(s.GasFeePolicy)
	ww.WriteString(s.PublicURL)
	return ww.Err
}

/////////////// avoiding circular imports: state <-> transaction //////////////////

func L1CommitmentFromAliasOutput(ao *iotago.AliasOutput) (*state.L1Commitment, error) {
	s, err := StateMetadataFromBytes(ao.StateMetadata)
	if err != nil {
		return nil, err
	}
	return s.L1Commitment, nil
}

func MustL1CommitmentFromAliasOutput(ao *iotago.AliasOutput) *state.L1Commitment {
	l1c, err := L1CommitmentFromAliasOutput(ao)
	if err != nil {
		panic(err)
	}
	return l1c
}
