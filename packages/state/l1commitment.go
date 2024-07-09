package state

import (
	"fmt"
	"io"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

// L1Commitment represents the data stored as metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	trieRoot trie.Hash
	// hash of the essence of the block
	blockHash BlockHash
}

const L1CommitmentSize = trie.HashSizeBytes + BlockHashSize

func newL1Commitment(c trie.Hash, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		trieRoot:  c,
		blockHash: blockHash,
	}
}

func NewL1CommitmentFromAnchor(anchor *iscmove.Anchor) (*L1Commitment, error) {
	trieRoot, err := trie.HashFromBytes(anchor.StateRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create trie root: %w", err)
	}
	blockHash, err := NewBlockHash(anchor.BlockHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create block hash: %w", err)
	}
	return &L1Commitment{
		trieRoot:  trieRoot,
		blockHash: blockHash,
	}, nil
}

func NewL1CommitmentFromBytes(data []byte) (*L1Commitment, error) {
	return rwutil.ReadFromBytes(data, new(L1Commitment))
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
	return rwutil.WriteToBytes(s)
}

func (s *L1Commitment) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(s.trieRoot[:])
	rr.ReadN(s.blockHash[:])
	return rr.Err
}

func (s *L1Commitment) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(s.trieRoot[:])
	ww.WriteN(s.blockHash[:])
	return ww.Err
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("<%s;%s>", s.TrieRoot(), s.BlockHash())
}

var L1CommitmentNil = &L1Commitment{}

/*func init() {
	zs, err := L1CommitmentFromBytes(make([]byte, L1CommitmentSize))
	if err != nil {
		panic(err)
	}
	L1CommitmentNil = zs
}*/

// PseudoRandL1Commitment is for testing only
func NewPseudoRandL1Commitment() *L1Commitment {
	d := make([]byte, L1CommitmentSize)
	_, _ = util.NewPseudoRand().Read(d)
	ret, err := NewL1CommitmentFromBytes(d)
	if err != nil {
		panic(err)
	}
	return ret
}
