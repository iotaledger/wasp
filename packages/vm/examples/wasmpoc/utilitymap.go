package wasmpoc

import (
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
)

type UtilityMap struct {
	MapObject
	hash string
}

func NewUtilityMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &UtilityMap{MapObject: MapObject{vm: h, name: "Utility"}}
}

func (o *UtilityMap) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyRandom:
		//TODO using GetEntropy correctly is painful, so we use tx hash instead
		// we need to be able to get the signature of a specific tx to have
		// deterministic entropy that cannot be interrupted
		txHash := o.vm.ctx.AccessRequest().ID()
		return int64(binary.LittleEndian.Uint64(txHash[10:18]))
	}
	return o.MapObject.GetInt(keyId)
}

func (o *UtilityMap) GetString(keyId int32) string {
	switch keyId {
	case KeyHash:
		return o.hash
	}
	return o.MapObject.GetString(keyId)
}

func (o *UtilityMap) SetString(keyId int32, value string) {
	switch keyId {
	case KeyHash: //TODO hash this!
		o.hash = value
	}
	o.MapObject.SetString(keyId, value)
}
