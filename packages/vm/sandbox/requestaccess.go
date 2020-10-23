package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/sctransaction"
)

// access to the request block
type requestWrapper struct {
	ref *sctransaction.RequestRef
}

func (r *requestWrapper) ID() coretypes.RequestID {
	return *r.ref.RequestID()
}

func (r *requestWrapper) Code() coretypes.EntryPointCode {
	return r.ref.RequestBlock().EntryPointCode()
}

func (r *requestWrapper) Args() kv.RCodec {
	return r.ref.RequestBlock().Args()
}

// addresses of request transaction inputs
func (r *requestWrapper) Sender() address.Address {
	return *r.ref.Tx.MustProperties().Sender()
}

//MintedBalances return total minted tokens minus number of
func (r *requestWrapper) NumFreeMintedTokens() int64 {
	return r.ref.Tx.MustProperties().NumFreeMintedTokens()
}
