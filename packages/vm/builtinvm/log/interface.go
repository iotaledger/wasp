package log

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/contract"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	Name        = "log"
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
	})
}

const (
	// state variables
	VarStateInitialized = "i"

	// request parameters
	ParamHash  = "hash"
	ParamField = "field"
	ParamBytes = "bytes"
	ParamLog   = "logData"

	// function names
	FuncGetLog   = "getLogInfo"
	FuncStoreLog = "storeLog"

	LogName = "logs"
)

func GetProcessor() vmtypes.Processor {
	return Interface
}
