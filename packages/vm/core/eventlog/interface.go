package eventlog

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "eventlog"
	description = "Event log Contract"
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
		contract.ViewFunc(FuncGetRecords, getRecords),
		contract.ViewFunc(FuncGetNumRecords, getNumRecords),
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
