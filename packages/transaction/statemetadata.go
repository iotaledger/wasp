package transaction

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/state"
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
	Version        byte
	L1Commitment   *state.L1Commitment
	GasFeePolicy   *gas.FeePolicy
	SchemaVersion  uint32
	CustomMetadata []byte
}

func NewStateMetadata(
	l1Commitment *state.L1Commitment,
	gasFeePolicy *gas.FeePolicy,
	schemaVersion uint32,
	customMetadata []byte,
) *StateMetadata {
	return &StateMetadata{
		Version:        StateMetadataSupportedVersion,
		L1Commitment:   l1Commitment,
		GasFeePolicy:   gasFeePolicy,
		SchemaVersion:  schemaVersion,
		CustomMetadata: customMetadata,
	}
}

func (s *StateMetadata) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(StateMetadataSupportedVersion) // Always write the new version.
	mu.WriteUint32(s.SchemaVersion)
	mu.WriteBytes(s.L1Commitment.Bytes())
	mu.WriteBytes(s.GasFeePolicy.Bytes())
	mu.WriteUint16(uint16(len(s.CustomMetadata)))
	mu.WriteBytes(s.CustomMetadata)
	return mu.Bytes()
}

func StateMetadataFromBytes(data []byte) (*StateMetadata, error) {
	ret := &StateMetadata{}
	mu := marshalutil.New(data)
	var err error

	ret.Version, err = mu.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unable to parse state metadata version, error: %w", err)
	}
	if ret.Version > StateMetadataSupportedVersion {
		return nil, fmt.Errorf("unsupported state metadata version: %d", ret.Version)
	}

	ret.SchemaVersion, err = mu.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("unable to parse schema version, error: %w", err)
	}

	l1CommitmentBytes, err := mu.ReadBytes(state.L1CommitmentSize)
	if err != nil {
		return nil, fmt.Errorf("unable to parse l1 commitment, error: %w", err)
	}

	ret.L1Commitment, err = state.L1CommitmentFromBytes(l1CommitmentBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse l1 commitment, error: %w", err)
	}

	ret.GasFeePolicy, err = gas.FeePolicyFromMarshalUtil(mu)
	if err != nil {
		return nil, fmt.Errorf("unable to parse gas fee policy, error: %w", err)
	}

	customMetadataLength, err := mu.ReadUint16()
	if err != nil {
		return nil, fmt.Errorf("unable to parse custom metadata length, error: %w", err)
	}

	customMetadataBytes, err := mu.ReadBytes(int(customMetadataLength))
	if err != nil {
		return nil, fmt.Errorf("unable to parse custom metadata, error: %w", err)
	}
	ret.CustomMetadata = customMetadataBytes

	return ret, nil
}

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
