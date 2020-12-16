package chainlog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

//
// Gets the length by Hname of the contract and the type of record of the log
//
// Parameters:
//	- ParamContractHname Hname of the contract to view the logs
//	- ParamRecordType Type of record that you want to query
func getLenByHnameAndTR(ctx vmtypes.SandboxView) (dict.Dict, error) {
	state := ctx.State()
	params := ctx.Params()

	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'contract hname' not found")
	}

	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamRecordType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamRecordType' not found")
	}

	ret := dict.New()
	switch byte(typeP) {
	case TRDeploy, TRViewCall, TRRequest, TRGenericData:
		log, entry := GetOrCreateDataLog(state, contractName, byte(typeP))
		ret.Set(kv.Key(entry.String()), codec.EncodeInt64(int64(log.Len())))
	default:
		return nil, fmt.Errorf("Type parameter 'ParamRecordType' is incorrect")
	}

	return ret, nil
}

//
// Gets the last N records by contractHname and record type
//
// Parameters:
//	- ParamContractHname Hname of the contract to view the logs
//	- ParamRecordType Type of record that you want to query
func getLastsNRecords(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	params := ctx.Params()

	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'contract hname' not found")
	}

	l, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamLastsRecords))
	if err != nil {
		return nil, err
	}
	if !ok {
		l = 0
	}

	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamRecordType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamRecordType' not found")
	}

	ret := dict.New()
	switch byte(typeP) {
	case TRDeploy, TRViewCall, TRRequest, TRGenericData:
		entry := append(contractName.Bytes(), byte(typeP))
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))

		if err != nil || log.Len() < uint32(l) {
			return nil, err
		}

		tts := log.TakeTimeSlice(log.Earliest(), log.Latest())
		if tts.IsEmpty() {
			// empty time slice
			return nil, nil
		}
		first, last := tts.FromToIndices()
		from := first
		nPoints := tts.NumPoints()
		if l != 0 && nPoints > uint32(l) {
			from = nPoints - uint32(l)
		}
		data := log.LoadRecordsRaw(from, last, false)

		a := datatypes.NewMustArray(ret, string(entry))
		for _, s := range data {
			a.Push(s)
		}

	default:
		return nil, fmt.Errorf("Type parameter 'ParamRecordType' is incorrect")
	}

	return ret, nil
}

// Gets logs between timestamp interval and last N number of records
//
// Parameters:
//  - ParamFromTs From interval
//  - ParamToTs To Interval
//  - ParamLastsRecords Amount of records that you want to return
//	- ParamContractHname Hname of the contract to view the logs
//	- ParamRecordType Type of record that you want to query
func getLogsBetweenTs(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	params := ctx.Params()

	fromTs, ok, err := codec.DecodeInt64(params.MustGet(ParamFromTs))
	if err != nil {
		return nil, err
	}
	if !ok {
		fromTs = 0
	}
	toTs, ok, err := codec.DecodeInt64(params.MustGet(ParamToTs))

	if err != nil {
		return nil, err
	}
	if !ok {
		toTs = ctx.GetTimestamp()
	}

	lastRecords, ok, err := codec.DecodeInt64(params.MustGet(ParamLastsRecords))

	if err != nil {
		return nil, err
	}
	if !ok {
		lastRecords = 0 // 0 means all
	}

	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'contract hname' not found")
	}

	typeP, ok, err := codec.DecodeInt64(params.MustGet(ParamRecordType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamRecordType' not found")
	}

	ret := dict.New()
	switch byte(typeP) {
	case TRDeploy, TRViewCall, TRRequest, TRGenericData:
		log, entry := GetOrCreateDataLog(state, contractName, byte(typeP))
		tts := log.TakeTimeSlice(fromTs, toTs)
		if tts.IsEmpty() {
			// empty time slice
			return nil, nil
		}
		first, last := tts.FromToIndices()
		from := first
		nPoints := tts.NumPoints()
		if lastRecords != 0 && nPoints > uint32(lastRecords) {
			from = nPoints - uint32(lastRecords)
		}
		data := log.LoadRecordsRaw(from, last, false)
		a := datatypes.NewMustArray(ret, entry.String())
		for _, s := range data {
			a.Push(s)
		}
	default:
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}

	return ret, nil
}
