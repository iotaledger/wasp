package transaction

import (
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

const (
	StateMetadataSupportedVersion = 0
)

type StateMetadata struct {
	L1Commitment   *state.L1Commitment
	GasFeePolicy   *gas.FeePolicy
	SchemaVersion  uint32
	CustomMetadata []byte
}

func (s *StateMetadata) Bytes() []byte {
	mu := marshalutil.New()
	mu.WriteByte(StateMetadataSupportedVersion)
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

	version, err := mu.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("unable to parse state metadata version, error: %w", err)
	}
	if version != StateMetadataSupportedVersion {
		return nil, fmt.Errorf("unsupported state metadata version: %d", version)
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
	if len(ao.StateMetadata) == state.L1CommitmentSize {
		return state.L1CommitmentFromBytes(ao.StateMetadata)
	}
	s, err := StateMetadataFromBytes(ao.StateMetadata)
	if err != nil {
		return nil, err
	}
	return s.L1Commitment, nil
}
