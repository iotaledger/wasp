package inccounter

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func eventCounter(ctx isc.Sandbox, val int64) {
	var buf []byte
	buf = append(buf, util.Int64To8Bytes(val)...)
	ctx.Event("inccounter.counter", buf)
}
