package governanceimpl

import (
	"bytes"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventRotate(ctx isc.Sandbox, newAddr iotago.Address, oldAddr iotago.Address) {
	w := new(bytes.Buffer)
	_ = rwutil.WriteN(w, isc.BytesFromAddress(newAddr))
	_ = rwutil.WriteN(w, isc.BytesFromAddress(oldAddr))
	ctx.Event("coregovernance.rotate", w.Bytes())
}
