package blob

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
)

func eventStore(ctx isc.Sandbox, blobHash hashing.HashValue) {
	var buf []byte
	buf = append(buf, blobHash.Bytes()...)
	ctx.Event("blob.store", buf)
}
