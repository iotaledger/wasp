package builtin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/processor"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
)

type builtinProcessor map[sctransaction.RequestCode]builtinEntryPoint

type builtinEntryPoint func(ctx processor.Sandbox)

var Processor = builtinProcessor{
	vmconst.RequestCodeInit:             initRequest,
	vmconst.RequestCodeNOP:              nopRequest,
	vmconst.RequestCodeSetMinimumReward: setMinimumRewardRequest,
	vmconst.RequestCodeSetDescription:   setDescriptionRequest,
}

func (v *builtinProcessor) GetEntryPoint(code sctransaction.RequestCode) (processor.EntryPoint, bool) {
	if !code.IsReserved() {
		return nil, false
	}
	ep, ok := Processor[code]
	return ep, ok
}

func (ep builtinEntryPoint) Run(ctx processor.Sandbox) {
	ep(ctx)
}

func stub(ctx processor.Sandbox, text string) {
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

func nopRequest(ctx processor.Sandbox) {
	stub(ctx, "nopRequest")
}

// request initializes SC state, must be called in 0 state (usually the origin transaction)
// TODO currently takes into account only owner addr and program hash
func initRequest(ctx processor.Sandbox) {
	stub(ctx, "initRequest")
	if ctx.State().Index() != 0 {
		// call not in the 0 state is ignored
		ctx.Rollback()
		return
	}
	var niladdr address.Address
	ownerAddress, ok := ctx.Request().GetAddressValue(vmconst.VarNameOwnerAddress)
	if !ok || ownerAddress == niladdr {
		// can't proceed if ownerAddress is not known
		ctx.Rollback()
		return
	}
	ctx.State().SetAddressValue(vmconst.VarNameOwnerAddress, ownerAddress)

	progHash, ok := ctx.Request().GetHashValue(vmconst.VarNameProgramHash)
	if !ok || progHash == *hashing.NilHash {
		// program hash not set, smart contract will be able to process only built-in requests
		return
	}
	ctx.State().SetHashValue(vmconst.VarNameProgramHash, progHash)
}

func setMinimumRewardRequest(ctx processor.Sandbox) {
	stub(ctx, "setMinimumRewardRequest")
	if v, ok := ctx.Request().GetInt64("value"); ok && v >= 0 {
		ctx.State().SetInt64(vmconst.VarNameMinimumReward, v)
	}
}

func setDescriptionRequest(ctx processor.Sandbox) {
	stub(ctx, "setDescriptionRequest")
	if v, ok := ctx.Request().GetString("value"); ok && v != "" {
		ctx.State().SetString("description", v)
	}
}
