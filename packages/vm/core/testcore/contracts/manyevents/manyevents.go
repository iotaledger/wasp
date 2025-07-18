// Package manyevents contains test helpers for contract events
package manyevents

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

var (
	ContractName = "ManyEventsContract"
	Contract     = coreutil.NewContract(ContractName)

	FuncManyEvents = Contract.Func("manyevents")
	FuncBigEvent   = Contract.Func("bigevent")

	Processor = Contract.Processor(nil,
		FuncManyEvents.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
			n := codec.MustDecode[uint32](ctx.Params().MustAt(0))
			for i := uint32(0); i < n; i++ {
				ctx.Event("event.test", codec.Encode(n))
			}
			return nil
		}),
		FuncBigEvent.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
			n := codec.MustDecode[uint32](ctx.Params().MustAt(0))
			ctx.Event("event.big", make([]byte, n))
			return nil
		}),
	)
)
