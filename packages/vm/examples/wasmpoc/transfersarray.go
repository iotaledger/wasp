package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type TransfersArray struct {
	ArrayObject
	transfers []int32
}

func NewTransfersArray(h *wasmVMPocProcessor) interfaces.HostObject {
	return &TransfersArray{ArrayObject: ArrayObject{vm: h, name: "Transfers"}}
}

func (a *TransfersArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(a.GetLength())
	}
	return a.ArrayObject.GetInt(keyId)
}

func (a *TransfersArray) GetLength() int32 {
	return int32(len(a.transfers))
}

func (a *TransfersArray) GetObjectId(keyId int32, typeId int32) int32 {
	return a.checkedObjectId(&a.transfers, keyId, NewTransferMap, typeId, objtype.OBJTYPE_MAP)
}

func (a *TransfersArray) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		// tell objects to clear themselves
		for i := len(a.transfers) - 1; i >= 0; i-- {
			a.vm.SetInt(a.transfers[i], keyId, 0)
		}
		//TODO move to pool for reuse of transfers
		a.transfers = nil
	default:
		a.error("SetInt: Invalid access")
	}
}

func (a *TransfersArray) SetString(keyId int32, value string) {
	a.error("SetString: Invalid access")
}
