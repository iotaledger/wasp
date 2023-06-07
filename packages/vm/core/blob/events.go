package blob

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
)

func eventStore(ctx isc.Sandbox, blobHash hashing.HashValue) {
	w := new(bytes.Buffer)
	_ = blobHash.Write(w)
	ctx.Event("coreblob.store", w.Bytes())
}
