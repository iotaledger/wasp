package chainlog

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
		contract.ViewFunc(FuncGetLogsBetweenTs, getLogsBetweenTs),
	})
	examples.AddProcessor(Interface.ProgramHash, Interface)

}

const (
	// state variables
	VarStateInitialized = "i" //+++ not used, delete

	//+++ where is the name of the tlog itself?
	// request parameters
	ParamLog           = "dataParam" // ParamRecordData is better?
	ParamContractHname = "contractHname"
	ParamLasts         = "lastsParam" //+++ not used, delete
	ParamFromTs        = "fromTs"
	ParamToTs          = "toTs"
	ParamLastsRecords  = "lastRecords"
	ParamType          = "ParamTypeOfRecords" // better ParamRecordType?

	// function names
	FuncGetLogsBetweenTs = "getLogsBetweenTs"
	FuncStoreLog         = "storeLog"

	//Type of records
	// +++ Go type of the record type code should be uint16
	// +++ are these record types defined at the system level? Doc-comments are needed
	TR_DEPLOY         = 1
	TR_TOKEN_TRANSFER = 2
	TR_VIEWCALL       = 3
	TR_REQUEST_FUNC   = 4
	TR_GENERIC_DATA   = 5
)

//+++ not needed
func GetProcessor() vmtypes.Processor {
	return Interface
}
