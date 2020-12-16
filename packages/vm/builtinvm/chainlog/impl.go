package log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	// this funtion can be empty
	ctx.Eventf("chainlog.initialize.begin")
	ctx.Eventf("chainlog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

//+++ describe the entry point, with parameters etc
func storeLog(ctx vmtypes.Sandbox) (dict.Dict, error) {
	// don't use Eventf. Instead use .Log().Debugf() or .Log().Infof()
	ctx.Eventf("chainlog.storeLog.begin")
	params := ctx.Params()
	state := ctx.State()

	logData, err := params.Get(ParamLog)
	if err != nil {
		return nil, err
	}
	//+++ logData can be missing, i.e. == nil. Is it OK?
	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamType' not found")
	}

	contractName := ctx.Caller().MustContractID().Hname()
	switch typeP {
	//+++ better define range for system record codes and check the range
	case TR_DEPLOY, TR_TOKEN_TRANSFER, TR_VIEWCALL, TR_REQUEST_FUNC, TR_GENERIC_DATA:
		entry := append(contractName.Bytes(), byte(typeP))
		//+++ better not use name 'log'. Very genetric. use 'tlog' or similar
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))
		log.Append(ctx.GetTimestamp(), logData)
	default:
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}
	return nil, nil
}

//+++ describe the entry point, with parameters etc
//+++ do not understand what it does and why do we need hname as a parameter
func getLogInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	params := ctx.Params()

	//TODO: check if the contract really exists in the chain
	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'contract hname' not found")
	}

	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamType' not found")
	}

	ret := dict.New()
	switch typeP {
	case TR_DEPLOY, TR_TOKEN_TRANSFER, TR_VIEWCALL, TR_REQUEST_FUNC, TR_GENERIC_DATA:
		//+++ 1. type is uint16, not byte
		//+++ 2. why the entry is used as a name of the tlog?
		entry := append(contractName.Bytes(), byte(typeP))
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))

		ret.Set(kv.Key(entry), codec.EncodeInt64(int64(log.Len())))
	default:
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}

	return ret, nil
}

// +++ must be described. The function should be like this:
// get last N records with filter. Filter can be:
//  - specific contract or all
//  - specific data type or all
// The log must be organized like this:
// each record has structure: 4 bytes of contract's hname, 2 bytes of record type code, 0 or any number of data bytes
// Each record type interpreted by core contracts will have specific interpretation of its data
// the user defined types will be interpreted bu the core as bte arrays
func getLasts(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	params := ctx.Params()

	//TODO: check if the contract really exists in the chain
	//+++ why we need to check it?
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

	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamType' not found")
	}

	ret := dict.New()
	switch typeP {
	case TR_DEPLOY, TR_TOKEN_TRANSFER, TR_VIEWCALL, TR_REQUEST_FUNC, TR_GENERIC_DATA:
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
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}

	return ret, nil
}

// Gets logs between timestamp interval and last N number of records
//
// Parameters:
//  - ParamFromTs From interval
//  - ParamToTs To Interval
//  - ParamLastsRecords Amount of records that you want to return
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

	l, ok, err := codec.DecodeInt64(params.MustGet(ParamLastsRecords))

	if err != nil {
		return nil, err
	}
	if !ok {
		l = 0 // 0 means all
	}

	//TODO: check if the contract really exists in the chain
	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'contract hname' not found")
	}

	typeP, ok, err := codec.DecodeInt64(params.MustGet(ParamType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamType' not found")
	}

	ret := dict.New()
	switch typeP {
	case TR_DEPLOY, TR_TOKEN_TRANSFER, TR_VIEWCALL, TR_REQUEST_FUNC, TR_GENERIC_DATA:
		entry := append(contractName.Bytes(), byte(typeP))
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))

		tts := log.TakeTimeSlice(fromTs, toTs)
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
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}

	return ret, nil
}
