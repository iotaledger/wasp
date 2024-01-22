package root

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func SetSchemaVersion(state kv.KVStore, v isc.SchemaVersion) {
	state.Set(VarSchemaVersion, codec.EncodeUint32(uint32(v)))
}

func getSchemaVersion(state kv.KVStoreReader) isc.SchemaVersion {
	return isc.SchemaVersion(codec.MustDecodeUint32(state.Get(VarSchemaVersion), 0))
}
