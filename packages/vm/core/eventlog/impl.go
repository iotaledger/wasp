package eventlog

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// initialize is mandatory
func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("eventlog.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

// getNumRecords gets the number of eventlog records for contarct
// Parameters:
//	- ParamContractHname Hname of the contract to view the logs
func getNumRecords(ctx vmtypes.SandboxView) (dict.Dict, error) {
	contractName, err := getHnameParameter(ctx.Params())
	if err != nil {
		return nil, err
	}
	ret := dict.New()
	thelog := datatypes.NewMustTimestampedLog(ctx.State(), kv.Key(contractName.Bytes()))
	ret.Set(ParamNumRecords, codec.EncodeInt64(int64(thelog.Len())))
	return ret, nil
}

// getLogRecords returns records between timestamp interval for the hname
// In time descending order
// Parameters:
//	- ParamContractHname Filter param, Hname of the contract to view the logs
//  - ParamFromTs From interval. Defaults to 0
//  - ParamToTs To Interval. Defaults to now (if both are missing means all)
//  - ParamMaxLastRecords Max amount of records that you want to return. Defaults to 50
func getLogRecords(ctx vmtypes.SandboxView) (dict.Dict, error) {
	contractHname, err := getHnameParameter(ctx.Params())
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
	theLog := datatypes.NewMustTimestampedLog(ctx.State(), kv.Key(contractHname.Bytes()))
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

// getHnameParameter internal utility function to parse and validate params
func getHnameParameter(params dict.Dict) (coretypes.Hname, error) {
	contractName, ok, err := codec.DecodeHname(params.MustGet(ParamContractHname))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("parameter 'contractHname' not found")
	}
	return contractName, nil
}
