package blob

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func mustGetBlobHash(fields dict.Dict) (hashing.HashValue, []kv.Key, [][]byte) {
	kSorted := fields.KeysSorted() // mind determinism
	values := make([][]byte, 0, len(kSorted))
	all := make([][]byte, 0, 2*len(kSorted))
	for _, k := range kSorted {
		v := fields.MustGet(k)
		values = append(values, v)
		all = append(all, v)
		all = append(all, []byte(k))
	}
	return *hashing.HashData(all...), kSorted, values
}

// MustGetBlobHash deterministically hashes map of binary values
func MustGetBlobHash(fields dict.Dict) hashing.HashValue {
	ret, _, _ := mustGetBlobHash(fields)
	return ret
}

// GetBlobField retrieves blob field from the state or returns nil
func GetBlobField(state kv.KVStore, blobHash hashing.HashValue, field []byte) []byte {
	return datatypes.NewMustMap(state, string(blobHash[:])).GetAt(field)
}

func LocateProgram(state kv.KVStore, programHash hashing.HashValue) (string, []byte, error) {
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
