package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// access to the request block
type requestWrapper struct {
	ref *sctransaction.RequestRef
}

func (r *requestWrapper) ID() sctransaction.RequestId {
	return *r.ref.RequestId()
}

func (r *requestWrapper) Code() sctransaction.RequestCode {
	return r.ref.RequestBlock().RequestCode()
}

func (r *requestWrapper) Args() kv.RCodec {
	return r.ref.RequestBlock().Args()
}

// addresses of request transaction inputs
func (r *requestWrapper) Sender() address.Address {
	return *r.ref.Tx.MustProperties().Sender()
}
