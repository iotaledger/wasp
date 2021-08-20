// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/mr-tron/base58"
)

// all type id values should exactly match their counterpart values on the client!

//nolint:revive
const (
	OBJTYPE_ARRAY    int32 = 0x20
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
)

// flag to indicate that this key id originally comes from a bytes key
// this allows us to display better readable tracing information
const KeyFromBytes int32 = 0x4000

var (
	HostTracing         = false
	ExtendedHostTracing = false
)

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
	keyIDToKey    [][]byte
	keyIDToKeyMap [][]byte
	keyToKeyID    map[string]int32
	log           *logger.Logger
	objIDToObj    []HostObject
}

func (host *KvStoreHost) Init(log *logger.Logger) {
	host.log = log
	host.objIDToObj = make([]HostObject, 0, 16)
	host.keyIDToKey = [][]byte{[]byte("<null>")}
	host.keyToKeyID = make(map[string]int32)
	host.keyIDToKeyMap = make([][]byte, len(keyMap)+1)
	for k, v := range keyMap {
		host.keyIDToKeyMap[-v] = []byte(k)
	}
}

func (host *KvStoreHost) CallFunc(objID, keyID int32, params []byte) []byte {
	return host.FindObject(objID).CallFunc(keyID, params)
}

func (host *KvStoreHost) Exists(objID, keyID, typeID int32) bool {
	return host.FindObject(objID).Exists(keyID, typeID)
}

func (host *KvStoreHost) FindObject(objID int32) HostObject {
	if objID < 0 || objID >= int32(len(host.objIDToObj)) {
		panic("FindObject: invalid objID")
	}
	return host.objIDToObj[objID]
}

func (host *KvStoreHost) FindSubObject(obj HostObject, keyID, typeID int32) HostObject {
	if obj == nil {
		// use root object
		obj = host.FindObject(1)
	}
	return host.FindObject(obj.GetObjectID(keyID, typeID))
}

func (host *KvStoreHost) GetBytes(objID, keyID, typeID int32) []byte {
	obj := host.FindObject(objID)
	if !obj.Exists(keyID, typeID) {
		host.Tracef("GetBytes o%d k%d missing key", objID, keyID)
		return nil
	}
	bytes := obj.GetBytes(keyID, typeID)
	switch typeID {
	case OBJTYPE_INT16:
		val16, _, err := codec.DecodeInt16(bytes)
		if err != nil {
			panic("GetBytes: invalid int16")
		}
		host.Tracef("GetBytes o%d k%d = %ds", objID, keyID, val16)
	case OBJTYPE_INT32:
		val32, _, err := codec.DecodeInt32(bytes)
		if err != nil {
			panic("GetBytes: invalid int32")
		}
		host.Tracef("GetBytes o%d k%d = %di", objID, keyID, val32)
	case OBJTYPE_INT64:
		val64, _, err := codec.DecodeInt64(bytes)
		if err != nil {
			panic("GetBytes: invalid int64")
		}
		host.Tracef("GetBytes o%d k%d = %dl", objID, keyID, val64)
	case OBJTYPE_STRING:
		host.Tracef("GetBytes o%d k%d = '%s'", objID, keyID, string(bytes))
	default:
		host.Tracef("GetBytes o%d k%d = '%s'", objID, keyID, base58.Encode(bytes))
	}
	return bytes
}

func (host *KvStoreHost) getKeyFromID(keyID int32) []byte {
	// find predefined key
	if keyID < 0 {
		return host.keyIDToKeyMap[-keyID]
	}

	// find user-defined key
	return host.keyIDToKey[keyID & ^KeyFromBytes]
}

func (host *KvStoreHost) GetKeyFromID(keyID int32) []byte {
	host.TraceAllf("GetKeyFromID(k%d)", keyID)
	key := host.getKeyFromID(keyID)
	if (keyID & (KeyFromBytes | -0x80000000)) == KeyFromBytes {
		// originally a byte slice key
		host.Tracef("GetKeyFromID k%d='%s'", keyID, base58.Encode(key))
		return key
	}
	// originally a string key
	host.Tracef("GetKeyFromID k%d='%s'", keyID, string(key))
	return key
}

