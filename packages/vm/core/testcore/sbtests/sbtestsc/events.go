package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func eventCounter(ctx isc.Sandbox, value uint64) {
	var buf []byte
	buf = append(buf, util.Uint64To8Bytes(value)...)
	ctx.Event("testcore.counter", buf)
}

func eventTest(ctx isc.Sandbox) {
	var buf []byte
	ctx.Event("testcore.test", buf)
}
