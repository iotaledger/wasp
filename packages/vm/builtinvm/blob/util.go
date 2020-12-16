package blob

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
)

const varStateDirectory = "d"

const valuesPrefix = "v"
const sizesPrefix = "s"

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

// GetDirectory retrieves the blob directory from the state
func GetDirectory(state kv.KVStore) *datatypes.MustMap {
	return datatypes.NewMustMap(state, varStateDirectory)
}

// GetBlobValues retrieves the blob field-value map from the state
func GetBlobValues(state kv.KVStore, blobHash hashing.HashValue) *datatypes.MustMap {
	return datatypes.NewMustMap(state, valuesPrefix+string(blobHash[:]))
}

// GetBlobSize retrieves the blob field-size map from the state
func GetBlobSizes(state kv.KVStore, blobHash hashing.HashValue) *datatypes.MustMap {
	return datatypes.NewMustMap(state, sizesPrefix+string(blobHash[:]))
}

func LocateProgram(state kv.KVStore, programHash hashing.HashValue) (string, []byte, error) {
	fmt.Printf("--- LocateProgram: %s\n", programHash.String())
	blbValues := GetBlobValues(state, programHash)
	programBinary := blbValues.GetAt([]byte(VarFieldProgramBinary))
	if programBinary == nil {
		return "", nil, fmt.Errorf("can't find program binary for hash %s", programHash.String())
	}
	v := blbValues.GetAt([]byte(VarFieldVMType))
	vmType := "wasmtimevm"
	if v != nil {
		vmType = string(v)
	}
	return vmType, programBinary, nil
}

func EncodeSize(size uint32) []byte {
	return util.Uint32To4Bytes(size)
}

func DecodeSize(size []byte) (uint32, error) {
	return util.Uint32From4Bytes(size)
}

func DecodeSizesMap(sizes dict.Dict) (map[string]uint32, error) {
	ret := make(map[string]uint32)
	for field, size := range sizes {
		v, err := DecodeSize(size)
		if err != nil {
			return nil, err
		}
		ret[string(field)] = uint32(v)
	}
	return ret, nil
}
