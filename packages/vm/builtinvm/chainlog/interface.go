package log

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/examples"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "chainlog"
	Version     = "0.1"
	fullName    = Name + "-" + Version
	description = "Log Contract"
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
		contract.Func(FuncStoreLog, storeLog),
		contract.ViewFunc(FuncGetLog, getLogInfo),
		contract.ViewFunc(FuncGetLasts, getLasts),
	})
	examples.AddProcessor(Interface.ProgramHash, Interface)

}

const (
	// state variables
	VarStateInitialized = "i"

	// request parameters
	ParamHash  = "hash"
	ParamField = "field"
	ParamBytes = "bytes"
	ParamLog   = "dataParam"
	ParamLasts = "lastsParam"

	// function names
	FuncGetLog   = "getLogInfo"
	FuncGetLasts = "getLasts"
	FuncStoreLog = "storeLog"

	VarLogName = "logs"
)

func GetProcessor() vmtypes.Processor {
	return Interface
}
