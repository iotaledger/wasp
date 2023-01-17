package solidity

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
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
func StorageEncodeShortString(s string) (ret common.Hash) {
	if len(s) > 31 {
		panic(fmt.Sprintf("string is too long: %q...", s[:8]))
	}
	ret[len(ret)-1] = uint8(len(s) * 2)
	copy(ret[:], s)
	return
}

// StorageEncodeBytes32 encodes a bytes32 value according to the storage spec.
func StorageEncodeBytes32(b []byte) (ret common.Hash) {
	if len(b) != 32 {
		panic("expected len(b) == 32")
	}
	copy(ret[:], b)
	return
}
