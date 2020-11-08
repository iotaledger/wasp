package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *sandbox) ID() coretypes.RequestID {
	return *s.vmctx.Request().RequestID()
}

func (s *sandbox) EntryPointCode() coretypes.EntryPointCode {
	return s.vmctx.Request().RequestSection().EntryPointCode()
}

// addresses of request transaction inputs
func (s *sandbox) MustSenderAddress() address.Address {
	sender := s.MustSender()
	if !sender.IsAddress() {
		panic("sender must be address, not contract")
	}
	return sender.MustAddress()
}

// addresses of request transaction inputs
func (s *sandbox) MustSender() coretypes.AgentID {
	req := s.vmctx.Request()
	prop := req.Tx.MustProperties()
	if prop.IsState() {
		return coretypes.NewAgentIDFromAddress(*s.vmctx.Request().Tx.MustProperties().Sender())
	}
	senderContractID := coretypes.NewContractID(*prop.MustChainID(), req.RequestSection().SenderContractIndex())
	return coretypes.NewAgentIDFromContractID(senderContractID)
}

//MintedBalances return total minted tokens minus number of
func (s *sandbox) NumFreeMintedTokens() int64 {
	return s.vmctx.Request().Tx.MustProperties().NumFreeMintedTokens()
}

func (s *sandbox) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}
