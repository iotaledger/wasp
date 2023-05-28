package accounts

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func eventFoundryCreated(ctx isc.Sandbox, foundrySN uint32) {
	var buf []byte
	buf = append(buf, util.Uint32To4Bytes(foundrySN)...)
	ctx.Event("coreaccounts.foundryCreated", buf)
}

func eventFoundryDestroyed(ctx isc.Sandbox, foundrySN uint32) {
	var buf []byte
	buf = append(buf, util.Uint32To4Bytes(foundrySN)...)
	ctx.Event("coreaccounts.foundryDestroyed", buf)
}

func eventFoundryModified(ctx isc.Sandbox, foundrySN uint32) {
	var buf []byte
	buf = append(buf, util.Uint32To4Bytes(foundrySN)...)
	ctx.Event("coreaccounts.foundryModified", buf)
}
