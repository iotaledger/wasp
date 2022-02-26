package iscp

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/kv/trie_merkle"
)

// StateData represents the parsed data stored as a metadata in the anchor output
type StateData struct {
	Commitment trie.VCommitment
}

func NewStateData(c trie.VCommitment) *StateData {
	return &StateData{Commitment: c}
}

func StateDataFromBytes(data []byte) (StateData, error) {
	ret := StateData{}
	var err error
	if ret.Commitment, err = trie_merkle.VectorCommitmentFromBytes(data); err != nil {
		return StateData{}, err
	}
	return ret, nil
}

func (s *StateData) Bytes() []byte {
	var buf bytes.Buffer

	buf.Write(s.Commitment.Bytes())
	return buf.Bytes()
}

var StateDataNil *StateData

func init() {
	var z [32]byte
	zs, err := StateDataFromBytes(z[:])
	if err != nil {
		panic(err)
	}
	StateDataNil = &zs
}
