package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventCounter(ctx isc.Sandbox, val int64) {
	ww := rwutil.NewBytesWriter()
	ww.WriteInt64(val)
	ctx.Event("inccounter.counter", ww.Bytes())
}
