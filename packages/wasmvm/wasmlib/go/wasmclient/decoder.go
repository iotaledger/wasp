package wasmclient

import "encoding/binary"

type Decoder struct{}

func checkDefault(bytes []byte, typeID int32) []byte {
	size := TypeSizes[typeID]
	if bytes == nil {
		// return default all-zero bytes value
		return make([]byte, size)
	}
	if size != 0 && len(bytes) != int(size) {
		panic("invalid type size")
	}
	return bytes
}

func toBase58(bytes []byte, typeID int32) string {
	return Base58Encode(checkDefault(bytes, typeID))
}

func (c Decoder) ToAddress(bytes []byte) Address {
	return Address(toBase58(bytes, TYPE_ADDRESS))
}

func (c Decoder) ToAgentID(bytes []byte) AgentID {
	return AgentID(toBase58(bytes, TYPE_AGENT_ID))
}

func (c Decoder) ToBool(bytes []byte) bool {
	return checkDefault(bytes, TYPE_BOOL)[0] != 0
}

func (c Decoder) ToBytes(bytes []byte) []byte {
	return checkDefault(bytes, TYPE_BYTES)
}

func (c Decoder) ToChainID(bytes []byte) ChainID {
	return ChainID(toBase58(bytes, TYPE_CHAIN_ID))
}

func (c Decoder) ToColor(bytes []byte) Color {
	color := toBase58(bytes, TYPE_COLOR)
	if color == COLOR_IOTA_BASE58 {
		color = COLOR_IOTA
	}
	return Color(color)
}

func (c Decoder) ToHash(bytes []byte) Hash {
	return Hash(toBase58(bytes, TYPE_HASH))
}

func (c Decoder) ToHname(bytes []byte) Hname {
	return Hname(binary.LittleEndian.Uint32(checkDefault(bytes, TYPE_HNAME)))
}

func (c Decoder) ToInt8(bytes []byte) int8 {
	return int8(c.ToUint8(bytes))
}

func (c Decoder) ToInt16(bytes []byte) int16 {
	return int16(c.ToUint16(bytes))
}

func (c Decoder) ToInt32(bytes []byte) int32 {
	return int32(c.ToUint32(bytes))
}

func (c Decoder) ToInt64(bytes []byte) int64 {
	return int64(c.ToUint64(bytes))
}

func (c Decoder) ToRequestID(bytes []byte) RequestID {
	return RequestID(toBase58(bytes, TYPE_REQUEST_ID))
}

func (c Decoder) ToString(bytes []byte) string {
	return string(checkDefault(bytes, TYPE_STRING))
}

func (c Decoder) ToUint8(bytes []byte) uint8 {
	return checkDefault(bytes, TYPE_INT8)[0]
}

func (c Decoder) ToUint16(bytes []byte) uint16 {
	return binary.LittleEndian.Uint16(checkDefault(bytes, TYPE_INT16))
}

func (c Decoder) ToUint32(bytes []byte) uint32 {
	return binary.LittleEndian.Uint32(checkDefault(bytes, TYPE_INT32))
}

func (c Decoder) ToUint64(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(checkDefault(bytes, TYPE_INT64))
}
