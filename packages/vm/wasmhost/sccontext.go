package wasmhost

const (
	KeyAccount     = KeyUserDefined
	KeyAgent       = KeyAccount - 1
	KeyAmount      = KeyAgent - 1
	KeyBalance     = KeyAmount - 1
	KeyBase58      = KeyBalance - 1
	KeyCalls       = KeyBase58 - 1
	KeyColor       = KeyCalls - 1
	KeyColors      = KeyColor - 1
	KeyContract    = KeyColors - 1
	KeyData        = KeyContract - 1
	KeyDelay       = KeyData - 1
	KeyDescription = KeyDelay - 1
	KeyExports     = KeyDescription - 1
	KeyFunction    = KeyExports - 1
	KeyHash        = KeyFunction - 1
	KeyId          = KeyHash - 1
	KeyIota        = KeyId - 1
	KeyLogs        = KeyIota - 1
	KeyName        = KeyLogs - 1
	KeyOwner       = KeyName - 1
	KeyParams      = KeyOwner - 1
	KeyPosts       = KeyParams - 1
	KeyRandom      = KeyPosts - 1
	KeyRequest     = KeyRandom - 1
	KeyResults     = KeyRequest - 1
	KeySender      = KeyResults - 1
	KeyState       = KeySender - 1
	KeyTimestamp   = KeyState - 1
	KeyTransfers   = KeyTimestamp - 1
	KeyUtility     = KeyTransfers - 1
	KeyViews       = KeyUtility - 1
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
	"agent":       KeyAgent,
	"amount":      KeyAmount,
	"balance":     KeyBalance,
	"base58":      KeyBase58,
	"calls":       KeyCalls,
	"color":       KeyColor,
	"colors":      KeyColors,
	"contract":    KeyContract,
	"data":        KeyData,
	"delay":       KeyDelay,
	"description": KeyDescription,
	"exports":     KeyExports,
	"function":    KeyFunction,
	"hash":        KeyHash,
	"id":          KeyId,
	"iota":        KeyIota,
	"logs":        KeyLogs,
	"name":        KeyName,
	"owner":       KeyOwner,
	"params":      KeyParams,
	"posts":       KeyPosts,
	"random":      KeyRandom,
	"request":     KeyRequest,
	"results":     KeyResults,
	"sender":      KeySender,
	"state":       KeyState,
	"timestamp":   KeyTimestamp,
	"transfers":   KeyTransfers,
	"utility":     KeyUtility,
	"views":       KeyViews,
}

type ScContext struct {
	MapObject
}

func NewScContext(vm *wasmProcessor) *ScContext {
	return &ScContext{MapObject: MapObject{ModelObject: ModelObject{vm: vm, name: "Root"}, objects: make(map[int32]int32)}}
}

func (o *ScContext) Exists(keyId int32) bool {
	if keyId == KeyExports {
		return o.vm.ctx == nil
	}
	return o.GetTypeId(keyId) >= 0
}

func (o *ScContext) Finalize() {
	o.objects = make(map[int32]int32)
	o.vm.objIdToObj = o.vm.objIdToObj[:2]
}

func (o *ScContext) GetObjectId(keyId int32, typeId int32) int32 {
	if keyId == KeyExports && o.vm.ctx != nil {
		// once map has entries (onLoad) this cannot be called any more
		return o.MapObject.GetObjectId(keyId, typeId)
	}

	return GetMapObjectId(o, keyId, typeId, MapFactories{
		KeyAccount:   func() WaspObject { return &ScAccount{} },
		KeyCalls:     func() WaspObject { return &ScPosts{} },
		KeyContract:  func() WaspObject { return &ScContract{} },
		KeyExports:   func() WaspObject { return &ScExports{} },
		KeyLogs:      func() WaspObject { return &ScLogs{} },
		KeyPosts:     func() WaspObject { return &ScPosts{} },
		KeyRequest:   func() WaspObject { return &ScRequest{} },
		KeyState:     func() WaspObject { return &ScState{} },
		KeyTransfers: func() WaspObject { return &ScTransfers{} },
		KeyUtility:   func() WaspObject { return &ScUtility{} },
		KeyViews:     func() WaspObject { return &ScPosts{} },
	})
}

func (o *ScContext) GetTypeId(keyId int32) int32 {
	switch keyId {
	case KeyAccount:
		return OBJTYPE_MAP
	case KeyCalls:
		return OBJTYPE_MAP_ARRAY
	case KeyContract:
		return OBJTYPE_MAP
	case KeyExports:
		return OBJTYPE_STRING_ARRAY
	case KeyLogs:
		return OBJTYPE_MAP
	case KeyPosts:
		return OBJTYPE_MAP_ARRAY
	case KeyRequest:
		return OBJTYPE_MAP
	case KeyState:
		return OBJTYPE_MAP
	case KeyTransfers:
		return OBJTYPE_MAP_ARRAY
	case KeyUtility:
		return OBJTYPE_MAP
	case KeyViews:
		return OBJTYPE_MAP_ARRAY
	}
	return -1
}
