package viewcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/util"
)

type sandboxview struct {
	vctx   *viewcontext
	params codec.ImmutableCodec
	state  codec.ImmutableMustCodec
}

func NewSandboxView(vctx *viewcontext, contractIndex uint16, params codec.ImmutableCodec) *sandboxview {
	return &sandboxview{
		vctx:   vctx,
		params: params,
		state:  codec.NewMustCodec(subrealm.New(vctx.state, kv.Key(util.Uint16To2Bytes(contractIndex)))),
	}
}

func (s *sandboxview) Params() codec.ImmutableCodec {
	return s.params
}

func (s *sandboxview) State() codec.ImmutableMustCodec {
	return s.state
}

func (s *sandboxview) Accounts() coretypes.ColoredAccountsImmutable {
	panic("not implemented") // TODO: Implement
}

func (s *sandboxview) Call(contractIndex uint16, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vctx.CallView(contractIndex, entryPoint, params)
}
