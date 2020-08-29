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
	vmconst.RequestCodeInit:             initSC,
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

func nopRequest(ctx vmtypes.Sandbox) {
	ctx.Publish("nopRequest")
}

// request initializes SC state, must be called in 0 state (the origin transaction)
func initSC(ctx vmtypes.Sandbox) {
	ctx.Publishf("initSC")
	if !ctx.IsOriginState() {
		// call not in the 0 state is ignored
		ctx.Publish("initSC: error: not in origin state.")
		return
	}
	// set owner address
	ownerAddress, ok, err := ctx.AccessRequest().Args().GetAddress(vmconst.VarNameOwnerAddress)
	if err != nil {
		ctx.Publishf("initSC: Could not read request argument: %s", err.Error())
		return
	}
	if !ok {
		ctx.Publishf("initSC: owner address not known.")
		return
	}
	ctx.AccessState().SetAddress(vmconst.VarNameOwnerAddress, ownerAddress)

	// set program hash
	progHash, ok, err := ctx.AccessRequest().Args().GetHashValue(vmconst.VarNameProgramHash)
	if err != nil {
		ctx.Publishf("init_sc error Could not read program hash from the request: %s", err.Error())
		return
	}
	if !ok {
		ctx.Publishf("init_sc warn program hash not set; smart contract will be able to process only built-in requests.")
		return
	}
	ctx.AccessState().SetHashValue(vmconst.VarNameProgramHash, progHash)
	ctx.Publishf("init_sc info program hash set to %s.", progHash.String())

	// set description
	dscr, ok, err := ctx.AccessRequest().Args().GetString(vmconst.VarNameDescription)
	if err != nil {
		ctx.Publishf("init_sc error Could not read description from the request: %s", err.Error())
		return
	}
	if !ok {
		ctx.Publishf("init_sc warn description not set")
		return
	}
	ctx.AccessState().SetString(vmconst.VarNameDescription, dscr)
	ctx.Publishf("init_sc info description set to '%s'", dscr)

	ctx.Publishf("init_sc success %s", ctx.GetSCAddress().String())
}

func setMinimumReward(ctx vmtypes.Sandbox) {
	ctx.Publish("setMinimumReward")
	if v, ok, _ := ctx.AccessRequest().Args().GetInt64("value"); ok && v >= 0 {
		ctx.AccessState().SetInt64(vmconst.VarNameMinimumReward, v)
	}
}

func setDescription(ctx vmtypes.Sandbox) {
	ctx.Publish("setDescription")
	if v, ok, _ := ctx.AccessRequest().Args().GetString("value"); ok && v != "" {
		ctx.AccessState().SetString("description", v)
	}
}
