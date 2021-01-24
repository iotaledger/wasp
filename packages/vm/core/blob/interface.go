package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "blob"
	description = "Blob Contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.Func(FuncStoreBlob, storeBlob),
		contract.ViewFunc(FuncGetBlobInfo, getBlobInfo),
		contract.ViewFunc(FuncGetBlobField, getBlobField),
		contract.ViewFunc(FuncListBlobs, listBlobs),
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
