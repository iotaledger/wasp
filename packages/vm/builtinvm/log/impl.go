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

	return nil, nil
}

func getLogInfo(ctx vmtypes.SandboxView) (dict.Dict, error) {

	state := ctx.State()
	log := datatypes.NewMustTimestampedLog(state, VarLogName)
	ret := dict.New()
	ret.Set(VarLogName, codec.EncodeInt64(int64(log.Len())))

	return ret, nil
}
