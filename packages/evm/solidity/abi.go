package solidity

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// StorageSlot returns the key for the given storage slot #.
// See https://docs.soliditylang.org/en/v0.8.16/internals/layout_in_storage.html
func StorageSlot(n uint8) common.Hash {
	return StorageEncodeUint8(n)
}

// StorageEncodeUint8 encodes an uint8 according to the storage spec.
func StorageEncodeUint8(n uint8) (ret common.Hash) {
	ret[len(ret)-1] = n
	return
}

// StorageEncodeShortString encodes a short string according to the storage spec.
func StorageEncodeShortString(s string) common.Hash {
	if len(s) > 31 {
		panic(fmt.Sprintf("string is too long: %q...", s[:8]))
	}
	ret := StorageEncodeUint8(uint8(len(s) * 2))
	copy(ret[:], s)
	return ret
}

// StorageEncodeString encodes a string according to the storage spec.
func StorageEncodeString(slotNumber uint8, s string) (ret map[common.Hash]common.Hash) {
	mainSlot := StorageSlot(slotNumber)
	ret = make(map[common.Hash]common.Hash)
	if len(s) <= 31 {
		ret[mainSlot] = StorageEncodeShortString(s)
		return
	}

	ret[mainSlot] = common.BigToHash(big.NewInt(int64(len(s)*2) + 1))

	i := 0
	for len(s) > 0 {
		var chunk common.Hash
		copy(chunk[:], s)

		// compute slot offset = keccak(slotNumber) + i
		slot := crypto.Keccak256Hash(common.BigToHash(big.NewInt(int64(slotNumber))).Bytes()).Big()
		slot = slot.Add(slot, big.NewInt(int64(i)))
		ret[common.BigToHash(slot)] = chunk

		if len(s) > 32 {
			s = s[32:]
		} else {
			s = ""
		}
		i++
	}
	return
}

// StorageEncodeBytes encodes a byte array according to the storage spec.
func StorageEncodeBytes(slotNumber uint8, b []byte) (ret map[common.Hash]common.Hash) {
	return StorageEncodeString(slotNumber, string(b))
}

// StorageEncodeUint256 encodes a uint256 value according to the storage spec.
func StorageEncodeUint256(n *big.Int) (ret common.Hash) {
	return common.BigToHash(n)
}

// StorageEncodeBytes32 encodes a bytes32 value according to the storage spec.
func StorageEncodeBytes32(b []byte) (ret common.Hash) {
	if len(b) != 32 {
		panic("expected len(b) == 32")
	}
	copy(ret[:], b)
	return
}
