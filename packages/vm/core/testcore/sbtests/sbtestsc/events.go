package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func eventCounter(ctx isc.Sandbox, value uint64) {
	ctx.Event("testcore.counter", bcs.MustMarshal(&value))
}

func eventTest(ctx isc.Sandbox) {
	ctx.Event("testcore.test", nil)
}
