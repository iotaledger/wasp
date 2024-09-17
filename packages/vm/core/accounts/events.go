package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

func eventCoinCreated(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ctx.Event("coreaccounts.coinCreated", bcs.MustMarshal(&treasuryCapID))
}

func eventCoinDestroyed(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ctx.Event("coreaccounts.coinDestroyed", bcs.MustMarshal(&treasuryCapID))
}

func eventCoinModified(ctx isc.Sandbox, treasuryCapID sui.ObjectID) {
	ctx.Event("coreaccounts.coinModified", bcs.MustMarshal(&treasuryCapID))
}
