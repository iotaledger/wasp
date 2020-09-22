package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
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

func (a *LogMap) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyTimestamp:
		a.timestamp = value
		return
	}
	a.error("SetInt: Invalid key")
}

func (a *LogMap) SetString(keyId int32, value string) {
	switch keyId {
	case KeyData:
		a.lines.Append(a.timestamp, []byte(value))
		return
	}
	a.error("SetString: Invalid key")
}
