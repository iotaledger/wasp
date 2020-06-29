package builtin

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type builtinProcessor map[sctransaction.RequestCode]builtinEntryPoint

type builtinEntryPoint func(ctx vmtypes.Sandbox)

var Processor = builtinProcessor{
	vmconst.RequestCodeInit:             initRequest,
	vmconst.RequestCodeNOP:              nopRequest,
	vmconst.RequestCodeSetMinimumReward: setMinimumRewardRequest,
	vmconst.RequestCodeSetDescription:   setDescriptionRequest,
}

func (v *builtinProcessor) GetEntryPoint(code sctransaction.RequestCode) (vmtypes.EntryPoint, bool) {
	if !code.IsReserved() {
		return nil, false
	}
	ep, ok := Processor[code]
	return ep, ok
}

func (ep builtinEntryPoint) Run(ctx vmtypes.Sandbox) {
	ep(ctx)
}

func (v builtinEntryPoint) WithGasLimit(_ int) vmtypes.EntryPoint {
	return v
}

func stub(ctx vmtypes.Sandbox, text string) {
	reqId := ctx.AccessRequest().ID()
	ctx.GetLog().Debugw("run builtInProcessor",
		"text", text,
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}

func nopRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "nopRequest")
}

// request initializes SC state, must be called in 0 state (usually the origin transaction)
// TODO currently takes into account only owner addr and program hash
func initRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "initRequest")
	if !ctx.IsOriginState() {
		// call not in the 0 state is ignored
		ctx.GetLog().Debugf("@@@@@@@@@@@ exit 1")
		return
	}
	var niladdr address.Address
	ownerAddress, ok := ctx.AccessRequest().GetAddressValue(vmconst.VarNameOwnerAddress)
	if !ok || ownerAddress == niladdr {
		// can't proceed if ownerAddress is not known
		ctx.GetLog().Debugf("@@@@@@@@@@@ exit 2")
		return
	}
	ctx.AccessState().SetAddressValue(vmconst.VarNameOwnerAddress, ownerAddress)

	progHash, ok := ctx.AccessRequest().GetHashValue(vmconst.VarNameProgramHash)
	if !ok || progHash == *hashing.NilHash {
		// program hash not set, smart contract will be able to process only built-in requests
		ctx.GetLog().Debugf("@@@@@@@@@@@ exit 3")
		return
	}
	ctx.AccessState().SetHashValue(vmconst.VarNameProgramHash, &progHash)
}

func setMinimumRewardRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "setMinimumRewardRequest")
	if v, ok := ctx.AccessRequest().GetInt64("value"); ok && v >= 0 {
		ctx.AccessState().SetInt64(vmconst.VarNameMinimumReward, v)
	}
}

func setDescriptionRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "setDescriptionRequest")
	if v, ok := ctx.AccessRequest().GetString("value"); ok && v != "" {
		ctx.AccessState().SetString("description", v)
	}
}
