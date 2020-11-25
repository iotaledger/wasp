package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *sandbox) RequestID() coretypes.RequestID {
	return *s.vmctx.Request().RequestID()
}

func (s *sandbox) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}
