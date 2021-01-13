// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/mr-tron/base58"
)

var TestMode = false

type ScUtility struct {
	ScSandboxObject
	base58Decoded []byte
	base58Encoded string
	hash          []byte
	random        []byte
	nextRandom    int
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

func (o *ScUtility) Exists(keyId int32) bool {
	switch keyId {
	case wasmhost.KeyBase58:
	case wasmhost.KeyHash:
	case wasmhost.KeyRandom:
	default:
		return false
	}
	return true
}

func (o *ScUtility) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyBase58:
		return o.base58Decoded
	case wasmhost.KeyHash:
		return o.hash
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScUtility) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyRandom:
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
		return int64(binary.LittleEndian.Uint64(o.random[i : i+8]))
	}
	o.invalidKey(keyId)
	return 0
}

func (o *ScUtility) GetString(keyId int32) string {
	switch keyId {
	case wasmhost.KeyBase58:
		return o.base58Encoded
	}
	o.invalidKey(keyId)
	return ""
}

func (o *ScUtility) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyBase58:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyHash:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_HASH
	case wasmhost.KeyRandom:
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

func (o *ScUtility) SetBytes(keyId int32, value []byte) {
	switch keyId {
	case wasmhost.KeyBase58:
		o.base58Encoded = base58.Encode(value)
	case wasmhost.KeyHash:
		h := hashing.HashData(value)
		o.hash = h[:]
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScUtility) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyBase58:
		o.base58Decoded, _ = base58.Decode(value)
	default:
		o.invalidKey(keyId)
	}
}
