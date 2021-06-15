// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

var TestMode = false

type ScUtility struct {
	ScSandboxObject
	nextRandom int
	random     []byte
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

func (o *ScUtility) CallFunc(keyId int32, bytes []byte) []byte {
	utils := o.vm.utils()
	switch keyId {
	case wasmhost.KeyBase58Decode:
		base58Decoded, err := utils.Base58().Decode(string(bytes))
		if err != nil {
			o.Panic(err.Error())
		}
		return base58Decoded
	case wasmhost.KeyBase58Encode:
		return []byte(utils.Base58().Encode(bytes))
	case wasmhost.KeyBlsAddress:
		address, err := utils.BLS().AddressFromPublicKey(bytes)
		if err != nil {
			o.Panic(err.Error())
		}
		return address.Bytes()
	case wasmhost.KeyBlsAggregate:
		return o.aggregateBLSSignatures(bytes)
	case wasmhost.KeyBlsValid:
		if o.validBLSSignature(bytes) {
			var flag [1]byte
			return flag[:]
		}
		return nil
	case wasmhost.KeyEd25519Address:
		address, err := utils.ED25519().AddressFromPublicKey(bytes)
		if err != nil {
			o.Panic(err.Error())
		}
		return address.Bytes()
	case wasmhost.KeyEd25519Valid:
		if o.validED25519Signature(bytes) {
			var flag [1]byte
			return flag[:]
		}
		return nil
	case wasmhost.KeyHashBlake2b:
		return utils.Hashing().Blake2b(bytes).Bytes()
	case wasmhost.KeyHashSha3:
		return utils.Hashing().Sha3(bytes).Bytes()
	case wasmhost.KeyHname:
		return codec.EncodeHname(utils.Hashing().Hname(string(bytes)))
	case wasmhost.KeyRandom:
		return o.getRandom8Bytes()
	}
	o.invalidKey(keyId)
	return nil
}

func (o *ScUtility) Exists(keyId int32, typeId int32) bool {
	return o.GetTypeId(keyId) > 0
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
	return wasmhost.OBJTYPE_BYTES
}

func (o *ScUtility) aggregateBLSSignatures(bytes []byte) []byte {
	decode := NewBytesDecoder(bytes)
	count := int(decode.Int32())
	pubKeysBin := make([][]byte, count)
	for i := 0; i < count; i++ {
		pubKeysBin[i] = decode.Bytes()
	}
	count = int(decode.Int32())
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
	return o.vm.utils().BLS().ValidSignature(data, pubKey, signature)
}

func (o *ScUtility) validED25519Signature(bytes []byte) bool {
	decode := NewBytesDecoder(bytes)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	return o.vm.utils().ED25519().ValidSignature(data, pubKey, signature)
}
