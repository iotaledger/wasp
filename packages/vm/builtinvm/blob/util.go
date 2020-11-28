package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func mustGetBlobHash(fields codec.ImmutableCodec) (hashing.HashValue, []kv.Key, [][]byte) {
	kSorted, err := fields.KeysSorted() // mind determinism
	if err != nil {
		panic(err)
	}
	values := make([][]byte, 0, len(kSorted))
	for _, k := range kSorted {
		v, err := fields.Get(k)
		if err != nil {
			panic(err)
		}
		values = append(values, v)
	}
	return *hashing.HashData(values...), kSorted, values
}

// MustGetBlobHash deterministically hashes map of binary values
func MustGetBlobHash(fields codec.ImmutableCodec) hashing.HashValue {
	ret, _, _ := mustGetBlobHash(fields)
	return ret
}
