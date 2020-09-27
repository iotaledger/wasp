package wasmpoc

import (
	"encoding/binary"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasplib/host/interfaces"
)

type UtilityMap struct {
	MapObject
	hash       []byte
	random     []byte
	nextRandom int
}

func NewUtilityMap(h *wasmVMPocProcessor) interfaces.HostObject {
	return &UtilityMap{MapObject: MapObject{vm: h, name: "Utility"}}
}

func (o *UtilityMap) GetBytes(keyId int32) []byte {
	switch keyId {
	case KeyHash:
		return o.hash
	}
	return o.MapObject.GetBytes(keyId)
}

func (o *UtilityMap) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyRandom:
		//TODO using GetEntropy correctly is painful, so we use tx hash instead
		// we need to be able to get the signature of a specific tx to have
		// deterministic entropy that cannot be interrupted
		if o.random == nil {
			id := o.vm.ctx.AccessRequest().ID()
			o.random = id.TransactionId().Bytes()
		}
		i := o.nextRandom
		if i == transaction.IDLength {
			o.random = hashing.HashData(o.random).Bytes()
			i = 0
		}
		o.nextRandom = i + 8
		return int64(binary.LittleEndian.Uint64(o.random[i : i+8]))
	}
	return o.MapObject.GetInt(keyId)
}

func (o *UtilityMap) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case KeyHash:
		o.hash = hashing.HashData(value).Bytes()
	}
	o.MapObject.SetBytes(keyId, value)
}
