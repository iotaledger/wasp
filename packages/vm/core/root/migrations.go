package root

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (s *StateWriter) SetSchemaVersion(v isc.SchemaVersion) {
	s.state.Set(varSchemaVersion, codec.Uint32.Encode(uint32(v)))
}

func (s *StateReader) GetSchemaVersion() isc.SchemaVersion {
	return isc.SchemaVersion(lo.Must(codec.Uint32.Decode(s.state.Get(varSchemaVersion), 0)))
}
