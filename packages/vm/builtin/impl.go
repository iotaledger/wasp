package builtin

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/sctransaction/origin"
	"github.com/iotaledger/wasp/packages/vm"
)

type builtinProcessor map[sctransaction.RequestCode]builtinEntryPoint

type builtinEntryPoint func(ctx vm.Sandbox)

var Processor = builtinProcessor{
	RequestCodeInit:             initRequest,
	RequestCodeNOP:              nopRequest,
	RequestCodeSetMinimumReward: setMinimumRewardRequest,
	RequestCodeSetDescription:   setDescriptionRequest,
}

func (v *builtinProcessor) GetEntryPoint(code sctransaction.RequestCode) (vm.EntryPoint, bool) {
	if !code.IsReserved() {
		return nil, false
	}
	ep, ok := Processor[code]
	return ep, ok
}

func (ep builtinEntryPoint) Run(ctx vm.Sandbox) {
	ep(ctx)
}

func stub(ctx vm.Sandbox, text string) {
	reqId := ctx.Request().ID()
	ctx.GetLog().Debugw("run builtInProcessor",
		"text", text,
		"request code", ctx.Request().Code(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"state index", ctx.State().Index(),
		"req", reqId.String(),
	)
}

func nopRequest(ctx vm.Sandbox) {
	stub(ctx, "nopRequest")
}

// request initializes SC state, must be called in 0 state (usually the origin transaction)
func initRequest(ctx vm.Sandbox) {
	stub(ctx, "initRequest")
	if ctx.State().Index() != 0 {
		// call not in the 0 state is ignored
		ctx.Rollback()
		return
	}
	ownerAddress, ok := ctx.Request().GetString(origin.VarNameOwnerAddress)
	if !ok || ownerAddress == "" {
		ctx.Rollback()
		return
	}
	ctx.State().SetString(origin.VarNameOwnerAddress, ownerAddress)

	progHashStr, ok := ctx.Request().GetString(origin.VarNameProgramHash)
	if !ok || progHashStr == "" {
		// program hash not set, smart contract will be able to process only built in requests
		return
	}
	ctx.State().SetString(origin.VarNameProgramHash, progHashStr)
}

func setMinimumRewardRequest(ctx vm.Sandbox) {
	stub(ctx, "setMinimumRewardRequest")
	if v, ok := ctx.Request().GetInt64("value"); ok && v >= 0 {
		ctx.State().SetInt64(origin.VarNameMinimumReward, v)
	}
}

func setDescriptionRequest(ctx vm.Sandbox) {
	stub(ctx, "setDescriptionRequest")
	if v, ok := ctx.Request().GetString("value"); ok && v != "" {
		ctx.State().SetString("description", v)
	}
}
