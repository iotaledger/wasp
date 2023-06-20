package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func eventStore(ctx isc.Sandbox, blobHash hashing.HashValue) {
	ww := rwutil.NewBytesWriter()
	ww.Write(&blobHash)
	ctx.Event("coreblob.store", ww.Bytes())
}
