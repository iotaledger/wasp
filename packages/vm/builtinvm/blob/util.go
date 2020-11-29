package blob

import (
	"fmt"
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

// GetBlobField retrieves blob field from the state or returns nil
func GetBlobField(state codec.ImmutableMustCodec, blobHash hashing.HashValue, field []byte) []byte {
	return state.GetMap(kv.Key(blobHash[:])).GetAt(field)
}

func LocateProgram(state codec.ImmutableMustCodec, programHash hashing.HashValue) (string, []byte, error) {
	fmt.Printf("--- LocateProgram: %s\n", programHash.String())
	programBinary := GetBlobField(state, programHash, []byte(VarFieldProgramBinary))
	if programBinary == nil {
		return "", nil, fmt.Errorf("can't find program binary for hash %s", programHash.String())
	}
	v := GetBlobField(state, programHash, []byte(VarFieldVMType))
	vmType := "wasmtimevm"
	if v != nil {
		vmType = string(v)
	}
	return vmType, programBinary, nil
}
