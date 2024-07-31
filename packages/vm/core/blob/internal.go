package blob

import (
	"encoding/binary"
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

const directoryPrefix = "d"

func valuesMapName(blobHash hashing.HashValue) string {
	return "v" + string(blobHash[:])
}

func sizesMapName(blobHash hashing.HashValue) string {
	return "s" + string(blobHash[:])
}

func mustGetBlobHash(fields dict.Dict) (hashing.HashValue, []kv.Key, [][]byte) {
	sorted := fields.KeysSorted() // mind determinism
	values := make([][]byte, 0, len(sorted))
	all := make([][]byte, 0, 2*len(sorted))

	// hashBlob = hash(KeyLen0|Key0|Val0 | KeyLen1|Key1|Val1 | ... | KeyLenN|KeyN|ValN)
	// by prepend the key length we can avoid the possible collision
	for _, key := range sorted {
		var prefix [4]byte
		v := fields.Get(key)
		values = append(values, v)
		binary.LittleEndian.PutUint32(prefix[:], uint32(len(key)))
		all = append(all, prefix[:], []byte(key), v)
	}
	return hashing.HashData(all...), sorted, values
}

// MustGetBlobHash deterministically hashes map of binary values
func MustGetBlobHash(fields dict.Dict) hashing.HashValue {
	ret, _, _ := mustGetBlobHash(fields)
	return ret
}

// GetDirectory retrieves the blob directory from the state
func (s *StateWriter) GetDirectory() *collections.Map {
	return collections.NewMap(s.state, directoryPrefix)
}

// GetDirectory retrieves the blob directory from the state
func (s *StateReader) GetDirectory() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, directoryPrefix)
}

// GetDirectoryR retrieves the blob directory from the read-only state
func (s *StateReader) contractStateGetDirectory() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, directoryPrefix)
}

// GetBlobValues retrieves the blob field-value map from the state
func (s *StateWriter) contractStateGetBlobValues(blobHash hashing.HashValue) *collections.Map {
	return collections.NewMap(s.state, valuesMapName(blobHash))
}

// GetBlobValues retrieves the blob field-value map from the read-only state
func (s *StateReader) GetBlobValues(blobHash hashing.HashValue) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, valuesMapName(blobHash))
}

// GetBlobValues retrieves the blob field-value map from the read-only state
func (s *StateWriter) GetBlobValues(blobHash hashing.HashValue) *collections.Map {
	return collections.NewMap(s.state, valuesMapName(blobHash))
}

// GetBlobSizes retrieves the writeable blob field-size map from the state
func (s *StateWriter) GetBlobSizes(blobHash hashing.HashValue) *collections.Map {
	return collections.NewMap(s.state, sizesMapName(blobHash))
}

// GetBlobSizes retrieves the blob field-size map from the read-only state
func (s *StateReader) GetBlobSizes(blobHash hashing.HashValue) *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, sizesMapName(blobHash))
}

// Note: only used in tests
func ListBlobs(state kv.KVStoreReader) map[kv.Key]uint32 {
	partition := subrealm.NewReadOnly(state, kv.Key(Contract.Hname().Bytes()))
	r := NewStateReader(partition)
	ret := make(map[kv.Key]uint32)
	r.contractStateGetDirectory().Iterate(func(hash []byte, totalSize []byte) bool {
		ret[kv.Key(hash)] = codec.Uint32.MustDecode(totalSize)
		return true
	})
	return ret
}

func (s *StateReader) LocateProgram(programHash hashing.HashValue) (string, []byte, error) {
	blbValues := s.GetBlobValues(programHash)
	programBinary := blbValues.GetAt([]byte(VarFieldProgramBinary))
	if programBinary == nil {
		return "", nil, fmt.Errorf("can't find program binary for hash %s", programHash.String())
	}
	v := blbValues.GetAt([]byte(VarFieldVMType))
	vmType := ""
	if v != nil {
		vmType = string(v)
	}
	return vmType, programBinary, nil
}

func EncodeSize(size uint32) []byte {
	return codec.Uint32.Encode(size)
}

func DecodeSize(size []byte) (uint32, error) {
	return codec.Uint32.Decode(size)
}

func decodeSizesMap(sizes dict.Dict) (map[string]uint32, error) {
	ret := make(map[string]uint32)
	for field, size := range sizes {
		v, err := DecodeSize(size)
		if err != nil {
			return nil, err
		}
		ret[string(field)] = v
	}
	return ret, nil
}

func decodeDirectory(blobs dict.Dict) (map[hashing.HashValue]uint32, error) {
	ret := make(map[hashing.HashValue]uint32)
	for hash, size := range blobs {
		v, err := DecodeSize(size)
		if err != nil {
			return nil, err
		}
		h, err := codec.HashValue.Decode([]byte(hash))
		if err != nil {
			return nil, err
		}
		ret[h] = v
	}
	return ret, nil
}
