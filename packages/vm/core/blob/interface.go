package blob

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
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

func GasForBlob(blob dict.Dict) uint64 {
	g := uint64(0)
	for k, v := range blob {
		g += uint64(len(k)) + uint64(len(v))
	}
	g = gas.StoreBytes(int(g))
	if g < gas.MinGasPerBlob {
		g = gas.MinGasPerBlob
	}
	return g
}
