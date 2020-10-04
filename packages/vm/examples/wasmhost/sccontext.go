package wasmhost

const (
	KeyAccount     = KeyUserDefined
	KeyAddress     = KeyAccount - 1
	KeyAmount      = KeyAddress - 1
	KeyBalance     = KeyAmount - 1
	KeyCode        = KeyBalance - 1
	KeyColor       = KeyCode - 1
	KeyColors      = KeyColor - 1
	KeyContract    = KeyColors - 1
	KeyData        = KeyContract - 1
	KeyDelay       = KeyData - 1
	KeyDescription = KeyDelay - 1
	KeyEvents      = KeyDescription - 1
	KeyFunction    = KeyEvents - 1
	KeyHash        = KeyFunction - 1
	KeyId          = KeyHash - 1
	KeyIota        = KeyId - 1
	KeyLogs        = KeyIota - 1
	KeyName        = KeyLogs - 1
	KeyOwner       = KeyName - 1
	KeyParams      = KeyOwner - 1
	KeyRandom      = KeyParams - 1
	KeyRequest     = KeyRandom - 1
	KeyState       = KeyRequest - 1
	KeyTimestamp   = KeyState - 1
	KeyTransfers   = KeyTimestamp - 1
	KeyUtility     = KeyTransfers - 1
)

var keyMap = map[string]int32{
	// predefined keys
	"error":     KeyError,
	"length":    KeyLength,
	"log":       KeyLog,
	"trace":     KeyTrace,
	"traceHost": KeyTraceHost,
	"warning":   KeyWarning,

	// user-defined keys
	"account":     KeyAccount,
	"address":     KeyAddress,
	"amount":      KeyAmount,
	"balance":     KeyBalance,
	"code":        KeyCode,
	"color":       KeyColor,
	"colors":      KeyColors,
	"contract":    KeyContract,
	"data":        KeyData,
	"delay":       KeyDelay,
	"description": KeyDescription,
	"events":      KeyEvents,
	"function":    KeyFunction,
	"hash":        KeyHash,
	"id":          KeyId,
	"iota":        KeyIota,
	"logs":        KeyLogs,
	"name":        KeyName,
	"owner":       KeyOwner,
	"params":      KeyParams,
	"random":      KeyRandom,
	"request":     KeyRequest,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"transfers":   KeyTransfers,
	"utility":     KeyUtility,
}

type ScContext struct {
	MapObject
	root map[int32]int32
}

func NewScContext(vm *wasmProcessor) HostObject {
	return &ScContext{MapObject: MapObject{vm: vm, name: "Root"}, root: make(map[int32]int32)}
}

type initVM interface {
	HostObject
	InitVM(vm *wasmProcessor, keyId int32)
}

func (o *ScContext) checkModelObjectId(keyId int32, newObject initVM, typeId int32, expectedTypeId int32) int32 {
	if typeId != expectedTypeId {
		o.error("GetObjectId: Invalid type")
		return 0
	}
	objId, ok := o.root[keyId]
	if ok {
		return objId
	}
	newObject.InitVM(o.vm, keyId)
	objId = o.vm.TrackObject(newObject)
	return objId
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyAccount:
		return o.checkModelObjectId(keyId, &ScAccount{}, typeId, OBJTYPE_MAP)
	case KeyContract:
		return o.checkModelObjectId(keyId, &ScContract{}, typeId, OBJTYPE_MAP)
	case KeyEvents:
		return o.checkModelObjectId(keyId, &ScEvents{}, typeId, OBJTYPE_MAP_ARRAY)
	case KeyLogs:
		return o.checkModelObjectId(keyId, &LogsMap{}, typeId, OBJTYPE_MAP)
	case KeyRequest:
		return o.checkModelObjectId(keyId, &ScRequest{}, typeId, OBJTYPE_MAP)
	case KeyState:
		return o.checkModelObjectId(keyId, &ScState{}, typeId, OBJTYPE_MAP)
	case KeyTransfers:
		return o.checkModelObjectId(keyId, &ScTransfers{}, typeId, OBJTYPE_MAP_ARRAY)
	case KeyUtility:
		return o.checkModelObjectId(keyId, &ScUtility{}, typeId, OBJTYPE_MAP)
	default:
		return o.MapObject.GetObjectId(keyId, typeId)
	}
}
