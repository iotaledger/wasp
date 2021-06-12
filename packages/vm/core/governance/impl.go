package governance

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// checkRotateStateControllerRequest the entry point is called when committee is about to be rotated to the new address
// If it fails, nothing happens and the state has trace of the failure in the state
// If it is successful VM takes over and replaces resulting transaction with
// governance transition. The state of the chain remains unchanged
func checkRotateStateControllerRequest(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "checkRotateStateControllerRequest")
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateControllerAddress)
	// check is address is allowed
	amap := collections.NewMapReadOnly(ctx.State(), StateVarAllowedStateControllerAddresses)
	a.Require(amap.MustHasAt(addr.Bytes()), "checkRotateStateControllerRequest: address is not allowed as next state address: %s", addr.Base58())
	// if check is successful, the block will be market as fake because this block will not be committed
	ctx.State().Set(StateVarFakeBlockMarker, []byte{0xFF})
	return nil, nil
}

func addAllowedStateControllerAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "addAllowedStateControllerAddress")
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateControllerAddress)
	amap := collections.NewMap(ctx.State(), StateVarAllowedStateControllerAddresses)
	amap.MustSetAt(addr.Bytes(), []byte{0xFF})
	return nil, nil
}

func removeAllowedStateControllerAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "removeAllowedStateControllerAddress")
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateControllerAddress)
	amap := collections.NewMap(ctx.State(), StateVarAllowedStateControllerAddresses)
	amap.MustDelAt(addr.Bytes())
	return nil, nil
}

func getAllowedStateControllerAddresses(ctx coretypes.SandboxView) (dict.Dict, error) {
	amap := collections.NewMapReadOnly(ctx.State(), StateVarAllowedStateControllerAddresses)
	if amap.MustLen() == 0 {
		return nil, nil
	}
	ret := dict.New()
	retArr := collections.NewArray16(ret, ParamAllowedStateControllerAddresses)
	amap.MustIterateKeys(func(elemKey []byte) bool {
		retArr.MustPush(elemKey)
		return true
	})
	return ret, nil
}
