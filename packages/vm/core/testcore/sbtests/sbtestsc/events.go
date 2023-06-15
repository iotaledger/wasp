package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventCounter(ctx isc.Sandbox, value uint64) {
	ww := rwutil.NewBytesWriter()
	ww.WriteUint64(value)
	ctx.Event("testcore.counter", ww.Bytes())
}

func eventTest(ctx isc.Sandbox) {
	ww := rwutil.NewBytesWriter()
	ctx.Event("testcore.test", ww.Bytes())
}
