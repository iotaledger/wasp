// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type InitHost struct {
	funcs  []ScFuncContextFunction
	views  []ScViewContextFunction
	params map[int32][]byte
}

var _ ScHost = &InitHost{}

func NewInitHost() *InitHost {
	return &InitHost{params: make(map[int32][]byte)}
}

func (h InitHost) AddFunc(f ScFuncContextFunction) []ScFuncContextFunction {
	if f != nil {
		h.funcs = append(h.funcs, f)
	}
	return h.funcs
}

func (h InitHost) AddView(v ScViewContextFunction) []ScViewContextFunction {
	if v != nil {
		h.views = append(h.views, v)
	}
	return h.views
}

func (h InitHost) CallFunc(objID, keyID int32, params []byte) []byte {
	Panic("InitHost::CallFunc")
	return nil
}

func (h InitHost) Exists(objID, keyID, typeID int32) bool {
	if objID == int32(KeyParams) {
		_, exists := h.params[keyID]
		return exists
	}
	Panic("InitHost::Exists")
	return false
}

func (h InitHost) GetBytes(objID, keyID, typeID int32) []byte {
	if objID == int32(KeyMaps) && keyID == int32(KeyLength) {
		return nil
	}
	if objID == int32(KeyParams) {
		return h.params[keyID]
	}
	Panic("InitHost::GetBytes")
	return nil
}

func (h InitHost) GetKeyIDFromBytes(bytes []byte) int32 {
	Panic("InitHost::GetKeyIDFromBytes")
	return 0
}

func (h InitHost) GetKeyIDFromString(key string) int32 {
	Panic("InitHost::GetKeyIDFromString")
	return 0
}

func (h InitHost) GetObjectID(objID, keyID, typeID int32) int32 {
	if objID == 1 && keyID == int32(KeyMaps) {
		return keyID
	}
	if objID == int32(KeyMaps) && keyID == 0 {
		return int32(KeyParams)
	}
	Panic("InitHost::GetObjectID")
	return 0
}

func (h InitHost) SetBytes(objID, keyID, typeID int32, value []byte) {
	if objID == 1 && keyID == int32(KeyPanic) {
		panic(string(value))
	}
	if objID == int32(KeyParams) {
		h.params[keyID] = value
		return
	}
	Panic("InitHost::SetBytes")
}
