package chainlog

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/util"
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
		contract.ViewFunc(FuncLenByHnameAndTR, getLenByHnameAndTR),
	})
}

const (
	// request parameters
	ParamContractHname  = "contractHname"
	ParamFromTs         = "fromTs"
	ParamToTs           = "toTs"
	ParamMaxLastRecords = "maxLastRecords"
	ParamRecordType     = "paramTypeOfRecords"
	ParamNumRecords     = "numRecords"
	ParamRecords        = "records"

	// function names
	FuncGetLogRecords   = "getLogRecords"
	FuncLenByHnameAndTR = "getLenByHnameAndTR"

	// Type of records
	// Constants that define the system-interpreted type of logged data. Different types are defined:
	// -TRDeploy -> contract deployment
	// -TREvent -> sc event
	// -TRRequest -> sc request
	// -TRGenericData -> user defined log record
	TRDeploy      = 1
	TREvent       = 2
	TRRequest     = 3
	TRGenericData = 4

	DefaultMaxNumberOfRecords = 50
)

func GetProcessor() vmtypes.Processor {
	return Interface
}

func IsSystemTR(t byte) bool {
	switch t {
	case TRDeploy, TREvent, TRRequest, TRGenericData:
		return true
	default:
		return false
	}
}

type RequestChainLogRecord struct {
	RequestID  coretypes.RequestID
	EntryPoint coretypes.Hname
	Error      string
}

const MaxRequestError = 100

// EncodeRequestChainLogRecord truncates too long error and encodes
func EncodeRequestChainLogRecord(rec *RequestChainLogRecord) []byte {
	var buf bytes.Buffer
	buf.Write(rec.RequestID[:])
	buf.Write(rec.EntryPoint.Bytes())
	s := rec.Error
	if len(rec.Error) > MaxRequestError {
		s = s[:MaxRequestError]
	}
	_ = util.WriteString16(&buf, s)
	return buf.Bytes()
}

func DecodeRequestChainLogRecord(data []byte) (*RequestChainLogRecord, error) {
	ret := new(RequestChainLogRecord)
	rdr := bytes.NewReader(data)
	if err := ret.RequestID.Read(rdr); err != nil {
		return nil, err
	}
	if err := ret.EntryPoint.Read(rdr); err != nil {
		return nil, err
	}
	var err error
	if ret.Error, err = util.ReadString16(rdr); err != nil {
		return nil, err
	}
	return ret, nil
}
