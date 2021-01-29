// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/mr-tron/base58"
)

var TestMode = false

type ScUtility struct {
	ScSandboxObject
	base58Decoded []byte
	base58Encoded string
	hash          hashing.HashValue
	hname         coretypes.Hname
	nextRandom    int
	random        []byte
}

func NewScUtility(vm *wasmProcessor) *ScUtility {
	o := &ScUtility{}
	o.vm = vm
	return o
}

func (o *ScUtility) InitObj(id int32, keyId int32, owner *ScDict) {
	o.ScSandboxObject.InitObj(id, keyId, owner)
	if TestMode {
		// preset randomizer to generate sequence 1..8 before
		// continuing with proper hashed values
		o.random = make([]byte, 8*8)
		for i := 0; i < len(o.random); i += 8 {
			o.random[i] = byte(i + 1)
		}
	}
}

func (o *ScUtility) Exists(keyId int32, typeId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScUtility) GetBytes(keyId int32, typeId int32) []byte {
	switch keyId {
	case wasmhost.KeyBase58Bytes:
		return o.base58Decoded
	case wasmhost.KeyBase58String:
		return []byte(o.base58Encoded)
	case wasmhost.KeyHash:
		return o.hash.Bytes()
	case wasmhost.KeyHname:
		return codec.EncodeHname(o.hname)
	case wasmhost.KeyRandom:
		return o.getRandom8Bytes()
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScUtility) getRandom8Bytes() []byte {
	if o.random == nil {
		// need to initialize pseudo-random generator with
		// a sufficiently random, yet deterministic, value
		id := o.vm.ctx.GetEntropy()
		o.random = id[:]
	}
	i := o.nextRandom
	if i+8 > len(o.random) {
		// not enough bytes left, generate more bytes
		h := hashing.HashData(o.random)
		o.random = h[:]
		i = 0
	}
	o.nextRandom = i + 8
	return o.random[i : i+8]
}

func (o *ScUtility) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyBase58Bytes:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyBase58String:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyHash:
		return wasmhost.OBJTYPE_HASH
	case wasmhost.KeyHname:
		return wasmhost.OBJTYPE_HNAME
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyRandom:
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

func (o *ScUtility) SetBytes(keyId int32, typeId int32, bytes []byte) {
	var err error
	switch keyId {
	case wasmhost.KeyBase58Bytes:
		o.base58Encoded = base58.Encode(bytes)
	case wasmhost.KeyBase58String:
		o.base58Decoded, err = base58.Decode(string(bytes))
	case wasmhost.KeyHash:
		o.hash = hashing.HashData(bytes)
	case wasmhost.KeyName:
		o.hname = coretypes.Hn(string(bytes))
	default:
		o.invalidKey(keyId)
	}
	if err != nil {
		o.Panic(err.Error())
	}
}
