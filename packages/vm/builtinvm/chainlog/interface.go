package chainlog

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
)

const (
	Name        = "chainlog"
	Version     = "0.1"
	description = "Chainlog Contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        Name,
		Description: description,
		ProgramHash: *hashing.HashStrings(Name),
	}
)

func init() {
	Interface.WithFunctions(initialize, []contract.ContractFunctionInterface{
		contract.ViewFunc(FuncGetLogRecords, getLogRecords),
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
	FuncGetLogRecords = "getLogRecords"
	FuncGetNumRecords = "getNumRecords"

	DefaultMaxNumberOfRecords = 50
)
