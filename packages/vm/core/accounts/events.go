package accounts

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
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
