// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

type ScState struct {
	ScMutableDict
	isArray     bool
	arrayTypeId int32
}

func (o *ScState) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.ScMutableDict.InitObj(id, keyId, owner)
	o.Dict = o.vm.State()
	if o.isArray {
		o.typeId = o.arrayTypeId
	}
	o.nested = true
}

func (o *ScState) GetObjectId(keyId int32, typeId int32) int32 {
	o.validate(keyId, typeId)
	var factory ObjFactory
	if typeId == OBJTYPE_MAP {
		factory = func() WaspObject { return &ScState{} }
	} else {
		if (typeId & OBJTYPE_ARRAY) != 0 {
			factory = func() WaspObject { return &ScState{isArray: true, arrayTypeId: typeId & ^OBJTYPE_ARRAY} }
		} else {
			o.Panic("GetObjectId: Invalid type")
		}
	}
	return GetMapObjectId(o, keyId, typeId, ObjFactories{
		keyId: factory,
	})
}

func (o *ScState) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		//TODO how to clear state?
		o.Panic("SetInt: Clear state")
	default:
		o.ScMutableDict.SetInt(keyId, value)
	}
}
