package state

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/common"
	"github.com/iotaledger/wasp/packages/util"
)

const BlockHashSize = 20

type BlockHash [BlockHashSize]byte

// L1Commitment represents the data stored as metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	trieRoot common.VCommitment
	// hash of the essence of the block
	blockHash BlockHash
}

var l1CommitmentSize = len(newL1Commitment(commitmentModel.NewVectorCommitment(), BlockHash{}).Bytes())

func BlockHashFromData(data []byte) (ret BlockHash) {
	r := blake2b.Sum256(data)
	copy(ret[:BlockHashSize], r[:BlockHashSize])
	return
}

func newL1Commitment(c common.VCommitment, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		trieRoot:  c,
		blockHash: blockHash,
	}
}

func (bh BlockHash) String() string {
	return iotago.EncodeHex(bh[:])
}

func (bh BlockHash) Equals(bh2 BlockHash) bool {
	return bh == bh2
}

func L1CommitmentFromBytes(data []byte) (*L1Commitment, error) {
	if len(data) != l1CommitmentSize {
		return nil, xerrors.New("L1CommitmentFromBytes: wrong data length")
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

func (s *L1Commitment) GetTrieRoot() common.VCommitment {
	return s.trieRoot
}

func (s *L1Commitment) GetBlockHash() BlockHash {
	return s.blockHash
}

func (s *L1Commitment) Equals(other *L1Commitment) bool {
	return s.GetBlockHash().Equals(other.GetBlockHash()) && EqualCommitments(s.GetTrieRoot(), other.GetTrieRoot())
}

func (s *L1Commitment) Bytes() []byte {
	return util.MustBytes(s)
}

func (s *L1Commitment) Write(w io.Writer) error {
	if err := s.GetTrieRoot().Write(w); err != nil {
		return err
	}
	blockHash := s.GetBlockHash()
	if _, err := w.Write(blockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) Read(r io.Reader) error {
	s.trieRoot = commitmentModel.NewVectorCommitment()
	if err := s.trieRoot.Read(r); err != nil {
		return err
	}
	if _, err := r.Read(s.blockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("trie root: %s, block hash: %s", s.GetTrieRoot(), s.GetBlockHash())
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
