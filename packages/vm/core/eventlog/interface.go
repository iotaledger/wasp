package eventlog

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractEventlog, "Event log Contract")

const (
	// request parameters
	ParamContractHname  = "contractHname"
	ParamFromTs         = "fromTs"
	ParamToTs           = "toTs"
	ParamMaxLastRecords = "maxLastRecords"
	ParamNumRecords     = "numRecords"
	ParamRecords        = "records"

	DefaultMaxNumberOfRecords = 50
)

var (
	FuncGetRecords    = coreutil.ViewFunc("getRecords")
	FuncGetNumRecords = coreutil.ViewFunc("getNumRecords")
)
