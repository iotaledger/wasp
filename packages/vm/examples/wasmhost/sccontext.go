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
	accountId   int32
	contractId  int32
	logsId      int32
	requestId   int32
	stateId     int32
	transfersId int32
	eventsId    int32
	utilityId   int32
}

func NewScContext(vm *wasmProcessor) HostObject {
	return &ScContext{MapObject: MapObject{vm: vm, name: "Root"}}
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyAccount:
		return o.checkedObjectId(&o.accountId, NewScAccount, typeId, OBJTYPE_MAP)
	case KeyContract:
		return o.checkedObjectId(&o.contractId, NewScContract, typeId, OBJTYPE_MAP)
	case KeyEvents:
		return o.checkedObjectId(&o.eventsId, NewScEvents, typeId, OBJTYPE_MAP_ARRAY)
	case KeyLogs:
		return o.checkedObjectId(&o.contractId, NewLogsMap, typeId, OBJTYPE_MAP)
	case KeyRequest:
		return o.checkedObjectId(&o.requestId, NewScRequest, typeId, OBJTYPE_MAP)
	case KeyState:
		return o.checkedObjectId(&o.stateId, NewScState, typeId, OBJTYPE_MAP)
	case KeyTransfers:
		return o.checkedObjectId(&o.transfersId, NewScTransfers, typeId, OBJTYPE_MAP_ARRAY)
	case KeyUtility:
		return o.checkedObjectId(&o.utilityId, NewScUtility, typeId, OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}
