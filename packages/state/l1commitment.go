package state

import (
	"bytes"
	"fmt"

	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// L1Commitment represents the data stored as metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	trieRoot trie.Hash `bcs:"export"`
	// hash of the essence of the block
	blockHash BlockHash `bcs:"export"`
}

const L1CommitmentSize = trie.HashSizeBytes + BlockHashSize

func newL1Commitment(c trie.Hash, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		trieRoot:  c,
		blockHash: blockHash,
	}
}

func NewL1CommitmentFromBytes(data []byte) (*L1Commitment, error) {
	return bcs.Unmarshal[*L1Commitment](data)
}

func (s *L1Commitment) TrieRoot() trie.Hash {
	return s.trieRoot
}

func (s *L1Commitment) BlockHash() BlockHash {
	return s.blockHash
}

func (s *L1Commitment) Equals(other *L1Commitment) bool {
	return s.blockHash.Equals(other.blockHash) && s.trieRoot.Equals(other.trieRoot)
}

func (s *L1Commitment) Bytes() []byte {
	return bcs.MustMarshal(s)
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("<%s;%s>", s.TrieRoot(), s.BlockHash())
}

func (s *L1Commitment) IsZero() bool {
	return !bytes.ContainsFunc(s.Bytes(), func(r rune) bool {
		return r >= 1
	})
}
