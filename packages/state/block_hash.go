package state

import (
	"crypto/rand"
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

const BlockHashSize = 20

type BlockHash [BlockHashSize]byte

func NewBlockHash(hash []byte) (BlockHash, error) {
	if len(hash) != BlockHashSize {
		return BlockHash{}, fmt.Errorf("array size is wrong: expected %v, got %v", BlockHashSize, len(hash))
	}
	result := BlockHash{}
	copy(result[:], hash)
	return result, nil
}

/*func newL1Commitment(c trie.Hash, blockHash BlockHash) *L1Commitment {
	return &L1Commitment{
		trieRoot:  c,
		blockHash: blockHash,
	}
}*/

func BlockHashFromString(hash string) (BlockHash, error) {
	byteSlice, err := cryptolib.DecodeHex(hash)
	if err != nil {
		return BlockHash{}, err
	}
	var ret BlockHash
	copy(ret[:], byteSlice)
	return ret, nil
}

func (bh BlockHash) Bytes() []byte {
	return bh[:]
}

func (bh BlockHash) String() string {
	return iotago.EncodeHex(bh[:])
}

func (bh BlockHash) Equals(other BlockHash) bool {
	return bh == other
}

func RandomBlockHash() BlockHash {
	var b BlockHash
	_, _ = rand.Read(b[:])
	return b
}
