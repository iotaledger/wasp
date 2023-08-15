package root

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func SetSchemaVersion(state kv.KVStore, v uint32) {
	state.Set(VarSchemaVersion, codec.EncodeUint32(v))
}

func GetSchemaVersion(state kv.KVStoreReader) uint32 {
	return codec.MustDecodeUint32(state.Get(VarSchemaVersion), 0)
}
