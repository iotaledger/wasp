package legacy

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
)

func WithDummyUint32(v isc.SchemaVersion, val uint32) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return codec.Encode(val)
	}

	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], val)
	return b[:]
}

func WithDummyUint16(v isc.SchemaVersion, val uint16) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return codec.Encode(val)
	}

	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], val)
	return b[:]
}
