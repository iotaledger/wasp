package state

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/util"
)

const BlockHashSize = 20

type BlockHash [BlockHashSize]byte

// L1Commitment represents parsed data stored as a metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	StateCommitment trie.VCommitment
	// hash of the essence of the last block
	BlockHash BlockHash
}

const (
	OriginStateCommitmentHex = "c4f09061cd63ea506f89b7cbb3c6e0984f124158"
)

var l1CommitmentSize = len(NewL1Commitment(model.NewVectorCommitment(), BlockHash{}).Bytes())

func BlockHashFromData(data []byte) (ret BlockHash) {
	r := blake2b.Sum256(data)
	copy(ret[:BlockHashSize], r[:BlockHashSize])
	return
}

func NewL1Commitment(c trie.VCommitment, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		StateCommitment: c,
		BlockHash:       blockHash,
	}
}

func (bh BlockHash) String() string {
	return iotago.EncodeHex(bh[:])
}

func L1CommitmentFromBytes(data []byte) (L1Commitment, error) {
	if len(data) != l1CommitmentSize {
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
	s.StateCommitment = model.NewVectorCommitment()
	if err := s.StateCommitment.Read(r); err != nil {
		return err
	}
	if _, err := r.Read(s.BlockHash[:]); err != nil {
		return err
	}
	return nil
}

func (s *L1Commitment) String() string {
	return fmt.Sprintf("L1Commitment(%s, %s)", s.StateCommitment.String(), iotago.EncodeHex(s.BlockHash[:]))
}

func L1CommitmentFromAnchorOutput(o *iotago.AliasOutput) (L1Commitment, error) {
	return L1CommitmentFromBytes(o.StateMetadata)
}

var L1CommitmentNil *L1Commitment

func init() {
	zs, err := L1CommitmentFromBytes(make([]byte, l1CommitmentSize))
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
	c := model.NewVectorCommitment()
	if err = c.Read(bytes.NewReader(retBin)); err != nil {
		panic(err)
	}
	return c
}

func OriginBlockHash() (ret BlockHash) {
	return
}

func OriginL1Commitment() *L1Commitment {
	return NewL1Commitment(OriginStateCommitment(), OriginBlockHash())
}

// RandL1Commitment for testing only
func RandL1Commitment() *L1Commitment {
	d := make([]byte, l1CommitmentSize)
	rand.Read(d)
	ret, err := L1CommitmentFromBytes(d)
	if err != nil {
		panic(err)
	}
	return &ret
}
