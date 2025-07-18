package sbtestsc

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

func eventCounter(ctx isc.Sandbox, value uint64) {
	ctx.Event("testcore.counter", bcs.MustMarshal(&value))
}

func eventTest(ctx isc.Sandbox) {
	ctx.Event("testcore.test", nil)
}
