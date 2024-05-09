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
	FuncStoreBlob = Contract.Func("storeBlob")

	ViewGetBlobInfo = coreutil.NewViewEP11(Contract, "getBlobInfo",
		coreutil.FieldWithCodec(ParamHash, codec.HashValue),
		OutputFieldSizesMap{},
	)
	ViewGetBlobField = coreutil.NewViewEP21(Contract, "getBlobField",
		coreutil.FieldWithCodec(ParamHash, codec.HashValue),
		coreutil.FieldWithCodec(ParamField, codec.Bytes),
		coreutil.FieldWithCodec(ParamBytes, codec.Bytes),
	)
	ViewListBlobs = coreutil.NewViewEP01(Contract, "listBlobs",
		OutputBlobDirectory{},
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

func (_ OutputFieldSizesMap) Encode(sizes map[string]uint32) dict.Dict {
	return lo.MapEntries(sizes, func(field string, size uint32) (kv.Key, []byte) {
		return kv.Key(field), EncodeSize(size)
	})
}

func (_ OutputFieldSizesMap) Decode(r dict.Dict) (map[string]uint32, error) {
	return decodeSizesMap(r)
}

type OutputBlobDirectory struct{}

func (_ OutputBlobDirectory) Encode(d map[hashing.HashValue]uint32) dict.Dict {
	return lo.MapEntries(d, func(hash hashing.HashValue, size uint32) (kv.Key, []byte) {
		return kv.Key(codec.HashValue.Encode(hash)), EncodeSize(size)
	})
}

func (_ OutputBlobDirectory) Decode(r dict.Dict) (map[hashing.HashValue]uint32, error) {
	return decodeDirectory(r)
}
