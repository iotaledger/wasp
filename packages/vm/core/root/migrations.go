package root

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

func (s *StateWriter) SetSchemaVersion(v isc.SchemaVersion) {
	s.state.Set(varSchemaVersion, codec.Encode[uint32](uint32(v)))
}

func (s *StateReader) GetSchemaVersion() isc.SchemaVersion {
	return isc.SchemaVersion(lo.Must(codec.Decode[uint32](s.state.Get(varSchemaVersion), 0)))
}
