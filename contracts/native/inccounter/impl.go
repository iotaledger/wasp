package inccounter

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/root"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const (
	Name        = "inccounter"
	description = "Increment counter, a PoC smart contract"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncIncCounter, incCounter),
		coreutil.Func(FuncIncAndRepeatOnceAfter5s, incCounterAndRepeatOnce),
		coreutil.Func(FuncIncAndRepeatMany, incCounterAndRepeatMany),
		coreutil.Func(FuncSpawn, spawn),
		coreutil.ViewFunc(FuncGetCounter, getCounter),
	})
}

const (
	FuncIncCounter              = "incCounter"
	FuncIncAndRepeatOnceAfter5s = "incAndRepeatOnceAfter5s"
	FuncIncAndRepeatMany        = "incAndRepeatMany"
	FuncSpawn                   = "spawn"
	FuncGetCounter              = "getCounter"
)

const (
	VarNumRepeats  = "numRepeats"
	VarCounter     = "counter"
	VarName        = "name"
	VarDescription = "dscr"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.init in %s", ctx.Contract().String())
	params := ctx.Params()
	val, _, err := codec.DecodeInt64(params.MustGet(VarCounter))
	if err != nil {
		return nil, fmt.Errorf("incCounter: %v", err)
	}
	ctx.State().Set(VarCounter, codec.EncodeInt64(val))
	ctx.Event(fmt.Sprintf("inccounter.init.success. counter = %d", val))
	return nil, nil
}

func incCounter(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounter in %s", ctx.Contract().String())
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	inc := par.MustGetInt64(VarCounter, 1)

	state := kvdecoder.New(ctx.State(), ctx.Log())
	val := state.MustGetInt64(VarCounter, 0)
	ctx.Log().Infof("incCounter: increasing counter value %d by %d, anchor index: #%d",
		val, inc, ctx.StateAnchor().StateIndex())
	tra := "(empty)"
	if ctx.IncomingTransfer() != nil {
		tra = ctx.IncomingTransfer().String()
	}
	ctx.Log().Infof("incCounter: incoming transfer: %s", tra)
	ctx.State().Set(VarCounter, codec.EncodeInt64(val+inc))
	return nil, nil
}

func incCounterAndRepeatOnce(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounterAndRepeatOnce")
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))

	ctx.Log().Debugf(fmt.Sprintf("incCounterAndRepeatOnce: increasing counter value: %d", val))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	if !ctx.Send(ctx.ChainID().AsAddress(), coretypes.NewTransferIotas(1), &coretypes.SendMetadata{
		TargetContract: ctx.Contract(),
		EntryPoint:     coretypes.Hn(FuncIncCounter),
	}, coretypes.SendOptions{
		TimeLock: 5 * 60,
	}) {
		return nil, fmt.Errorf("incCounterAndRepeatOnce: not enough funds")
	}
	ctx.Log().Debugf("incCounterAndRepeatOnce: PostRequestToSelfWithDelay RequestInc 5 sec")
	return nil, nil
}

func incCounterAndRepeatMany(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany")

	state := ctx.State()
	params := ctx.Params()

	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	state.Set(VarCounter, codec.EncodeInt64(val+1))
	ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: increasing counter value: %d", val)

	numRepeats, ok, err := codec.DecodeInt64(params.MustGet(VarNumRepeats))
	if err != nil {
		ctx.Log().Panicf("%s", err)
	}
	if !ok {
		numRepeats, _, _ = codec.DecodeInt64(state.MustGet(VarNumRepeats))
	}
	if numRepeats == 0 {
		ctx.Log().Debugf("inccounter.incCounterAndRepeatMany: finished chain of requests. counter value: %d", val)
		return nil, nil
	}

	ctx.Log().Debugf("chain of %d requests ahead", numRepeats)

	state.Set(VarNumRepeats, codec.EncodeInt64(numRepeats-1))

	if !ctx.Send(ctx.ChainID().AsAddress(), coretypes.NewTransferIotas(1), &coretypes.SendMetadata{
		TargetContract: ctx.Contract(),
		EntryPoint:     coretypes.Hn(FuncIncAndRepeatMany),
	}, coretypes.SendOptions{
		TimeLock: 1 * 60,
	}) {
		ctx.Log().Debugf("incCounterAndRepeatMany. remaining repeats = %d", numRepeats-1)
	} else {
		ctx.Log().Debugf("incCounterAndRepeatMany FAILED. remaining repeats = %d", numRepeats-1)
	}
	return nil, nil
}

// spawn deploys new contract and calls it
func spawn(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("inccounter.spawn")

	state := kvdecoder.New(ctx.State(), ctx.Log())
	val := state.MustGetInt64(VarCounter)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	name := par.MustGetString(VarName)
	dscr := par.MustGetString(VarDescription, "N/A")

	a := assert.NewAssert(ctx.Log())

	callPar := dict.New()
	callPar.Set(VarCounter, codec.EncodeInt64(val+1))
	err := ctx.DeployContract(Interface.ProgramHash, name, dscr, callPar)
	a.RequireNoError(err)

	// increase counter in newly spawned contract
	hname := coretypes.Hn(name)
	_, err = ctx.Call(hname, coretypes.Hn(FuncIncCounter), nil, nil)
	a.RequireNoError(err)

	res, err := ctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncGetChainInfo), nil, nil)
	a.RequireNoError(err)

	creg := collections.NewMapReadOnly(res, root.VarContractRegistry)
	a.Require(int(creg.MustLen()) == len(core.AllCoreContractsByHash)+2, "unexpected contract registry len %d", creg.MustLen())
	ctx.Log().Debugf("inccounter.spawn: new contract name = %s hname = %s", name, hname.String())
	return nil, nil
}

func getCounter(ctx coretypes.SandboxView) (dict.Dict, error) {
	state := ctx.State()
	val, _, _ := codec.DecodeInt64(state.MustGet(VarCounter))
	ret := dict.New()
	ret.Set(VarCounter, codec.EncodeInt64(val))
	return ret, nil
}
