// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmclient

import (
	"strconv"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// The Arguments struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
type Arguments struct {
	args dict.Dict
}

func (a *Arguments) set(key string, val []byte) {
	if a.args == nil {
		a.args = make(dict.Dict)
	}
	a.args[kv.Key(key)] = val
}

func (a *Arguments) setBase58(key, val string, typeID int32) {
	bytes := Base58Decode(val)
	if len(bytes) != int(TypeSizes[typeID]) {
		panic("invalid byte size")
	}
	a.set(key, bytes)
}

func (a *Arguments) IndexedKey(key string, index int) string {
	return key + "." + strconv.Itoa(index)
}

func (a *Arguments) Mandatory(key string) {
	if a.args != nil {
		if _, ok := a.args[kv.Key(key)]; ok {
			return
		}
	}
	panic("missing mandatory " + key)
}

func (a *Arguments) SetAddress(key string, val Address) {
	a.setBase58(key, string(val), TYPE_ADDRESS)
}

func (a *Arguments) SetAgentID(key string, val AgentID) {
	a.setBase58(key, string(val), TYPE_AGENT_ID)
}

func (a *Arguments) SetBool(key string, val bool) {
	bytes := []byte{0}
	if val {
		bytes[0] = 1
	}
	a.set(key, bytes)
}

func (a *Arguments) SetBytes(key string, val []byte) {
	a.set(key, val)
}

func (a *Arguments) SetColor(key string, val Color) {
	a.setBase58(key, string(val), TYPE_COLOR)
}

func (a *Arguments) SetChainID(key string, val ChainID) {
	a.setBase58(key, string(val), TYPE_CHAIN_ID)
}

func (a *Arguments) SetHash(key string, val Hash) {
	a.setBase58(key, string(val), TYPE_HASH)
}

func (a *Arguments) SetHname(key string, val Hname) {
	a.SetUint32(key, uint32(val))
}

func (a *Arguments) SetInt8(key string, val int8) {
	a.set(key, []byte{byte(val)})
}

func (a *Arguments) SetInt16(key string, val int16) {
	a.setUint64(key, uint64(val), 2)
}

func (a *Arguments) SetInt32(key string, val int32) {
	a.setUint64(key, uint64(val), 4)
}

func (a *Arguments) SetInt64(key string, val int64) {
	a.setUint64(key, uint64(val), 4)
}

func (a *Arguments) SetRequestID(key string, val RequestID) {
	a.setBase58(key, string(val), TYPE_REQUEST_ID)
}

func (a *Arguments) SetString(key, val string) {
	a.set(key, []byte(val))
}

func (a *Arguments) SetUint8(key string, val uint8) {
	a.set(key, []byte{val})
}

func (a *Arguments) SetUint16(key string, val uint16) {
	a.setUint64(key, uint64(val), 2)
}

func (a *Arguments) SetUint32(key string, val uint32) {
	a.setUint64(key, uint64(val), 4)
}

func (a *Arguments) SetUint64(key string, val uint64) {
	a.setUint64(key, val, 4)
}

func (a *Arguments) setUint64(key string, val uint64, size int) {
	bytes := make([]byte, size)
	for i := 0; i < size; i++ {
		bytes[i] = byte(val)
		val >>= 8
	}
	a.set(key, bytes)
}
