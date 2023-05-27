package governanceimpl

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func eventRotate(ctx isc.Sandbox, newAddr iotago.Address, oldAddr iotago.Address) {
	var buf []byte
	buf = append(buf, isc.BytesFromAddress(newAddr)...)
	buf = append(buf, isc.BytesFromAddress(oldAddr)...)
	ctx.Event("governance.rotate", buf)
}
