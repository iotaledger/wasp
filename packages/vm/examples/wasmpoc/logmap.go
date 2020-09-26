package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasplib/host/interfaces"
)

type LogMap struct {
	MapObject
	lines *kv.MustTimestampedLog
	timestamp int64
}

func NewLogMap(h *wasmVMPocProcessor, a *kv.MustTimestampedLog ) interfaces.HostObject {
	return &LogMap{MapObject: MapObject{vm: h, name: "LogMap"}, lines: a}
}

func (a *LogMap) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(a.lines.Len())
	}
	return a.MapObject.GetInt(keyId)
}

func (a *LogMap) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyData:
		a.lines.Append(a.timestamp, value)
		return
	}
	a.error("SetBytes: Invalid key")
}

func (a *LogMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyTimestamp:
		a.timestamp = value
		return
	}
	a.error("SetInt: Invalid key")
}
