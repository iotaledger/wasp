// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/mr-tron/base58"
)

// all type id values should exactly match their counterpart values on the client!

//nolint:revive
const (
	OBJTYPE_ARRAY    int32 = 0x20
	OBJTYPE_ARRAY16  int32 = 0x30
	OBJTYPE_CALL     int32 = 0x40
	OBJTYPE_TYPEMASK int32 = 0x0f

	OBJTYPE_ADDRESS    int32 = 1
	OBJTYPE_AGENT_ID   int32 = 2
	OBJTYPE_BYTES      int32 = 3
	OBJTYPE_CHAIN_ID   int32 = 4
	OBJTYPE_COLOR      int32 = 5
	OBJTYPE_HASH       int32 = 6
	OBJTYPE_HNAME      int32 = 7
	OBJTYPE_INT16      int32 = 8
	OBJTYPE_INT32      int32 = 9
	OBJTYPE_INT64      int32 = 10
	OBJTYPE_MAP        int32 = 11
	OBJTYPE_REQUEST_ID int32 = 12
	OBJTYPE_STRING     int32 = 13

	OBJID_NULL    int32 = 0
	OBJID_ROOT    int32 = 1
	OBJID_STATE   int32 = 2
	OBJID_PARAMS  int32 = 3
	OBJID_RESULTS int32 = 4
)

// flag to indicate that this key id originally comes from a bytes key
// this allows us to display better readable tracing information
const KeyFromBytes int32 = 0x4000

type HostObject interface {
	CallFunc(keyID int32, params []byte) []byte
	Exists(keyID, typeID int32) bool
	GetBytes(keyID, typeID int32) []byte
	GetObjectID(keyID, typeID int32) int32
	GetTypeID(keyID int32) int32
	SetBytes(keyID, typeID int32, bytes []byte)
}

// KvStoreHost implements WaspLib.client.ScHost interface
// it allows WasmGoVM to bypass Wasm and access the sandbox directly
// so that it is possible to debug into SC code that was written in Go
type KvStoreHost struct {
	scKeys     *KvStoreHost
	keyIDToKey [][]byte
	keyToKeyID map[string]int32
	objIDToObj []HostObject
}

func (h *KvStoreHost) Init(scKeys *KvStoreHost) {
	h.scKeys = scKeys
	h.keyIDToKey = [][]byte{[]byte("<null>")}
	h.keyToKeyID = make(map[string]int32)
	h.objIDToObj = make([]HostObject, 0, 16)
}

func (h *KvStoreHost) CallFunc(objID, keyID int32, params []byte) []byte {
	return h.FindObject(objID).CallFunc(keyID, params)
}

func (h *KvStoreHost) Exists(objID, keyID, typeID int32) bool {
	return h.FindObject(objID).Exists(keyID, typeID)
}

func (h *KvStoreHost) FindObject(objID int32) HostObject {
	if objID < 0 || objID >= int32(len(h.objIDToObj)) {
		panic("FindObject: invalid objID")
	}
	return h.objIDToObj[objID]
}

func (h *KvStoreHost) GetBytes(objID, keyID, typeID int32) []byte {
	obj := h.FindObject(objID)
	if !obj.Exists(keyID, typeID) {
		h.Tracef("GetBytes o%d k%d missing key", objID, keyID)
		return nil
	}
	bytes := obj.GetBytes(keyID, typeID)
	switch typeID {
	case OBJTYPE_INT16:
		val16, err := codec.DecodeInt16(bytes, 0)
		if err != nil {
			panic("GetBytes: invalid int16")
		}
		h.Tracef("GetBytes o%d k%d = %ds", objID, keyID, val16)
	case OBJTYPE_INT32:
		val32, err := codec.DecodeInt32(bytes, 0)
		if err != nil {
			panic("GetBytes: invalid int32")
		}
		h.Tracef("GetBytes o%d k%d = %di", objID, keyID, val32)
	case OBJTYPE_INT64:
		val64, err := codec.DecodeInt64(bytes, 0)
		if err != nil {
			panic("GetBytes: invalid int64")
		}
		h.Tracef("GetBytes o%d k%d = %dl", objID, keyID, val64)
	case OBJTYPE_STRING:
		h.Tracef("GetBytes o%d k%d = '%s'", objID, keyID, string(bytes))
	default:
		h.Tracef("GetBytes o%d k%d = '%s'", objID, keyID, base58.Encode(bytes))
	}
	return bytes
}

func (h *KvStoreHost) getKeyFromID(keyID int32) []byte {
	if keyID < 0 {
		return predefinedKeys[-keyID]
	}

	// find user-defined key
	keyID &= ^KeyFromBytes
	if h.scKeys != nil {
		// get common SC user-defined keys count (skip null)
		scKeysLen := int32(len(h.scKeys.keyIDToKey)) - 1

		// find common SC user-defined key
		if keyID <= scKeysLen {
			return h.scKeys.keyIDToKey[keyID]
		}

		// correct for SC keys count
		keyID -= scKeysLen
	}

	// user-defined key must be local
	return h.keyIDToKey[keyID]
}

