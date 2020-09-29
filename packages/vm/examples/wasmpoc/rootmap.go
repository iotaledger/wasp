package wasmpoc

import (
	"github.com/iotaledger/wasplib/host/interfaces"
	"github.com/iotaledger/wasplib/host/interfaces/objtype"
)

const (
	KeyAccount     = interfaces.KeyUserDefined
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
	"error":     interfaces.KeyError,
	"length":    interfaces.KeyLength,
	"log":       interfaces.KeyLog,
	"trace":     interfaces.KeyTrace,
	"traceHost": interfaces.KeyTraceHost,

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

type RootObject struct {
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

func NewRootObject(h *wasmVMPocProcessor) interfaces.HostObject {
	return &RootObject{MapObject: MapObject{vm: h, name: "Root"}}
}

func (o *RootObject) GetObjectId(keyId int32, typeId int32) int32 {
	switch keyId {
	case KeyAccount:
		return o.checkedObjectId(&o.accountId, NewAccountMap, typeId, objtype.OBJTYPE_MAP)
	case KeyContract:
		return o.checkedObjectId(&o.contractId, NewContractMap, typeId, objtype.OBJTYPE_MAP)
	case KeyEvents:
		return o.checkedObjectId(&o.eventsId, NewEventsArray, typeId, objtype.OBJTYPE_MAP_ARRAY)
	case KeyLogs:
		return o.checkedObjectId(&o.contractId, NewLogsMap, typeId, objtype.OBJTYPE_MAP)
	case KeyRequest:
		return o.checkedObjectId(&o.requestId, NewRequestMap, typeId, objtype.OBJTYPE_MAP)
	case KeyState:
		return o.checkedObjectId(&o.stateId, NewStateObject, typeId, objtype.OBJTYPE_MAP)
	case KeyTransfers:
		return o.checkedObjectId(&o.transfersId, NewTransfersArray, typeId, objtype.OBJTYPE_MAP_ARRAY)
	case KeyUtility:
		return o.checkedObjectId(&o.utilityId, NewUtilityMap, typeId, objtype.OBJTYPE_MAP)
	}
	return o.MapObject.GetObjectId(keyId, typeId)
}
