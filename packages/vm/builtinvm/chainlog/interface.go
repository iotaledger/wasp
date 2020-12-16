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
		contract.ViewFunc(FuncGetLogsBetweenTs, getLogsBetweenTs),
		contract.ViewFunc(FuncLastsNRecords, getLastsNRecords),
		contract.ViewFunc(FuncLenByHnameAndTR, getLenByHnameAndTR),
	})
}

const (
	// request parameters
	ParamRecordData    = "recordData"
	ParamContractHname = "contractHname"
	ParamFromTs        = "fromTs"
	ParamToTs          = "toTs"
	ParamLastsRecords  = "lastRecords"
	ParamRecordType    = "ParamTypeOfRecords"

	// function names
	FuncGetLogsBetweenTs = "getLogsBetweenTs"
	FuncLastsNRecords    = "getLastsNRecords"
	FuncLenByHnameAndTR  = "getLenByHnameAndTR"

	// Type of records
	// Constants that define the type of logged data. Different types are defined:
	// -TRDeploy -> Every time you want to log a contract deploy (in root)
	// -TRViewCall -> Every time you want to log a viewcall, e.g an external sc
	// -TRRequest -> Every time you want to log a request to a sc
	// -TRGenericData -> Every time you want to log generic data set by the user
	TRDeploy      = byte(1)
	TRViewCall    = byte(2)
	TRRequest     = byte(3)
	TRGenericData = byte(4)
)

func GetProcessor() vmtypes.Processor {
	return Interface
}
