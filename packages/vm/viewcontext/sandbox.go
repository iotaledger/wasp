package viewcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type sandboxview struct {
	vctx   *viewcontext
	params codec.ImmutableCodec
}

func NewSandboxView(vctx *viewcontext, params codec.ImmutableCodec) *sandboxview {
	return &sandboxview{
		vctx:   vctx,
		params: params,
	}
}

func (s *sandboxview) Params() codec.ImmutableCodec {
	return s.params
}

func (s *sandboxview) State() codec.ImmutableMustCodec {
	return s.vctx.state
}

func (s *sandboxview) Accounts() coretypes.ColoredAccountsImmutable {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractIndex uint16, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vctx.CallView(contractIndex, entryPoint, params)
}
