package state

import (
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const BlockHashSize = 20

type BlockHash [BlockHashSize]byte

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

func BlockHashFromString(hash string) (BlockHash, error) {
	byteSlice, err := iotago.DecodeHex(hash)
	if err != nil {
		return BlockHash{}, err
	}
	var ret BlockHash
	copy(ret[:], byteSlice)
	return ret, nil
}

func (bh BlockHash) String() string {
	return iotago.EncodeHex(bh[:])
}

func (bh BlockHash) Equals(other BlockHash) bool {
	return bh == other
}

func L1CommitmentFromBytes(data []byte) (*L1Commitment, error) {
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

func init() {
	zs, err := L1CommitmentFromBytes(make([]byte, L1CommitmentSize))
	if err != nil {
		panic(err)
	}
	L1CommitmentNil = zs
}

// PseudoRandL1Commitment is for testing only
func PseudoRandL1Commitment() *L1Commitment {
	d := make([]byte, L1CommitmentSize)
	_, _ = util.NewPseudoRand().Read(d)
	ret, err := L1CommitmentFromBytes(d)
	if err != nil {
		panic(err)
	}
	return ret
}
