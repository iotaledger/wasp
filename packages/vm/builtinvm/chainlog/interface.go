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

	//Type of records
	//Constantes que definen de que tipo son los datos logueados. Se definen distintos tipos:
	//	-TRDeploy       -> Cada vez que se quiera loguear un deploy
	//	-TRViewCall     -> Cada vez que se quiera loguear una viewcall a un sc externo
	//	-TRRequest -> Cada vez que se quiera loguear una request a un sc
	//	-TRGenericData -> Cada vez que se quiera  loguear datos genericos establecidos por el usuario
	TRDeploy      = byte(1)
	TRViewCall    = byte(2)
	TRRequest     = byte(3)
	TRGenericData = byte(4)
)

func GetProcessor() vmtypes.Processor {
	return Interface
}
