package chainlog

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "chainlog"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Chainlog Contract"
)

var (
	Interface = &contract.ContractInterface{
		Name:        fullName,
		Description: description,
		ProgramHash: *hashing.HashStrings(fullName),
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

func GetProcessor() vmtypes.Processor {
	return Interface
}

const MaxRequestError = 100
