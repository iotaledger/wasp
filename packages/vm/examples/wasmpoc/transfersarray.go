package wasmpoc

import "github.com/iotaledger/wart/host/interfaces"

type TransfersArray struct {
	vm        *wasmVMPocProcessor
	transfers []int32
}

func NewTransfersArray(h *wasmVMPocProcessor) *TransfersArray {
	return &TransfersArray{vm: h}
}

func (a *TransfersArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(len(a.transfers))
	}
	a.vm.SetError("Invalid access")
	return 0
}

func (a *TransfersArray) GetLength() int32 {
	return int32(len(a.transfers))
}

func (a *TransfersArray) GetObjectId(keyId int32, typeId int32) int32 {
	length := a.GetLength()
	if keyId < 0 || keyId > length {
		a.vm.SetError("Invalid index")
		return 0
	}
	if keyId < length {
		return a.transfers[keyId]
	}
	objId := a.vm.AddObject(NewTransferObject(a.vm))
	a.transfers = append(a.transfers, objId)
	return objId
}

func (a *TransfersArray) GetString(keyId int32) string {
	a.vm.SetError("Invalid access")
	return ""
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
		return
	}
	a.vm.SetError("Invalid access")
}

func (a *TransfersArray) SetString(keyId int32, value string) {
	a.vm.SetError("Invalid access")
}
