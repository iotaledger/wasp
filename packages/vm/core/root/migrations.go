package root

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func SetSchemaVersion(state kv.KVStore, v isc.SchemaVersion) {
	state.Set(VarSchemaVersion, codec.Uint32.Encode(uint32(v)))
}

func getSchemaVersion(state kv.KVStoreReader) isc.SchemaVersion {
	return isc.SchemaVersion(codec.Uint32.MustDecode(state.Get(VarSchemaVersion), 0))
}
