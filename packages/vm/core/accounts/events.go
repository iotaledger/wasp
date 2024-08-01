package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func eventCoinCreated(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ww := rwutil.NewBytesWriter()
	ww.WriteN(treasuryCapID[:])
	ctx.Event("coreaccounts.coinCreated", ww.Bytes())
}

func eventCoinDestroyed(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ww := rwutil.NewBytesWriter()
	ww.WriteN(treasuryCapID[:])
	ctx.Event("coreaccounts.coinDestroyed", ww.Bytes())
}

func eventCoinModified(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ww := rwutil.NewBytesWriter()
	ww.WriteN(treasuryCapID[:])
	ctx.Event("coreaccounts.coinModified", ww.Bytes())
}
