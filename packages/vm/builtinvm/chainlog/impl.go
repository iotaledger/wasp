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
	ctx.Eventf("chainlog.initialize.begin")
	ctx.Eventf("chainlog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

func storeLog(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("chainlog.storeLog.begin")
	params := ctx.Params()
	state := ctx.State()

	logData, err := params.Get(ParamLog)
	if err != nil {
		return nil, err
	}

	typeP, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamType))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("paremeter 'ParamType' not found")
	}

	contractName := ctx.Caller().MustContractID().Hname()

	switch typeP {
	case TR_DEPLOY, TR_TOKEN_TRANSFER, TR_VIEWCALL, TR_REQUEST_FUNC, TR_GENERIC_DATA:
		entry := append(contractName.Bytes(), byte(typeP))
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))
		log.Append(ctx.GetTimestamp(), logData)
	default:
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}
	return nil, nil
}

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
		entry := append(contractName.Bytes(), byte(typeP))
		log := datatypes.NewMustTimestampedLog(state, kv.Key(entry))

		ret.Set(kv.Key(entry), codec.EncodeInt64(int64(log.Len())))
	default:
		return nil, fmt.Errorf("Type parameter 'ParamType' is incorrect")
	}

	return ret, nil
}

func getLasts(ctx vmtypes.SandboxView) (dict.Dict, error) {

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
