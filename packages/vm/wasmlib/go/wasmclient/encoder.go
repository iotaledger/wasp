package wasmclient

import "encoding/binary"

type Encoder struct{}

func fromBase58(value string, typeID int32) []byte {
	bytes := Base58Decode(value)
	if len(bytes) != int(TypeSizes[typeID]) {
		panic("invalid byte size")
	}
	return bytes
}

func (c Encoder) FromAddress(value Address) []byte {
	return fromBase58(string(value), TYPE_ADDRESS)
}

func (c Encoder) FromAgentID(value AgentID) []byte {
	return fromBase58(string(value), TYPE_AGENT_ID)
}

func (c Encoder) FromBool(value bool) []byte {
	bytes := []byte{0}
	if value {
		bytes[0] = 1
	}
	return bytes
}

func (c Encoder) FromBytes(value []byte) []byte {
	if value == nil {
		panic("missing mandatory byte array")
	}
	return value
}

func (c Encoder) FromChainID(value ChainID) []byte {
	return fromBase58(string(value), TYPE_CHAIN_ID)
}

func (c Encoder) FromColor(value Color) []byte {
	color := string(value)
	if color == COLOR_IOTA {
		color = COLOR_IOTA_BASE58
	}
	return fromBase58(color, TYPE_COLOR)
}

func (c Encoder) FromHash(value Hash) []byte {
	return fromBase58(string(value), TYPE_HASH)
}

func (c Encoder) FromHname(value Hname) []byte {
	return c.FromUint32(uint32(value))
}

func (c Encoder) FromInt8(value int8) []byte {
	return c.FromUint8(uint8(value))
}

func (c Encoder) FromInt16(value int16) []byte {
	return c.FromUint16(uint16(value))
}

func (c Encoder) FromInt32(value int32) []byte {
	return c.FromUint32(uint32(value))
}

func (c Encoder) FromInt64(value int64) []byte {
	return c.FromUint64(uint64(value))
}

func (c Encoder) FromRequestID(value RequestID) []byte {
	return fromBase58(string(value), TYPE_REQUEST_ID)
}

func (c Encoder) FromString(value string) []byte {
	return []byte(value)
}

func (c Encoder) FromUint8(value uint8) []byte {
	return []byte{value}
}

func (c Encoder) FromUint16(value uint16) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, value)
	return bytes
}

func (c Encoder) FromUint32(value uint32) []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, value)
	return bytes
}

func (c Encoder) FromUint64(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)
	return bytes
}
