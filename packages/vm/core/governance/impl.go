package blocklog

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

func addAllowedCommitteeAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "addAllowedCommitteeAddress")
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateAddress)
	amap := collections.NewMap(ctx.State(), StateVarAllowedCommitteeAddresses)
	amap.MustSetAt(addr.Bytes(), []byte{0xFF})
	return nil, nil
}

func removeAllowedCommitteeAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	a.RequireChainOwner(ctx, "removeAllowedCommitteeAddress")
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateAddress)
	amap := collections.NewMap(ctx.State(), StateVarAllowedCommitteeAddresses)
	amap.MustDelAt(addr.Bytes())
	return nil, nil
}

func isAllowedCommitteeAddress(ctx coretypes.SandboxView) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	addr := par.MustGetAddress(ParamStateAddress)
	amap := collections.NewMapReadOnly(ctx.State(), StateVarAllowedCommitteeAddresses)
	exists := amap.MustHasAt(addr.Bytes())
	ret := dict.New()
	if exists {
		ret.Set(ParamIsAllowedAddress, []byte{0xFF})
	}
	return ret, nil
}

func moveToAddress(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Panicf("moveToAddress: not implemented")
	return nil, nil
}
