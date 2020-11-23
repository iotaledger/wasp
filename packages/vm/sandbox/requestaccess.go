package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *sandbox) RequestID() coretypes.RequestID {
	return *s.vmctx.Request().RequestID()
}

// addresses of request transaction inputs
func (s *sandbox) MustSender() coretypes.AgentID {
	req := s.vmctx.Request()
	prop := req.Tx.MustProperties()
	if !prop.IsState() {
		return coretypes.NewAgentIDFromAddress(*s.vmctx.Request().Tx.MustProperties().SenderAddress())
	}
	senderContractID := coretypes.NewContractID(*prop.MustChainID(), req.RequestSection().SenderContractHname())
	return coretypes.NewAgentIDFromContractID(senderContractID)
}

func (s *sandbox) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}
