package wasmclient

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type Request struct {
	err error
	id  *iscp.RequestID
}

func (r Request) Error() error {
	return r.err
}
