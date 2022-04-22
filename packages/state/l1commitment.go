package state

import (
	"bytes"
	"encoding/hex"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"golang.org/x/xerrors"
	"math/rand"
)

// L1Commitment represents parsed data stored as a metadata in the anchor output
type L1Commitment struct {
	// root commitment to the state
	StateCommitment trie.VCommitment
	// hash of the las block
	BlockHash [32]byte
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
	ret := L1Commitment{}
	ret.StateCommitment = CommitmentModel.NewVectorCommitment()
	rdr := bytes.NewReader(data)
	if err := ret.StateCommitment.Read(rdr); err != nil {
		return L1Commitment{}, err
	}
	l, err := rdr.Read(ret.BlockHash[:])
	if err != nil {
		return L1Commitment{}, err
	}
	if l != 32 {
		return L1Commitment{}, xerrors.New("wrong data length")
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

	_ = s.StateCommitment.Write(&buf)
	_, _ = buf.Write(s.BlockHash[:])

	return buf.Bytes()
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
