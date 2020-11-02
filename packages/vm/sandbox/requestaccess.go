package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *sandbox) ID() coretypes.RequestID {
	return *s.vmctx.Request().RequestID()
}

func (s *sandbox) Code() coretypes.EntryPointCode {
	return s.vmctx.Request().RequestSection().EntryPointCode()
}

func (s *sandbox) Args() codec.ImmutableCodec {
	return s.vmctx.Request().RequestSection().Args()
}

// addresses of request transaction inputs
func (s *sandbox) SenderAddress() address.Address {
	return *s.vmctx.Request().Tx.MustProperties().Sender()
}

//MintedBalances return total minted tokens minus number of
func (s *sandbox) NumFreeMintedTokens() int64 {
	return s.vmctx.Request().Tx.MustProperties().NumFreeMintedTokens()
}
