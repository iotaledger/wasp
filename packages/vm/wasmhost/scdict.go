// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScImmutableDict struct {
	MapObject
	Dict dict.Dict
}

func (o *ScImmutableDict) InitObj(id int32, keyId int32, owner *ModelObject) {
	o.MapObject.InitObj(id, keyId, owner)
	if o.Dict == nil {
		o.Dict = dict.New()
	}
}

func (o *ScImmutableDict) Exists(keyId int32) bool {
	key := o.vm.GetKey(keyId)
	exists, _ := o.Dict.Has(key)
	return exists
}

func (o *ScImmutableDict) GetBytes(keyId int32) []byte {
	key := o.vm.GetKey(keyId)
	value, _ := o.Dict.Get(key)
	return value
}

func (o *ScImmutableDict) GetInt(keyId int32) int64 {
	key := o.vm.GetKey(keyId)
	value, _, _ := codec.DecodeInt64(o.Dict.MustGet(key))
	return value
}

func (o *ScImmutableDict) GetString(keyId int32) string {
	key := o.vm.GetKey(keyId)
	value, _, _ := codec.DecodeString(o.Dict.MustGet(key))
	return value
}

//TODO keep track of field types
func (o *ScImmutableDict) GetTypeId(keyId int32) int32 {
	return o.MapObject.GetTypeId(keyId)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScMutableDict struct {
	ScImmutableDict
}

func (o *ScMutableDict) SetBytes(keyId int32, value []byte) {
	key := o.vm.GetKey(keyId)
	o.Dict.Set(key, value)
}

func (o *ScMutableDict) SetInt(keyId int32, value int64) {
	switch keyId {
	case KeyLength:
		o.Dict = dict.New()
	default:
		key := o.vm.GetKey(keyId)
		o.Dict.Set(key, codec.EncodeInt64(value))
	}
}

func (o *ScMutableDict) SetString(keyId int32, value string) {
	key := o.vm.GetKey(keyId)
	o.Dict.Set(key, codec.EncodeString(value))
}