func (host *KvStoreHost) getKeyID(key []byte, fromBytes bool) int32 {
	// cannot use []byte as key in maps
	// so we will convert to (non-utf8) string
	// most will have started out as string anyway
	keyString := string(key)
	keyID, ok := host.keyToKeyID[keyString]
	if ok {
		return keyID
	}

	// unknown key, add it to user-defined key map
	keyID = int32(len(host.keyIDToKey))
	if fromBytes {
		keyID |= KeyFromBytes
	}
	host.keyToKeyID[keyString] = keyID
	host.keyIDToKey = append(host.keyIDToKey, key)
	return keyID
}

func (host *KvStoreHost) GetKeyIDFromBytes(bytes []byte) int32 {
	keyID := host.getKeyID(bytes, true)
	host.Tracef("GetKeyIDFromBytes '%s'=k%d", base58.Encode(bytes), keyID)
	return keyID
}

func (host *KvStoreHost) GetKeyIDFromString(key string) int32 {
	keyID := host.getKeyID([]byte(key), false)
	host.Tracef("GetKeyIDFromString '%s'=k%d", key, keyID)
	return keyID
}

func (host *KvStoreHost) GetKeyStringFromID(keyID int32) string {
	return string(host.GetKeyFromID(keyID))
}

func (host *KvStoreHost) GetObjectID(objID, keyID, typeID int32) int32 {
	host.TraceAllf("GetObjectID(o%d,k%d,t%d)", objID, keyID, typeID)
	subID := host.FindObject(objID).GetObjectID(keyID, typeID)
	host.Tracef("GetObjectID o%d k%d t%d = o%d", objID, keyID, typeID, subID)
	return subID
}

func (host *KvStoreHost) PopFrame(frame []HostObject) {
	host.objIDToObj = frame
}

func (host *KvStoreHost) PushFrame() []HostObject {
	// reset frame to contain only null and root object
	// create a fresh slice to allow garbage collection
	// it's up to the caller to save and/or restore the old frame
	pushed := host.objIDToObj
	host.objIDToObj = make([]HostObject, 2, 16)
	copy(host.objIDToObj, pushed[:2])
	return pushed
}

func (host *KvStoreHost) SetBytes(objID, keyID, typeID int32, bytes []byte) {
	host.FindObject(objID).SetBytes(keyID, typeID, bytes)
	switch typeID {
	case OBJTYPE_INT16:
		val16, _, err := codec.DecodeInt16(bytes)
		if err != nil {
			panic("SetBytes: invalid int16")
		}
		host.Tracef("SetBytes o%d k%d v=%ds", objID, keyID, val16)
	case OBJTYPE_INT32:
		val32, _, err := codec.DecodeInt32(bytes)
		if err != nil {
			panic("SetBytes: invalid int32")
		}
		host.Tracef("SetBytes o%d k%d v=%di", objID, keyID, val32)
	case OBJTYPE_INT64:
		val64, _, err := codec.DecodeInt64(bytes)
		if err != nil {
			panic("SetBytes: invalid int64")
		}
		host.Tracef("SetBytes o%d k%d v=%dl", objID, keyID, val64)
	case OBJTYPE_STRING:
		host.Tracef("SetBytes o%d k%d v='%s'", objID, keyID, string(bytes))
	default:
		host.Tracef("SetBytes o%d k%d v='%s'", objID, keyID, base58.Encode(bytes))
	}
}

func (host *KvStoreHost) Tracef(format string, a ...interface{}) {
	if HostTracing {
		host.log.Debugf(format, a...)
	}
}

func (host *KvStoreHost) TraceAllf(format string, a ...interface{}) {
	if ExtendedHostTracing {
		host.Tracef(format, a...)
	}
}

func (host *KvStoreHost) TrackObject(obj HostObject) int32 {
	objID := int32(len(host.objIDToObj))
	host.objIDToObj = append(host.objIDToObj, obj)
	return objID
}
