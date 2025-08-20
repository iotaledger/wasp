package legacy

import (
	"encoding/binary"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

// EncodeUint32ForDummyTX encodes uint32s in the same way as it was on Stardust
// This is only needed for DummyTX to ensure legacy compatibility.
func EncodeUint32ForDummyTX(v isc.SchemaVersion, val uint32) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return codec.Encode(val)
	}

	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], val)
	return b[:]
}

// EncodeUint16ForDummyTX encodes uint16s in the same way as it was on Stardust
// This is only needed for DummyTX to ensure legacy compatibility.
func EncodeUint16ForDummyTX(v isc.SchemaVersion, val uint16) []byte {
	if v > allmigrations.SchemaVersionMigratedRebased {
		return codec.Encode(val)
	}

	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], val)
	return b[:]
}
