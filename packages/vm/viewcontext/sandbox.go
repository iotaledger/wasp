package viewcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type sandboxview struct {
	vctx   *viewcontext
	params codec.ImmutableCodec
	state  codec.ImmutableMustCodec
}

func NewSandboxView(vctx *viewcontext, contractHname coretypes.Hname, params codec.ImmutableCodec) *sandboxview {
	return &sandboxview{
		vctx:   vctx,
		params: params,
		state:  codec.NewMustCodec(subrealm.New(vctx.state, kv.Key(contractHname.Bytes()))),
	}
}

func (s *sandboxview) Params() codec.ImmutableCodec {
	return s.params
}

func (s *sandboxview) State() codec.ImmutableMustCodec {
	return s.state
}

func (s *sandboxview) Account() coretypes.ColoredBalances {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vctx.CallView(contractHname, entryPoint, params)
}
