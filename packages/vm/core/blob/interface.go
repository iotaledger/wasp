package blob

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlob)

var (
	FuncStoreBlob = coreutil.NewEP11(Contract, "storeBlob", coreutil.FieldWithCodec(codec.Dict), coreutil.FieldWithCodec(codec.HashValue))

	ViewGetBlobInfo = coreutil.NewViewEP11(Contract, "getBlobInfo",
		coreutil.FieldWithCodec(codec.HashValue),
		OutputFieldSizesMap{},
	)
	ViewGetBlobField = coreutil.NewViewEP21(Contract, "getBlobField",
		coreutil.FieldWithCodec(codec.HashValue),
		coreutil.FieldWithCodec(codec.Bytes),
		coreutil.FieldWithCodec(codec.Bytes),
	)
)

// state variables
const (
	// variable names of standard blob's field
	// user-defined field must be different
	VarFieldProgramBinary      = "p"
	VarFieldVMType             = "v"
	VarFieldProgramDescription = "d"
)

// request parameters
const (
	ParamHash  = "hash"
	ParamField = "field"
	ParamBytes = "bytes"
)

// FieldValueKey returns key of the blob field value in the SC state.
func FieldValueKey(blobHash hashing.HashValue, fieldName string) []byte {
	return []byte(collections.MapElemKey(valuesMapName(blobHash), []byte(fieldName)))
}

type OutputFieldSizesMap struct{}

func (OutputFieldSizesMap) Encode(sizes map[string]uint32) []byte {
	var entries dict.Dict = lo.MapEntries(sizes, func(field string, size uint32) (kv.Key, []byte) {
		return kv.Key(field), EncodeSize(size)
	})

	return entries.Bytes()
}

func (OutputFieldSizesMap) Decode(r []byte) (map[string]uint32, error) {
	result, err := dict.FromBytes(r)

	if err != nil {
		return nil, err
	}
	return decodeSizesMap(result)
}

type OutputBlobDirectory struct{}

func (OutputBlobDirectory) Encode(d map[hashing.HashValue]uint32) dict.Dict {
	return lo.MapEntries(d, func(hash hashing.HashValue, size uint32) (kv.Key, []byte) {
		return kv.Key(codec.HashValue.Encode(hash)), EncodeSize(size)
	})
}

func (OutputBlobDirectory) Decode(r dict.Dict) (map[hashing.HashValue]uint32, error) {
	return decodeDirectory(r)
}
