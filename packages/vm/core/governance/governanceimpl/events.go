package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventRotate(ctx isc.Sandbox, newAddr *cryptolib.Address, oldAddr *cryptolib.Address) {
	ww := rwutil.NewBytesWriter()
	ww.Write(newAddr)
	ww.Write(oldAddr)
	ctx.Event("coregovernance.rotate", ww.Bytes())
}
