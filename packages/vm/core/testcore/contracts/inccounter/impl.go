// Package inccounter contains counter testing logic
package inccounter

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kv/dict"
)

var Processor = Contract.Processor(nil,
	FuncIncCounter.WithHandler(incCounter),
	ViewGetCounter.WithHandler(getCounter),
)

func InitParams(initialValue int64) dict.Dict {
	return dict.Dict{VarCounter: codec.Encode(initialValue)}
}

const (
	VarNumRepeats = "numRepeats"
	VarCounter    = "counter"
	VarName       = "name"
)

func SetInitialState(contractPartition kv.KVStore) {
	contractPartition.Set(VarCounter, codec.Encode[int64](0))
}

func incCounter(ctx isc.Sandbox, incOpt *int64) {
	inc := coreutil.FromOptional(incOpt, 1)
	ctx.Log().Debugf("inccounter.incCounter in %s", ctx.Contract().String())

	val := codec.MustDecode[int64](ctx.State().Get(VarCounter))
	ctx.Log().Infof("incCounter: increasing counter value %d by %d", val, inc)
	tra := "(empty)"
	if ctx.AllowanceAvailable() != nil {
		tra = ctx.AllowanceAvailable().String()
	}
	ctx.Log().Infof("incCounter: allowance available: %s", tra)
	ctx.State().Set(VarCounter, codec.Encode(val+inc))
}

func getCounter(ctx isc.SandboxView) int64 {
	return lo.Must(codec.Decode[int64](ctx.StateR().Get(VarCounter), 0))
}
