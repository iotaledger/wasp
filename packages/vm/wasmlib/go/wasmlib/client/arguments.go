package client

import (
	"sort"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

// The Arguments struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
type Arguments struct {
	args map[string][]byte
}

func (a Arguments) set(key string, val []byte) {
	if a.args == nil {
		a.args = make(map[string][]byte)
	}
	a.args[key] = val
}

func (a Arguments) setBase58(key, val string, typeID int32) {
	bytes := Base58Decode(val)
	if len(bytes) != int(wasmlib.TypeSizes[typeID]) {
		panic("invalid byte size")
	}
	a.set(key, bytes)
}

func (a Arguments) Mandatory(key string) {
	if a.args != nil {
		if _, ok := a.args[key]; ok {
			return
		}
	}
	panic("missing mandatory " + key)
}

func (a Arguments) SetAddress(key string, val AgentID) {
	a.setBase58(key, string(val), wasmlib.TYPE_ADDRESS)
}

func (a Arguments) SetAgentID(key string, val AgentID) {
	a.setBase58(key, string(val), wasmlib.TYPE_AGENT_ID)
}

func (a Arguments) SetBool(key string, val bool) {
	bytes := []byte{0}
	if val {
		bytes[0] = 1
	}
	a.set(key, bytes)
}

func (a Arguments) SetBytes(key string, val []byte) {
	a.set(key, val)
}

func (a Arguments) SetColor(key string, val Color) {
	a.setBase58(key, string(val), wasmlib.TYPE_COLOR)
}

func (a Arguments) SetChainID(key string, val ChainID) {
	a.setBase58(key, string(val), wasmlib.TYPE_CHAIN_ID)
}

func (a Arguments) SetHash(key string, val Hash) {
	a.setBase58(key, string(val), wasmlib.TYPE_HASH)
}

func (a Arguments) SetInt8(key string, val int8) {
	a.set(key, []byte{byte(val)})
}

func (a Arguments) SetInt16(key string, val int16) {
	a.setUint64(key, uint64(val), 2)
}

func (a Arguments) SetInt32(key string, val int32) {
	a.setUint64(key, uint64(val), 4)
}

func (a Arguments) SetInt64(key string, val int64) {
	a.setUint64(key, uint64(val), 4)
}

func (a Arguments) SetRequestID(key string, val RequestID) {
	a.setBase58(key, string(val), wasmlib.TYPE_REQUEST_ID)
}

func (a Arguments) SetString(key, val string) {
	a.set(key, []byte(val))
}

func (a Arguments) SetUint8(key string, val uint8) {
	a.set(key, []byte{val})
}

func (a Arguments) SetUint16(key string, val uint16) {
	a.setUint64(key, uint64(val), 2)
}

func (a Arguments) SetUint32(key string, val uint32) {
	a.setUint64(key, uint64(val), 4)
}

func (a Arguments) SetUint64(key string, val uint64) {
	a.setUint64(key, val, 4)
}

func (a Arguments) setUint64(key string, val uint64, size int) {
	bytes := make([]byte, size)
	for i := 0; i < size; i++ {
		bytes[i] = byte(val)
		val >>= 8
	}
	a.set(key, bytes)
}

// Encode returns a byte array that encodes the Arguments as follows:
// Sort all keys in ascending order (very important, because this data
// will be part of the data that will be signed, so the order needs to
// be 100% deterministic). Then emit a 2-byte argument count.
// Next for each argument emit a 2-byte key length, the key prepended
// with a minus sign, a 4-byte value length, and then the value bytes.
func (a Arguments) Encode() []byte {
	keys := make([]string, 0, len(a.args))
	total := 2
	for k, v := range a.args {
		keys = append(keys, k)
		total += 2 + 1 + len(k) + 4 + len(v)
	}
	sort.Strings(keys)

	buf := make([]byte, 0, total)
	buf = append(buf, codec.EncodeUint16(uint16(len(keys)))...)
	for _, k := range keys {
		buf = append(buf, codec.EncodeUint16(uint16(len(k)+1))...)
		buf = append(buf, '-')
		buf = append(buf, []byte(k)...)
		v := a.args[k]
		buf = append(buf, codec.EncodeUint32(uint32(len(v)))...)
		buf = append(buf, v...)
	}
	return buf
}
