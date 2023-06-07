package accounts

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func eventFoundryCreated(ctx isc.Sandbox, foundrySN uint32) {
	w := new(bytes.Buffer)
	_ = util.WriteUint32(w, foundrySN)
	ctx.Event("coreaccounts.foundryCreated", w.Bytes())
}

func eventFoundryDestroyed(ctx isc.Sandbox, foundrySN uint32) {
	w := new(bytes.Buffer)
	_ = util.WriteUint32(w, foundrySN)
	ctx.Event("coreaccounts.foundryDestroyed", w.Bytes())
}

func eventFoundryModified(ctx isc.Sandbox, foundrySN uint32) {
	w := new(bytes.Buffer)
	_ = util.WriteUint32(w, foundrySN)
	ctx.Event("coreaccounts.foundryModified", w.Bytes())
}
