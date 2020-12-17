package chainlog

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is mandatory
func initialize(_ vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// getLenByHnameAndTR gets the number of chainlog records filtered by hname and type
// Parameters:
//	- ParamContractHname Hname of the contract to view the logs
//	- ParamRecordType Type of record that you want to query
func getLenByHnameAndTR(ctx vmtypes.SandboxView) (dict.Dict, error) {
	contractName, recType, err := getFilterParameters(ctx.Params())
	if err != nil {
		return nil, err
	}
	ret := dict.New()
	thelog := datatypes.NewMustTimestampedLog(ctx.State(), chainLogName(contractName, recType))
	ret.Set(ParamNumRecords, codec.EncodeInt64(int64(thelog.Len())))
	return ret, nil
}

// getLogRecords gets logs between timestamp interval and last N number of records
// In time descending order
// Parameters:
//	- ParamContractHname Filter param, Hname of the contract to view the logs
//	- ParamRecordType Filter param, Type of record that you want to query
//  - ParamFromTs From interval. Defaults to 0
//  - ParamToTs To Interval. Defaults to now (if both are missing means all)
//  - ParamMaxLastRecords Max amount of records that you want to return. Defaults to 50
func getLogRecords(ctx vmtypes.SandboxView) (dict.Dict, error) {
	contractHname, recType, err := getFilterParameters(ctx.Params())
	if err != nil {
		return nil, err
	}
	maxLast, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamMaxLastRecords))
	if err != nil {
		return nil, err
	}
	if !ok {
		// taking default
		maxLast = DefaultMaxNumberOfRecords
	}
	fromTs, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamFromTs))
	if err != nil {
		return nil, err
	}
	if !ok {
		fromTs = 0
	}
	toTs, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamToTs))
	if err != nil {
		return nil, err
	}
	if !ok {
		toTs = ctx.GetTimestamp()
	}
	theLog := datatypes.NewMustTimestampedLog(ctx.State(), chainLogName(contractHname, recType))
	tts := theLog.TakeTimeSlice(fromTs, toTs)
	if tts.IsEmpty() {
		// empty time slice
		return nil, nil
	}
	ret := dict.New()
	first, last := tts.FromToIndicesCapped(uint32(maxLast))
	data := theLog.LoadRecordsRaw(first, last, true) // descending
	a := datatypes.NewMustArray(ret, ParamRecords)
	for _, s := range data {
		a.Push(s)
	}
	return ret, nil
}

// getFilterParameters internal utility function to parse and validate params
func getFilterParameters(params dict.Dict) (coretypes.Hname, byte, error) {
	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, fmt.Errorf("parameter 'contractHname' not found")
	}
	typeP, ok, err := codec.DecodeInt64(params.MustGet(ParamRecordType))
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, fmt.Errorf("paremeter 'recordType' not found")
	}
	if !IsSystemTR(byte(typeP)) {
		return 0, 0, fmt.Errorf("parameter 'recordType' is wrong")
	}
	return contractName, byte(typeP), nil
}
