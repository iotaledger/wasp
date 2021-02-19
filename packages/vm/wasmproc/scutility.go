// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

var TestMode = false

type ScUtility struct {
	ScSandboxObject
	address       address.Address
	aggregatedBls []byte
	base58Decoded []byte
	base58Encoded string
	hash          hashing.HashValue
	hname         coretypes.Hname
	nextRandom    int
	random        []byte
	valid         bool
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
	case wasmhost.KeyAddress:
		return o.address.Bytes()
	case wasmhost.KeyBlsAggregate:
		return o.aggregatedBls
	case wasmhost.KeyBase58Bytes:
		return o.base58Decoded
	case wasmhost.KeyBase58String:
		return []byte(o.base58Encoded)
	case wasmhost.KeyHashBlake2b:
		return o.hash.Bytes()
	case wasmhost.KeyHashSha3:
		return o.hash.Bytes()
	case wasmhost.KeyHname:
		return codec.EncodeHname(o.hname)
	case wasmhost.KeyRandom:
		return o.getRandom8Bytes()
	case wasmhost.KeyValid:
		bytes := make([]byte, 8)
		if o.valid {
			bytes[0] = 1
		}
		return bytes
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
	case wasmhost.KeyAddress:
		return wasmhost.OBJTYPE_ADDRESS
	case wasmhost.KeyBlsAddress:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyBlsAggregate:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyBlsValid:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyBase58Bytes:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyBase58String:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyEd25519Address:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyEd25519Valid:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyHashBlake2b:
		return wasmhost.OBJTYPE_HASH
	case wasmhost.KeyHashSha3:
		return wasmhost.OBJTYPE_HASH
	case wasmhost.KeyHname:
		return wasmhost.OBJTYPE_HNAME
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyRandom:
		return wasmhost.OBJTYPE_INT
	case wasmhost.KeyValid:
		return wasmhost.OBJTYPE_INT
	}
	return 0
}

func (o *ScUtility) SetBytes(keyId int32, typeId int32, bytes []byte) {
	utils := o.vm.utils()
	var err error
	switch keyId {
	case wasmhost.KeyBlsAddress:
		o.address, err = utils.BLS().AddressFromPublicKey(bytes)
	case wasmhost.KeyEd25519Address:
		o.address, err = utils.ED25519().AddressFromPublicKey(bytes)
	case wasmhost.KeyBlsAggregate:
		o.aggregatedBls = o.aggregateBLSSignatures(bytes)
	case wasmhost.KeyBase58Bytes:
		o.base58Encoded = utils.Base58().Encode(bytes)
	case wasmhost.KeyBase58String:
		o.base58Decoded, err = utils.Base58().Decode(string(bytes))
	case wasmhost.KeyHashBlake2b:
		o.hash = utils.Hashing().Blake2b(bytes)
	case wasmhost.KeyHashSha3:
		o.hash = utils.Hashing().Sha3(bytes)
	case wasmhost.KeyName:
		o.hname = utils.Hashing().Hname(string(bytes))
	case wasmhost.KeyBlsValid:
		o.valid = o.validBLSSignature(bytes)
	case wasmhost.KeyEd25519Valid:
		o.valid = o.validED25519Signature(bytes)
	default:
		o.invalidKey(keyId)
	}
	if err != nil {
		o.Panic(err.Error())
	}
}

func (o *ScUtility) aggregateBLSSignatures(bytes []byte) []byte {
	decode := NewBytesDecoder(bytes)
	count := int(decode.Int())
	pubKeysBin := make([][]byte, count)
	for i := 0; i < count; i++ {
		pubKeysBin[i] = decode.Bytes()
	}
	count = int(decode.Int())
	sigsBin := make([][]byte, count)
	for i := 0; i < count; i++ {
		sigsBin[i] = decode.Bytes()
	}
	pubKeyBin, sigBin, err := o.vm.utils().BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	if err != nil {
		o.Panic(err.Error())
	}
	return NewBytesEncoder().Bytes(pubKeyBin).Bytes(sigBin).Data()
}

func (o *ScUtility) validBLSSignature(bytes []byte) bool {
	decode := NewBytesDecoder(bytes)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	ok, err := o.vm.utils().BLS().ValidSignature(data, pubKey, signature)
	if err != nil {
		o.Panic(err.Error())
	}
	return ok
}

func (o *ScUtility) validED25519Signature(bytes []byte) bool {
	decode := NewBytesDecoder(bytes)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	ok, err := o.vm.utils().ED25519().ValidSignature(data, pubKey, signature)
	if err != nil {
		o.Panic(err.Error())
	}
	return ok
}
