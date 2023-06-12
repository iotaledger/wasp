package inccounter

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventCounter(ctx isc.Sandbox, val int64) {
	w := new(bytes.Buffer)
	_ = rwutil.WriteInt64(w, val)
	ctx.Event("inccounter.counter", w.Bytes())
}
