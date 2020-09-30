package wasmhost

type EventsArray struct {
	ArrayObject
	events []int32
}

func NewEventsArray(vm *wasmVMPocProcessor) HostObject {
	return &EventsArray{ArrayObject: ArrayObject{vm: vm, name: "Events"}}
}

func (a *EventsArray) GetInt(keyId int32) int64 {
	switch keyId {
	case KeyLength:
		return int64(a.GetLength())
	}
	return a.ArrayObject.GetInt(keyId)
}

func (a *EventsArray) GetLength() int32 {
	return int32(len(a.events))
}

func (a *EventsArray) GetObjectId(keyId int32, typeId int32) int32 {
	return a.checkedObjectId(&a.events, keyId, NewEventMap, typeId, OBJTYPE_MAP)
}

func (a *EventsArray) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		for i := a.GetLength() - 1; i >= 0; i-- {
			a.vm.SetInt(a.events[i], keyId, 0)
		}
		//todo move to pool for reuse of events?
		a.events = nil
		return
	default:
		a.ArrayObject.SetInt(keyId, value)
	}
}
