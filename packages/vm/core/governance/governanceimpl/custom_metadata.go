package governanceimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

func setCustomMetadata(ctx isc.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	SetCustomMetadata(ctx.State(), string(ctx.Params().MustGet(governance.ParamCustomMetadata)))
	return nil
}

func getCustomMetadata(ctx isc.SandboxView) dict.Dict {
	return dict.Dict{
		governance.ParamCustomMetadata: ctx.StateR().MustGet(governance.VarCustomMetadata),
	}
}

func SetCustomMetadata(state kv.KVStore, s string) {
	state.Set(governance.VarCustomMetadata, codec.EncodeString(s))
}

func GetCustomMetadata(state kv.KVStoreReader) string {
	return codec.MustDecodeString(state.MustGet(governance.VarCustomMetadata), "")
}
