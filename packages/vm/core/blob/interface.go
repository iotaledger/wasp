package blob

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
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
