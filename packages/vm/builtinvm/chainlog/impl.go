package log

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("logsc.initialize.begin")
	ctx.Eventf("logsc.initialize.success hname = %s", Interface.Hname().String())
	return nil, nil
}

func storeLog(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Eventf("logsc.storeLog.begin")
	logData, err := ctx.Params().Get(ParamLog)
	if err != nil {
		ctx.Panic(err)
	}
	state := ctx.State()

	log := datatypes.NewMustTimestampedLog(state, VarLogName)

	log.Append(ctx.GetTimestamp(), logData)

	ctx.Eventf("---------------------------")

	return nil, nil
}

func getLogInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	log := datatypes.NewMustTimestampedLog(state, VarLogName)
	ret := dict.New()
	ret.Set(VarLogName, codec.EncodeInt64(int64(log.Len())))

	return ret, nil
}

func getLasts(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	l, ok, err := codec.DecodeInt64(ctx.Params().MustGet(ParamLog))
	if err != nil {
		return nil, err
	}
	if !ok {
		l = 0
	}
	log, err := datatypes.NewTimestampedLog(state, VarLogName)

	if err != nil || log.Len() < uint32(l) {
		return nil, err
	}

	tts, _ := log.TakeTimeSlice(log.Earliest(), log.Latest())
	_, last := tts.FromToIndices()
	total := tts.NumPoints()
	data, erraw := log.LoadRecordsRaw(total-uint32(l), last, false)
	//fmt.Println("RAW DATA: ", data)

	if erraw != nil {
		return nil, err
	}

	ret := dict.New()
	a, err := datatypes.NewArray(ret, VarLogName)
	if err != nil {
		return nil, err
	}
	for _, s := range data {
		a.Push(s)
	}

	return ret, nil
}
