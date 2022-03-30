package state

import (
	"bytes"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/trie"
)

// L1Commitment represents the parsed data stored as a metadata in the anchor output
type L1Commitment struct {
	Commitment trie.VCommitment
}

func NewL1Commitment(c trie.VCommitment) *L1Commitment {
	return &L1Commitment{Commitment: c}
}

func L1CommitmentFromBytes(data []byte) (L1Commitment, error) {
	ret := L1Commitment{}
	var err error
	if ret.Commitment, err = CommitmentModel.VectorCommitmentFromBytes(data); err != nil {
		return L1Commitment{}, err
	}
	return ret, nil
}

func L1CommitmentFromAliasOutput(output *iotago.AliasOutput) (*L1Commitment, error) {
	l1c, err := L1CommitmentFromBytes(output.StateMetadata)
	if err != nil {
		return nil, err
	}
	return &l1c, nil
}

func (s *L1Commitment) Bytes() []byte {
	var buf bytes.Buffer

	buf.Write(s.Commitment.Bytes())
	return buf.Bytes()
}

func L1CommitmentFromAnchorOutput(o *iotago.AliasOutput) (L1Commitment, error) {
	return L1CommitmentFromBytes(o.StateMetadata)
}

var L1CommitmentNil *L1Commitment

func init() {
	var z [32]byte
	zs, err := L1CommitmentFromBytes(z[:])
	if err != nil {
		panic(err)
	}
	L1CommitmentNil = &zs
}
