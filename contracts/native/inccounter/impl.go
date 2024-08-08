package inccounter

import (
	"fmt"
	"math"
	"time"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

var Contract = coreutil.NewContract("inccounter")

var (
	FuncIncCounter = coreutil.NewEP1(Contract, "incCounter",
		coreutil.FieldWithCodecOptional(codec.Int64),
	)
	FuncIncAndRepeatOnceAfter2s = coreutil.NewEP0(Contract, "incAndRepeatOnceAfter5s")
	FuncIncAndRepeatMany        = coreutil.NewEP2(Contract, "incAndRepeatMany",
		coreutil.FieldWithCodecOptional(codec.Int64),
		coreutil.FieldWithCodecOptional(codec.Int64),
	)
	FuncSpawn = coreutil.NewEP1(Contract, "spawn",
		coreutil.FieldWithCodec(codec.String),
	)
	ViewGetCounter = coreutil.NewViewEP01(Contract, "getCounter",
		coreutil.FieldWithCodec(codec.Int64),
	)
)

var Processor = Contract.Processor(initialize,
	FuncIncCounter.WithHandler(incCounter),
	FuncIncAndRepeatOnceAfter2s.WithHandler(incCounterAndRepeatOnce),
	FuncIncAndRepeatMany.WithHandler(incCounterAndRepeatMany),
	FuncSpawn.WithHandler(spawn),
	ViewGetCounter.WithHandler(getCounter),
)

func InitParams(initialValue int64) dict.Dict {
	return dict.Dict{VarCounter: codec.Int64.Encode(initialValue)}
}

const (
	VarNumRepeats = "numRepeats"
	VarCounter    = "counter"
	VarName       = "name"
)

func initialize(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Debugf("inccounter.init in %s", ctx.Contract().String())
	params := ctx.Params()
	val := codec.Int64.MustDecode(params.Get(VarCounter), 0)
	ctx.State().Set(VarCounter, codec.Int64.Encode(val))
	eventCounter(ctx, val)

	return isc.CallArguments{}
}

func incCounter(ctx isc.Sandbox, incOpt *int64) {
	inc := coreutil.FromOptional(incOpt, 1)
	ctx.Log().Debugf("inccounter.incCounter in %s", ctx.Contract().String())

	state := kvdecoder.New(ctx.State(), ctx.Log())
	val := state.MustGetInt64(VarCounter, 0)
	ctx.Log().Infof("incCounter: increasing counter value %d by %d, anchor index: #%d",
		val, inc, ctx.StateAnchor().StateIndex)
	tra := "(empty)"
	if ctx.AllowanceAvailable() != nil {
		tra = ctx.AllowanceAvailable().String()
	}
	ctx.Log().Infof("incCounter: allowance available: %s", tra)
	ctx.State().Set(VarCounter, codec.Int64.Encode(val+inc))
	eventCounter(ctx, val+inc)
}

func incCounterAndRepeatOnce(ctx isc.Sandbox) {
	ctx.Log().Debugf("inccounter.incCounterAndRepeatOnce")
	state := ctx.State()
	val := codec.Int64.MustDecode(state.Get(VarCounter), 0)

	ctx.Log().Debugf(fmt.Sprintf("incCounterAndRepeatOnce: increasing counter value: %d", val))
	state.Set(VarCounter, codec.Int64.Encode(val+1))
	eventCounter(ctx, val+1)
	allowance := ctx.AllowanceAvailable()
	ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Send(isc.RequestParameters{
		TargetAddress:                 ctx.ChainID().AsAddress(),
		Assets:                        isc.NewAssets(allowance.BaseTokens()),
		AdjustToMinimumStorageDeposit: true,
		Metadata: &isc.SendMetadata{
			Message:   isc.NewMessage(ctx.Contract(), FuncIncCounter.Hname()),
			GasBudget: math.MaxUint64,
		},
		Options: isc.SendOptions{
			Timelock: ctx.Timestamp().Add(2 * time.Second),
		},
	})
	ctx.Log().Debugf("incCounterAndRepeatOnce: PostRequestToSelfWithDelay RequestInc 2 sec")
}

func incCounterAndRepeatMany(ctx isc.Sandbox, valOpt, numRepeatsOpt *int64) {
	val := coreutil.FromOptional(valOpt, 0)
	numRepeats := coreutil.FromOptional(valOpt, lo.Must(codec.Int64.Decode(ctx.State().Get(VarNumRepeats), 0)))
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany")

	state := ctx.State()

	state.Set(VarCounter, codec.Int64.Encode(val+1))
	eventCounter(ctx, val+1)
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: increasing counter value: %d", val)

	if numRepeats == 0 {
		ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return
	}

	ctx.Log().Debugf("chain of %d requests ahead", numRepeats)

	state.Set(VarNumRepeats, codec.Int64.Encode(numRepeats-1))
	ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Send(isc.RequestParameters{
		TargetAddress:                 ctx.ChainID().AsAddress(),
		Assets:                        isc.NewAssets(1000),
		AdjustToMinimumStorageDeposit: true,
		Metadata: &isc.SendMetadata{
			Message:   isc.NewMessage(ctx.Contract(), FuncIncAndRepeatMany.Hname()),
			GasBudget: math.MaxUint64,
			Allowance: isc.NewAssets(1000),
		},
		Options: isc.SendOptions{
			Timelock: ctx.Timestamp().Add(2 * time.Second),
		},
	})

	ctx.Log().Debugf("incCounterAndRepeatMany. remaining repeats = %d", numRepeats-1)
}

// spawn deploys new contract and calls it
func spawn(ctx isc.Sandbox, name string) {
	ctx.Log().Debugf("inccounter.spawn")

	state := kvdecoder.New(ctx.State(), ctx.Log())
	val := state.MustGetInt64(VarCounter)

	callPar := dict.New()
	callPar.Set(VarCounter, codec.Int64.Encode(val+1))
	eventCounter(ctx, val+1)
	ctx.DeployContract(Contract.ProgramHash, name, callPar)

	// increase counter in newly spawned contract
	ctx.Call(FuncIncCounter.Message(nil), nil)
}

func getCounter(ctx isc.SandboxView) int64 {
	return lo.Must(codec.Int64.Decode(ctx.StateR().Get(VarCounter), 0))
}
