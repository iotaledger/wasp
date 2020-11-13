package accountsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

func initialize(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	ctx.Eventf("accountsc.initialize.begin")
	state := ctx.AccessState()
	if state.Get(VarStateInitialized) != nil {
		return nil, fmt.Errorf("accountsc.initialize.fail: already_initialized")
	}
	state.Set(VarStateInitialized, []byte{0xFF})
	ctx.Eventf("sccountsc.initialize.success hname = %s", Hname.String())
	return nil, nil
}

func fetchAccount(ctx vmtypes.Sandbox) (codec.ImmutableCodec, error) {
	panic("implement me")
}