func (h *KvStoreHost) GetKeyFromID(keyID int32) []byte {
	h.TraceAllf("GetKeyFromID(k%d)", keyID)
	key := h.getKeyFromID(keyID)
	if (keyID & (KeyFromBytes | -0x80000000)) == KeyFromBytes {
		// originally a byte slice key
		h.Tracef("GetKeyFromID k%d='%s'", keyID, base58.Encode(key))
		return key
	}
	// originally a string key
	h.Tracef("GetKeyFromID k%d='%s'", keyID, string(key))
	return key
}

func (h *KvStoreHost) getKeyID(key []byte, fromBytes bool) int32 {
	// cannot use []byte as key in maps
	// so we will convert to (non-utf8) string
	// most will have started out as string anyway
	keyString := string(key)

	var scKeysLen int32
	if h.scKeys != nil {
		// check common SC user-defined keys
		keyID, ok := h.scKeys.keyToKeyID[keyString]
		if ok {
			return keyID
		}

		// get common SC user-defined keys count (skip null)
		scKeysLen = int32(len(h.scKeys.keyIDToKey)) - 1
	}

	// check local user-defined keys
	keyID, ok := h.keyToKeyID[keyString]
	if ok {
		return keyID
	}

	// unknown key, add it to local user-defined keys
	keyID = int32(len(h.keyIDToKey)) + scKeysLen
	if fromBytes {
		keyID |= KeyFromBytes
	}
	h.keyToKeyID[keyString] = keyID
	h.keyIDToKey = append(h.keyIDToKey, key)
	return keyID
}

func (h *KvStoreHost) GetKeyIDFromBytes(bytes []byte) int32 {
	keyID := h.getKeyID(bytes, true)
	h.Tracef("GetKeyIDFromBytes '%s'=k%d", base58.Encode(bytes), keyID)
	return keyID
}

func (h *KvStoreHost) GetKeyIDFromString(key string) int32 {
	keyID := h.getKeyID([]byte(key), false)
	h.Tracef("GetKeyIDFromString '%s'=k%d", key, keyID)
	return keyID
}

func (h *KvStoreHost) GetKeyStringFromID(keyID int32) string {
	return string(h.GetKeyFromID(keyID))
}

func (h *KvStoreHost) GetObjectID(objID, keyID, typeID int32) int32 {
	h.TraceAllf("GetObjectID(o%d,k%d,t%d)", objID, keyID, typeID)
	subID := h.FindObject(objID).GetObjectID(keyID, typeID)
	h.Tracef("GetObjectID o%d k%d t%d = o%d", objID, keyID, typeID, subID)
	return subID
}

func (h *KvStoreHost) SetBytes(objID, keyID, typeID int32, bytes []byte) {
	h.FindObject(objID).SetBytes(keyID, typeID, bytes)
	switch typeID {
	case OBJTYPE_INT16:
		val16, err := codec.DecodeInt16(bytes, 0)
		if err != nil {
			panic("SetBytes: invalid int16")
		}
		h.Tracef("SetBytes o%d k%d v=%ds", objID, keyID, val16)
	case OBJTYPE_INT32:
		val32, err := codec.DecodeInt32(bytes, 0)
		if err != nil {
			panic("SetBytes: invalid int32")
		}
		h.Tracef("SetBytes o%d k%d v=%di", objID, keyID, val32)
	case OBJTYPE_INT64:
		val64, err := codec.DecodeInt64(bytes, 0)
		if err != nil {
			panic("SetBytes: invalid int64")
		}
		h.Tracef("SetBytes o%d k%d v=%dl", objID, keyID, val64)
	case OBJTYPE_STRING:
		if keyID != KeyTrace {
			h.Tracef("SetBytes o%d k%d v='%s'", objID, keyID, string(bytes))
		}
	default:
		h.Tracef("SetBytes o%d k%d v='%s'", objID, keyID, base58.Encode(bytes))
	}
}

func (h *KvStoreHost) Tracef(format string, a ...interface{}) {
	if HostTracing {
		text := fmt.Sprintf(format, a...)
		h.SetBytes(OBJID_ROOT, KeyTrace, OBJTYPE_STRING, []byte(text))
	}
}

func (h *KvStoreHost) TraceAllf(format string, a ...interface{}) {
	if HostTracingAll {
		h.Tracef(format, a...)
	}
}

func (h *KvStoreHost) TrackObject(obj HostObject) int32 {
	objID := int32(len(h.objIDToObj))
	h.objIDToObj = append(h.objIDToObj, obj)
	return objID
}
