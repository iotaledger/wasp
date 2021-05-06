package eventlog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// initialize is mandatory
func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("eventlog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

// getNumRecords gets the number of eventlog records for contract
// Parameters:
//	- ParamContractHname Hname of the contract to view the logs
func getNumRecords(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	contractHname, err := params.GetHname(ParamContractHname)
	if err != nil {
		return nil, err
	}
	ret := dict.New()
	thelog := collections.NewTimestampedLogReadOnly(ctx.State(), kv.Key(contractHname.Bytes()))
	ret.Set(ParamNumRecords, codec.EncodeInt64(int64(thelog.MustLen())))
	return ret, nil
}

// getRecords returns records between timestamp interval for the hname
// In time descending order
// Parameters:
//	- ParamContractHname Filter param, Hname of the contract to view the logs
//  - ParamFromTs From timestamp. Defaults to 0
//  - ParamToTs To timestamp. Defaults to now (if both are missing means all)
//  - ParamMaxLastRecords Max amount of records that you want to return. Default is 50
func getRecords(ctx coretypes.SandboxView) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())

	contractHname, err := params.GetHname(ParamContractHname)
	if err != nil {
		return nil, err
	}
	maxLast, err := params.GetInt64(ParamMaxLastRecords, DefaultMaxNumberOfRecords)
	if err != nil {
		return nil, err
	}
	fromTs, err := params.GetInt64(ParamFromTs, 0)
	if err != nil {
		return nil, err
	}
	toTs, err := params.GetInt64(ParamToTs, ctx.GetTimestamp())
	if err != nil {
		return nil, err
	}

	theLog := collections.NewTimestampedLogReadOnly(ctx.State(), kv.Key(contractHname.Bytes()))
	tts := theLog.MustTakeTimeSlice(fromTs, toTs)
	if tts.IsEmpty() {
		// empty time slice
		return nil, nil
	}
	ret := dict.New()
	first, last := tts.FromToIndicesCapped(uint32(maxLast))
	data := theLog.MustLoadRecordsRaw(first, last, true) // descending
	a := collections.NewArray16(ret, ParamRecords)
	for _, s := range data {
		a.MustPush(s)
	}
	return ret, nil
}
