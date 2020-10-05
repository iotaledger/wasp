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
}

func NewScContext(vm *wasmProcessor) HostObject {
	return &ScContext{MapObject: MapObject{ModelObject: ModelObject{vm: vm, name: "Root"}, objects: make(map[int32]int32)}}
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	return o.GetMapObjectId(keyId, typeId, map[int32]MapObjDesc{
		KeyAccount:   {OBJTYPE_MAP, func() WaspObject { return &ScAccount{} }},
		KeyContract:  {OBJTYPE_MAP, func() WaspObject { return &ScContract{} }},
		KeyEvents:    {OBJTYPE_MAP_ARRAY, func() WaspObject { return &ScEvents{} }},
		KeyLogs:      {OBJTYPE_MAP, func() WaspObject { return &ScLogs{} }},
		KeyRequest:   {OBJTYPE_MAP, func() WaspObject { return &ScRequest{} }},
		KeyState:     {OBJTYPE_MAP, func() WaspObject { return &ScState{} }},
		KeyTransfers: {OBJTYPE_MAP_ARRAY, func() WaspObject { return &ScTransfers{} }},
		KeyUtility:   {OBJTYPE_MAP, func() WaspObject { return &ScUtility{} }},
	})
}
