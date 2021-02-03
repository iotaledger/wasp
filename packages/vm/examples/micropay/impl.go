package micropay

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func initialize(_ coretypes.Sandbox) (dict.Dict, error) {
	return nil, nil
}

// Params:
// - ParamServiceAddress address.Address
func openChannel(ctx coretypes.Sandbox) (dict.Dict, error) {
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	a := coreutil.NewAssert(ctx.Log())

	addr := par.MustGetAddress(ParamServiceAddress)

	a.Require(!ctx.State().MustHas(kv.Key(addr[:])), fmt.Sprintf("channel already open: %v", addr))

	totalWarrant := ctx.IncomingTransfer().Balance(balance.ColorIOTA)
	a.Require(totalWarrant >= MinimumWarrantIotas, fmt.Sprintf("warrant must be larger than %d iotas", MinimumWarrantIotas))

	sendBack := ctx.IncomingTransfer().TakeOutColor(balance.ColorIOTA)
	err := accounts.Deposit(ctx, ctx.Caller(), sendBack)
	a.RequireNoError(err)
	ctx.Event(fmt.Sprintf("[micropay.openChannel] %s, warrant: %d iotas", addr, totalWarrant))

	ctx.State().Set(kv.Key(addr[:]), codec.EncodeInt64(totalWarrant))
	return nil, nil
}

func getChannelInfo(_ coretypes.SandboxView) (dict.Dict, error) {
	return nil, nil
}
