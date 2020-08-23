package wasmpoc

import "github.com/iotaledger/wart/host/interfaces"

type RequestsArray struct {
	vm       *wasmVMPocProcessor
	requests []int32
}

func NewRequestsArray(h *wasmVMPocProcessor) *RequestsArray {
	return &RequestsArray{vm: h}
}

func (a *RequestsArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(len(a.requests))
	}
	a.vm.SetError("Invalid access")
	return 0
}

func (a *RequestsArray) GetLength() int32 {
	return int32(len(a.requests))
}

func (a *RequestsArray) GetObjectId(keyId int32, typeId int32) int32 {
	length := a.GetLength()
	if keyId < 0 || keyId > length {
		a.vm.SetError("Invalid index")
		return 0
	}
	if keyId < length {
		return a.requests[keyId]
	}
	obj := NewRequestObject(a.vm)
	objId := a.vm.AddObject(obj)
	a.requests = append(a.requests, objId)
	return objId
}

func (a *RequestsArray) GetString(keyId int32) string {
	a.vm.SetError("Invalid access")
	return ""
}

func (a *RequestsArray) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		for i := len(a.requests) - 1; i >= 0; i-- {
			a.vm.SetInt(a.requests[i], keyId, 0)
		}
		//todo move to pool for reuse of requests
		a.requests = nil
		return
	}
	a.vm.SetError("Invalid access")
}

func (a *RequestsArray) SetString(keyId int32, value string) {
	a.vm.SetError("Invalid access")
}
