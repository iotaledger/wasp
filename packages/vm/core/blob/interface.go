package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

var Contract = coreutil.NewContract(coreutil.CoreContractBlob, "Blob Contract")

const (
	// request parameters
	ParamHash  = "hash"
	ParamField = "field"
	ParamBytes = "bytes"

	// variable names of standard blob's field
	// user-defined field must be different
	VarFieldProgramBinary      = "p"
	VarFieldVMType             = "v"
	VarFieldProgramDescription = "d"
)

var (
	FuncStoreBlob    = coreutil.Func("storeBlob")
	FuncGetBlobInfo  = coreutil.ViewFunc("getBlobInfo")
	FuncGetBlobField = coreutil.ViewFunc("getBlobField")
	FuncListBlobs    = coreutil.ViewFunc("listBlobs")
)

// FieldValueKey returns key of the blob field value in the SC state.
func FieldValueKey(blobHash hashing.HashValue, fieldName string) []byte {
	return collections.MapElemKey(valuesMapName(blobHash), []byte(fieldName))
}
