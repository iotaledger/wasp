package blob

import (
	builtinutil "github.com/iotaledger/wasp/packages/vm/builtinvm/util"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "blob"
	Version     = "0.1"
	Description = "Blob Contract"
)

var (
	Interface = contract.ContractInterface{
		Name:        Name,
		Version:     Version,
		Description: Description,
		VMType:      "builtin",
		Functions: contract.Funcs(initialize, []contract.ContractFunctionInterface{
			contract.Func(FuncStoreBlob, storeBlob),
			contract.ViewFunc(FuncGetBlobInfo, getBlobInfo),
			contract.ViewFunc(FuncGetBlobField, getBlobField),
		}),
	}

	ProgramHash = builtinutil.BuiltinProgramHash(Name, Version)
	Hname       = builtinutil.BuiltinHname(Name, Version)
	FullName    = builtinutil.BuiltinFullName(Name, Version)

	// variable names of standard blob's field
	// user-defined field must be different
	VarFieldProgramBinary      = "p"
	VarFieldProgramDescription = "d"
	VarFieldProgramSource      = "s"
)

// state variables
const (
	VarStateInitialized = "i"
)

// param/return variables
const (
	ParamHash  = "hash"
	ParamField = "field"
	ParamBytes = "bytes"
)

// function names
const (
	FuncGetBlobInfo  = "getBlobInfo"
	FuncGetBlobField = "getBlobField"
	FuncStoreBlob    = "storeBlob"
)

func GetProcessor() vmtypes.Processor {
	return &Interface
}
