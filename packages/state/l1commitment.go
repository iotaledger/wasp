package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"io"
	"math/rand"
)

// L1Commitment represents parsed data stored as a metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	StateCommitment trie.VCommitment
	// hash of the essence of the last block
	BlockHash hashing.HashValue
}

const (
	// L1CommitmentSizeBlake2b is size of the L1 commitment with blake2b mode = 32 bytes for the merkle root + 32 bytes block hash
	L1CommitmentSizeBlake2b  = 64
	OriginStateCommitmentHex = "5924dc2f04542fc93b02fa5c8b230f62110a9fbda78fca024cf58087bd32204f"
)

func NewL1Commitment(c trie.VCommitment, blockHash hashing.HashValue) *L1Commitment {
	return &L1Commitment{
		StateCommitment: c,
		BlockHash:       blockHash,
	}
}

func L1CommitmentFromBytes(data []byte) (L1Commitment, error) {
	if len(data) != L1CommitmentSizeBlake2b {
		return L1Commitment{}, xerrors.New("L1CommitmentFromBytes: wrong data length")
	}
	ret := L1Commitment{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
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
	return util.MustBytes(s)
}

func (s *L1Commitment) Write(w io.Writer) error {
	if err := s.StateCommitment.Write(w); err != nil {
		return err
	}
	if _, err := w.Write(s.BlockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) Read(r io.Reader) error {
	s.StateCommitment = CommitmentModel.NewVectorCommitment()
	if err := s.StateCommitment.Read(r); err != nil {
		return err
	}
	l, err := r.Read(s.BlockHash[:])
	if err != nil {
		return err
	}
	if l != 32 {
		return xerrors.New("wrong data length")
	}
	return nil
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("L1Commitment(%s, %s)", s.StateCommitment.String(), s.BlockHash.String())
}

func L1CommitmentFromAnchorOutput(o *iotago.AliasOutput) (L1Commitment, error) {
	return L1CommitmentFromBytes(o.StateMetadata)
}

var L1CommitmentNil *L1Commitment

func init() {
	var emptyBytes [64]byte
	zs, err := L1CommitmentFromBytes(emptyBytes[:])
	if err != nil {
		panic(err)
	}
	L1CommitmentNil = &zs
}

func OriginStateCommitment() trie.VCommitment {
	retBin, err := hex.DecodeString(OriginStateCommitmentHex)
	if err != nil {
		panic(err)
	}
	ret, err := CommitmentModel.VectorCommitmentFromBytes(retBin)
	if err != nil {
		panic(err)
	}
	return ret
}

func OriginBlockHash() (ret [32]byte) {
	return
}

func OriginL1Commitment() *L1Commitment {
	return NewL1Commitment(OriginStateCommitment(), OriginBlockHash())
}

func RandL1Commitment() *L1Commitment {
	var d [64]byte
	rand.Read(d[:])
	ret, err := L1CommitmentFromBytes(d[:])
	if err != nil {
		panic(err)
	}
	return &ret
}
