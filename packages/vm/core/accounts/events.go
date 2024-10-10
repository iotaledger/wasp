package accounts

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func eventCoinCreated(ctx isc.Sandbox, treasuryCapID iotago.ObjectID) {
	ctx.Event("coreaccounts.coinCreated", bcs.MustMarshal(&treasuryCapID))
}

func eventCoinDestroyed(ctx isc.Sandbox, treasuryCapID iotago.ObjectID) {
	ctx.Event("coreaccounts.coinDestroyed", bcs.MustMarshal(&treasuryCapID))
}

func eventCoinModified(ctx isc.Sandbox, treasuryCapID iotago.ObjectID) {
	ctx.Event("coreaccounts.coinModified", bcs.MustMarshal(&treasuryCapID))
}
