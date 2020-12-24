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
	ScDict
	base58Decoded []byte
	base58Encoded string
	hash          []byte
	random        []byte
	nextRandom    int
}

func (o *ScUtility) InitObj(id int32, keyId int32, owner *ScDict) {
	o.ScDict.InitObj(id, keyId, owner)
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
	return o.ScDict.GetBytes(keyId)
}

func (o *ScUtility) GetInt(keyId int32) int64 {
	switch keyId {
	case wasmhost.KeyRandom:
		//TODO using GetEntropy correctly is painful, so we use tx hash instead
		// we need to be able to get the signature of a specific tx to have
		// deterministic entropy that cannot be interrupted
		if o.random == nil {
			// need to initialize pseudo-random generator with
			// a sufficiently random, yet deterministic, value
			id := o.vm.ctx.RequestID()
			o.random = id.TransactionID().Bytes()
		}
		i := o.nextRandom
		if i+8 > len(o.random) {
			// not enough bytes left, generate more bytes
			o.random = hashing.HashData(o.random).Bytes()
			i = 0
		}
		o.nextRandom = i + 8
		return int64(binary.LittleEndian.Uint64(o.random[i : i+8]))
	}
	return o.ScDict.GetInt(keyId)
}

func (o *ScUtility) GetString(keyId int32) string {
	switch keyId {
	case wasmhost.KeyBase58:
		return o.base58Encoded
	}
	return o.ScDict.GetString(keyId)
}

func (o *ScUtility) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyBase58:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyHash:
		return wasmhost.OBJTYPE_BYTES
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
		o.hash = hashing.HashData(value).Bytes()
	default:
		o.ScDict.SetBytes(keyId, value)
	}
}

func (o *ScUtility) SetString(keyId int32, value string) {
	switch keyId {
	case wasmhost.KeyBase58:
		o.base58Decoded, _ = base58.Decode(value)
	default:
		o.ScDict.SetString(keyId, value)
	}
}
