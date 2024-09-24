package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func eventStore(ctx isc.Sandbox, blobHash hashing.HashValue) {
	ctx.Event("coreblob.store", bcs.MustMarshal(&blobHash))
}
