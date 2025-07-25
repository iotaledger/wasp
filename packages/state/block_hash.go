package state

import (
	"crypto/rand"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
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
	return hexutil.Encode(bh[:])
}

func (bh BlockHash) Equals(other BlockHash) bool {
	return bh == other
}

func RandomBlockHash() BlockHash {
	var b BlockHash
	_, _ = rand.Read(b[:])
	return b
}
