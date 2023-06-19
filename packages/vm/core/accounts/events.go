package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventFoundryCreated(ctx isc.Sandbox, foundrySN uint32) {
	ww := rwutil.NewBytesWriter()
	ww.WriteUint32(foundrySN)
	ctx.Event("coreaccounts.foundryCreated", ww.Bytes())
}

func eventFoundryDestroyed(ctx isc.Sandbox, foundrySN uint32) {
	ww := rwutil.NewBytesWriter()
	ww.WriteUint32(foundrySN)
	ctx.Event("coreaccounts.foundryDestroyed", ww.Bytes())
}

func eventFoundryModified(ctx isc.Sandbox, foundrySN uint32) {
	ww := rwutil.NewBytesWriter()
	ww.WriteUint32(foundrySN)
	ctx.Event("coreaccounts.foundryModified", ww.Bytes())
}
