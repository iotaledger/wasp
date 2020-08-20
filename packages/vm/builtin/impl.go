package builtin

import (
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/vmconst"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type builtinProcessor map[sctransaction.RequestCode]builtinEntryPoint

type builtinEntryPoint func(ctx vmtypes.Sandbox)

var Processor = builtinProcessor{
	vmconst.RequestCodeNOP:              nopRequest,
	vmconst.RequestCodeInit:             initRequest,
	vmconst.RequestCodeSetMinimumReward: setMinimumReward,
	vmconst.RequestCodeSetDescription:   setDescription,
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
	ctx.GetWaspLog().Debugw("run builtInProcessor",
		"text", text,
		"request code", ctx.AccessRequest().Code(),
		"addr", ctx.GetSCAddress().String(),
		"ts", ctx.GetTimestamp(),
		"req", reqId.String(),
	)
}

func nopRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "nopRequest")
}

// request initializes SC state, must be called in 0 state (the origin transaction)
// TODO currently takes into account only owner addr and program hash
func initRequest(ctx vmtypes.Sandbox) {
	stub(ctx, "initRequest")
	if !ctx.IsOriginState() {
		// call not in the 0 state is ignored
		ctx.GetWaspLog().Debugf("initRequest: not in origin state.")
		return
	}
	ownerAddress, ok, err := ctx.AccessRequest().Args().GetAddress(vmconst.VarNameOwnerAddress)
	if err != nil {
		ctx.GetWaspLog().Errorf("initRequest: Could not read request argument: %s", err.Error())
		return
	}
	if !ok {
		ctx.GetWaspLog().Debugf("initRequest: owner address not known.")
		return
	}
	ctx.AccessState().SetAddress(vmconst.VarNameOwnerAddress, ownerAddress)

	progHash, ok, err := ctx.AccessRequest().Args().GetHashValue(vmconst.VarNameProgramHash)
	if err != nil {
		ctx.GetWaspLog().Errorf("initRequest: Could not read request argument: %s", err.Error())
		return
	}
	if !ok {
		ctx.GetWaspLog().Debugf("initRequest: program hash not set; smart contract will be able to process only built-in requests.")
		return
	}
	ctx.GetWaspLog().Debugf("initRequest: Setting program hash to %s.", progHash.String())
	ctx.AccessState().SetHashValue(vmconst.VarNameProgramHash, progHash)
}

func setMinimumReward(ctx vmtypes.Sandbox) {
	stub(ctx, "setMinimumReward")
	if v, ok, _ := ctx.AccessRequest().Args().GetInt64("value"); ok && v >= 0 {
		ctx.AccessState().SetInt64(vmconst.VarNameMinimumReward, v)
	}
}

func setDescription(ctx vmtypes.Sandbox) {
	stub(ctx, "setDescription")
	if v, ok, _ := ctx.AccessRequest().Args().GetString("value"); ok && v != "" {
		ctx.AccessState().SetString("description", v)
	}
}
