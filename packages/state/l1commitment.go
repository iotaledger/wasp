package state

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util"
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

var l1CommitmentSize = trie.HashSizeBytes + BlockHashSize

func BlockHashFromData(data []byte) (ret BlockHash) {
	r := blake2b.Sum256(data)
	copy(ret[:BlockHashSize], r[:BlockHashSize])
	return
}

func BlockHashFromString(blockHashString string) (BlockHash, error) {
	result := BlockHash{}
	slice, err := iotago.DecodeHex(blockHashString)
	if err != nil {
		return result, fmt.Errorf("Error decoding block hash from string %s: %w", blockHashString, err)
	}
	if len(slice) != BlockHashSize {
		return result, fmt.Errorf("Error decoding block hash from string %s: %v bytes obtained; expected %v bytes", blockHashString, len(slice), BlockHashSize)
	}
	copy(result[:], slice)
	return result, nil
}

func newL1Commitment(c trie.Hash, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		trieRoot:  c,
		blockHash: blockHash,
	}
}

func (bh BlockHash) String() string {
	return iotago.EncodeHex(bh[:])
}

func (bh BlockHash) Equals(other BlockHash) bool {
	return bh == other
}

func L1CommitmentFromBytes(data []byte) (*L1Commitment, error) {
	if len(data) != l1CommitmentSize {
		return nil, errors.New("L1CommitmentFromBytes: wrong data length")
	}
	ret := L1Commitment{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return &ret, nil
}

func L1CommitmentFromMarshalUtil(mu *marshalutil.MarshalUtil) (*L1Commitment, error) {
	byteCount, err := mu.ReadUint16()
	if err != nil {
		return nil, err
	}
	data, err := mu.ReadBytes(int(byteCount))
	if err != nil {
		return nil, err
	}
	l1c, err := L1CommitmentFromBytes(data)
	if err != nil {
		return nil, err
	}
	return l1c, nil
}

func L1CommitmentFromAliasOutput(output *iotago.AliasOutput) (*L1Commitment, error) {
	l1c, err := L1CommitmentFromBytes(output.StateMetadata)
	if err != nil {
		return nil, err
	}
	return l1c, nil
}

func (s *L1Commitment) TrieRoot() trie.Hash {
	return s.trieRoot
}

func (s *L1Commitment) BlockHash() BlockHash {
	return s.blockHash
}

func (s *L1Commitment) Equals(other *L1Commitment) bool {
	return s.blockHash == other.blockHash && s.trieRoot == other.trieRoot
}

func (s *L1Commitment) Bytes() []byte {
	return util.MustBytes(s)
}

func (s *L1Commitment) Write(w io.Writer) error {
	hash := s.TrieRoot()
	if err := hash.Write(w); err != nil {
		return err
	}
	blockHash := s.BlockHash()
	if _, err := w.Write(blockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) Read(r io.Reader) error {
	var err error
	s.trieRoot, err = trie.ReadHash(r)
	if err != nil {
		return err
	}
	if _, err := r.Read(s.blockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("<%s;%s>", s.TrieRoot(), s.BlockHash())
}

func L1CommitmentFromAnchorOutput(o *iotago.AliasOutput) (*L1Commitment, error) {
	return L1CommitmentFromBytes(o.StateMetadata)
}

var L1CommitmentNil *L1Commitment

func init() {
	zs, err := L1CommitmentFromBytes(make([]byte, l1CommitmentSize))
	if err != nil {
		panic(err)
	}
	L1CommitmentNil = zs
}

// RandL1Commitment is for testing only
func RandL1Commitment() *L1Commitment {
	d := make([]byte, l1CommitmentSize)
	rand.Read(d)
	ret, err := L1CommitmentFromBytes(d)
	if err != nil {
		panic(err)
	}
	return ret
}
