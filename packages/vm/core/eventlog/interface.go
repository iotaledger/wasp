package eventlog

import (
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	Name        = coreutil.CoreContractEventlog
	description = "Event log Contract"
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
		coreutil.ViewFunc(FuncGetRecords, getRecords),
		coreutil.ViewFunc(FuncGetNumRecords, getNumRecords),
	})
}

const (
	// request parameters
	ParamContractHname  = "contractHname"
	ParamFromTs         = "fromTs"
	ParamToTs           = "toTs"
	ParamMaxLastRecords = "maxLastRecords"
	ParamNumRecords     = "numRecords"
	ParamRecords        = "records"

	// function names
	FuncGetRecords    = "getRecords"
	FuncGetNumRecords = "getNumRecords"

	DefaultMaxNumberOfRecords = 50
)
