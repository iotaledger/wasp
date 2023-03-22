package root

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func SetSchemaVersion(state kv.KVStore, v uint32) {
	state.Set(StateVarSchemaVersion, codec.EncodeUint32(v))
}

func GetSchemaVersion(state kv.KVStoreReader) uint32 {
	return codec.MustDecodeUint32(state.MustGet(StateVarSchemaVersion), 0)
}
