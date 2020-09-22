package wasmpoc

import (
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces"
	"github.com/iotaledger/wasp/packages/vm/examples/wasmpoc/wasplib/host/interfaces/objtype"
)

type EventsArray struct {
	ArrayObject
	events []int32
}

func NewEventsArray(h *wasmVMPocProcessor) interfaces.HostObject {
	return &EventsArray{ArrayObject: ArrayObject{vm: h, name: "Events"}}
}

func (a *EventsArray) GetInt(keyId int32) int64 {
	switch keyId {
	case interfaces.KeyLength:
		return int64(a.GetLength())
	}
	return a.GetInt(keyId)
}

func (a *EventsArray) GetLength() int32 {
	return int32(len(a.events))
}

func (a *EventsArray) GetObjectId(keyId int32, typeId int32) int32 {
	return a.checkedObjectId(&a.events, keyId, NewEventMap, typeId, objtype.OBJTYPE_MAP)
}

func (a *EventsArray) SetInt(keyId int32, value int64) {
	switch keyId {
	case interfaces.KeyLength:
		for i := len(a.events) - 1; i >= 0; i-- {
			a.vm.SetInt(a.events[i], keyId, 0)
		}
		//todo move to pool for reuse of events?
		a.events = nil
		return
	default:
		a.error("SetInt: Invalid access")
	}
}

func (a *EventsArray) SetString(keyId int32, value string) {
	a.error("SetString: Invalid access")
}
