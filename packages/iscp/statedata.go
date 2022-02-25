package iscp

import (
	"bytes"
	"encoding/hex"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/kv/trie_merkle"
)

// StateData represents the parsed data stored as a metadata in the anchor output
type StateData struct {
	Commitment trie.VCommitment
}

func StateDataFromBytes(data []byte) (StateData, error) {
	ret := StateData{}
	var err error
	if ret.Commitment, err = trie_merkle.NewVectorCommitmentFromBytes(data); err != nil {
		return StateData{}, err
	}
	return ret, nil
}

func (s *StateData) Bytes() []byte {
	var buf bytes.Buffer

	buf.Write(trie.MustBytes(s.Commitment))
	return buf.Bytes()
}

const OriginStateCommitmentHex = "5924dc2f04542fc93b02fa5c8b230f62110a9fbda78fca024cf58087bd32204f"

func OriginStateCommitment() trie.VCommitment {
	retBin, err := hex.DecodeString(OriginStateCommitmentHex)
	if err != nil {
		panic(err)
	}
	ret, err := trie_merkle.NewVectorCommitmentFromBytes(retBin)
	if err != nil {
		panic(err)
	}
	return ret
}
