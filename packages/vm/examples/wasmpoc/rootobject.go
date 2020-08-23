package wasmpoc

import (
	"encoding/binary"
	"github.com/iotaledger/wart/host/interfaces"
	"github.com/iotaledger/wart/host/interfaces/objtype"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

const (
	KeyBalance     = interfaces.KeyUserDefined
	KeyConfig      = KeyBalance - 1
	KeyOwner       = KeyConfig - 1
	KeyParams      = KeyOwner - 1
	KeyRandom      = KeyParams - 1
	KeyReqAddress  = KeyRandom - 1
	KeyReqBalance  = KeyReqAddress - 1
	KeyReqCode     = KeyReqBalance - 1
	KeyReqDelay    = KeyReqCode - 1
	KeyReqHash     = KeyReqDelay - 1
	KeyRequests    = KeyReqHash - 1
	KeyScAddress   = KeyRequests - 1
	KeySender      = KeyScAddress - 1
	KeyState       = KeySender - 1
	KeyTimestamp   = KeyState - 1
	KeyTransfers   = KeyTimestamp - 1
	KeyXferAddress = KeyTransfers - 1
	KeyXferAmount  = KeyXferAddress - 1
	KeyXferColor   = KeyXferAmount - 1
)

var keyMap = map[string]int32{
	// predefined keys
	"error":     interfaces.KeyError,
	"length":    interfaces.KeyLength,
	"log":       interfaces.KeyLog,
	"trace":     interfaces.KeyTrace,
	"traceHost": interfaces.KeyTraceHost,

	// user-defined keys
	"balance":     KeyBalance,
	"config":      KeyConfig,
	"owner":       KeyOwner,
	"params":      KeyParams,
	"random":      KeyRandom,
	"reqAddress":  KeyReqAddress,
	"reqBalance":  KeyReqBalance,
	"reqCode":     KeyReqCode,
	"reqDelay":    KeyReqDelay,
	"reqHash":     KeyReqHash,
	"requests":    KeyRequests,
	"scAddress":   KeyScAddress,
	"sender":      KeySender,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"transfers":   KeyTransfers,
	"xferAddress": KeyXferAddress,
	"xferAmount":  KeyXferAmount,
	"xferColor":   KeyXferColor,
}

type RootObject struct {
	vm      *wasmVMPocProcessor
	objects map[int32]int32
	types   map[int32]int32
}

func NewRootObject(h *wasmVMPocProcessor, ctx vmtypes.Sandbox) *RootObject {
	h.ctx = ctx
	h.Init(h, &keyMap)
	o := &RootObject{vm: h, objects: make(map[int32]int32), types: make(map[int32]int32)}
	h.AddObject(o)

	o.objects[KeyBalance] = h.AddObject(NewBalanceObject(h, false))
	o.types[KeyBalance] = objtype.OBJTYPE_MAP

	o.objects[KeyConfig] = h.AddObject(NewConfigObject(h))
	o.types[KeyConfig] = objtype.OBJTYPE_MAP

	o.objects[KeyParams] = h.AddObject(NewParamsObject(h))
	o.types[KeyParams] = objtype.OBJTYPE_MAP

	o.objects[KeyReqBalance] = h.AddObject(NewBalanceObject(h, true))
	o.types[KeyReqBalance] = objtype.OBJTYPE_MAP

	o.objects[KeyRequests] = h.AddObject(NewRequestsArray(h))
	o.types[KeyRequests] = objtype.OBJTYPE_MAP_ARRAY

	o.objects[KeyState] = h.AddObject(NewStateObject(h))
	o.types[KeyState] = objtype.OBJTYPE_MAP

	o.objects[KeyTransfers] = h.AddObject(NewTransfersArray(h))
	o.types[KeyTransfers] = objtype.OBJTYPE_MAP_ARRAY

	return o
}

func (o *RootObject) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyRandom:
		//TODO using GetEntropy is painful, so we use tx hash instead
		// we need to be able to get the signature of a specific tx to
		// have deterministic entropy that cannot be interrupted
		hash := o.vm.ctx.AccessRequest().ID()
		return int64(binary.LittleEndian.Uint64(hash[10:18]))
	case KeyTimestamp:
		return o.vm.ctx.GetTimestamp()
	default:
		o.vm.SetError("Invalid key")
	}
	return 0
}

func (o *RootObject) GetObjectId(keyId int32, typeId int32) int32 {
	if !o.valid(keyId, typeId) {
		return 0
	}
	objId, ok := o.objects[keyId]
	if ok {
		return objId
	}
	o.vm.SetError("Invalid key")
	return 0
}

func (o *RootObject) GetString(keyId int32) string {
	switch keyId {
	case KeyOwner:
		return o.vm.ctx.GetOwnerAddress().String()
	case KeyReqHash:
		id := o.vm.ctx.AccessRequest().ID()
		return id.String()
	case KeyScAddress:
		return o.vm.ctx.GetSCAddress().String()
	case KeySender:
		return o.vm.ctx.AccessRequest().Sender().String()
	default:
		o.vm.SetError("Invalid key")
	}
	return ""
}

func (o *RootObject) SetInt(keyId int32, value int64) {
	switch keyId {
	default:
		o.vm.SetError("Invalid key")
	}
}

func (o *RootObject) SetString(keyId int32, value string) {
	switch keyId {
	default:
		o.vm.SetError("Invalid key")
	}
}

func (o *RootObject) valid(keyId int32, typeId int32) bool {
	fieldType, ok := o.types[keyId]
	if !ok {
		//if o.readonly {
		//	o.vm.SetError("Readonly")
		//	return false
		//}
		o.types[keyId] = typeId
		return true
	}
	if fieldType != typeId {
		o.vm.SetError("Invalid access")
		return false
	}
	return true
}
