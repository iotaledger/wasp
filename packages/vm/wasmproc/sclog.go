// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScLogs struct {
	ScSandboxObject
}

func NewScLogs(vm *wasmProcessor) *ScLogs {
	o := &ScLogs{}
	o.vm = vm
	return o
}

func (o *ScLogs) Exists(keyId int32) bool {
	_, ok := o.objects[keyId]
	return ok
}

func (o *ScLogs) GetObjectId(keyId int32, typeId int32) int32 {
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		keyId: func() WaspObject { return NewScLog(o.vm) },
	})
}

func (o *ScLogs) GetTypeId(keyId int32) int32 {
	return wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScLog struct {
	ScSandboxObject
	lines      *datatypes.MustTimestampedLog
	logEntry   *ScLogEntry
	logEntryId int32
}

func NewScLog(vm *wasmProcessor) *ScLog {
	o := &ScLog{}
	o.vm = vm
	return o
}

func (a *ScLog) InitObj(id int32, keyId int32, owner *ScDict) {
	a.ScSandboxObject.InitObj(id, keyId, owner)
	key := a.host.GetKeyFromId(keyId)
	a.lines = datatypes.NewMustTimestampedLog(a.vm.state(), kv.Key(key))
	a.logEntry = &ScLogEntry{lines: a.lines, current: ^uint32(0)}
	a.logEntryId = a.host.TrackObject(a.logEntry)
}

func (a *ScLog) Exists(keyId int32) bool {
	return uint32(keyId) <= a.lines.Len()
}

func (a *ScLog) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyLength:
		return int64(a.lines.Len())
	}
	a.invalidKey(keyId)
	return 0
}

func (a *ScLog) GetObjectId(keyId int32, typeId int32) int32 {
	if typeId != wasmhost.OBJTYPE_MAP {
		a.Panic("GetObjectId: Invalid type")
	}
	index := uint32(keyId)
	if index > a.lines.Len() {
		a.Panic("GetObjectId: Invalid index")
	}
	a.Trace("LoadRecord %s%s", a.name, a.Suffix(keyId))
	a.logEntry.LoadRecord(index)
	return a.logEntryId
}

func (a *ScLog) GetTypeId(keyId int32) int32 {
	if a.Exists(keyId) {
		return wasmhost.OBJTYPE_MAP
	}
	return 0
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScLogEntry struct {
	ScSandboxObject
	current   uint32
	lines     *datatypes.MustTimestampedLog
	record    []byte
	timestamp int64
}

func (o *ScLogEntry) Exists(keyId int32) bool {
	if o.current < o.lines.Len() {
		return o.GetTypeId(keyId) > 0
	}
	return false
}

func (o *ScLogEntry) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyData:
		if o.current < o.lines.Len() {
			return o.record[8:]
		}
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScLogEntry) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyTimestamp:
		if o.current < o.lines.Len() {
			return int64(util.MustUint64From8Bytes(o.record[:8]))
		}
	}
	o.invalidKey(keyId)
	return 0
}

func (o *ScLogEntry) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyData:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyTimestamp:
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

func (o *ScLogEntry) LoadRecord(index uint32) {
	if index != o.current {
		o.current = index
		if index < o.lines.Len() {
			o.record = o.lines.LoadRecordsRaw(index, index, false)[0]
		}
	}
}

func (o *ScLogEntry) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case wasmhost.KeyData:
		// can only append
		if o.current != o.lines.Len() {
			o.Panic("SetBytes: Invalid log append index: %d", keyId)
		}
		o.lines.Append(o.timestamp, value)
		o.current = ^uint32(0)
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScLogEntry) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyTimestamp:
		// can only append
		if o.current != o.lines.Len() {
			o.Panic("SetInt: Invalid log append index: %d", keyId)
		}
		o.timestamp = value
	default:
		o.invalidKey(keyId)
	}
}
