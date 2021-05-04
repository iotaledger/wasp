package blob

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = coreutil.CoreContractBlob
	description = "Blob Contract"
)

var (
	Interface = &coreutil.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncStoreBlob, storeBlob),
		coreutil.ViewFunc(FuncGetBlobInfo, getBlobInfo),
		coreutil.ViewFunc(FuncGetBlobField, getBlobField),
		coreutil.ViewFunc(FuncListBlobs, listBlobs),
	})
}

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

	// function names
	FuncGetBlobInfo  = "getBlobInfo"
	FuncGetBlobField = "getBlobField"
	FuncStoreBlob    = "storeBlob"
	FuncListBlobs    = "listBlobs"
)
