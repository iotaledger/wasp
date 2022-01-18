// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

type ScUtility struct {
	ScSandboxObject
	utils iscp.Utils
	wc    *WasmContext
}

func NewScUtility(wc *WasmContext, gasProcessor interface{}) *ScUtility {
	//if gasProcessor == nil {
	//	if wc.ctx != nil {
	//		gasProcessor = wc.ctx.Gas()
	//	} else {
	//		gasProcessor = wc.ctxView.Gas()
	//	}
	//}
	//return &ScUtility{utils: sandbox.NewUtils(gasProcessor), wc: wc}
	return &ScUtility{utils: sandbox_utils.NewUtils(), wc: wc}
}

func (o *ScUtility) CallFunc(keyID int32, bytes []byte) []byte {
	utils := o.utils
	switch keyID {
	case wasmhost.KeyBase58Decode:
		base58Decoded, err := utils.Base58().Decode(string(bytes))
		if err != nil {
			o.Panicf(err.Error())
		}
		return base58Decoded
	case wasmhost.KeyBase58Encode:
		return []byte(utils.Base58().Encode(bytes))
	case wasmhost.KeyBlsAddress:
		address, err := utils.BLS().AddressFromPublicKey(bytes)
		if err != nil {
			o.Panicf(err.Error())
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
			o.Panicf(err.Error())
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
	}
	o.InvalidKey(keyID)
	return nil
}

func (o *ScUtility) Exists(keyID, typeID int32) bool {
	return o.GetTypeID(keyID) > 0
}

func (o *ScUtility) GetTypeID(keyID int32) int32 {
	return wasmhost.OBJTYPE_BYTES
}

func (o *ScUtility) aggregateBLSSignatures(bytes []byte) []byte {
	decode := wasmlib.NewBytesDecoder(bytes)
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
	pubKeyBin, sigBin, err := o.utils.BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	if err != nil {
		o.Panicf(err.Error())
	}
	return wasmlib.NewBytesEncoder().Bytes(pubKeyBin).Bytes(sigBin).Data()
}

func (o *ScUtility) validBLSSignature(bytes []byte) bool {
	decode := wasmlib.NewBytesDecoder(bytes)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	return o.utils.BLS().ValidSignature(data, pubKey, signature)
}

func (o *ScUtility) validED25519Signature(bytes []byte) bool {
	decode := wasmlib.NewBytesDecoder(bytes)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	return o.utils.ED25519().ValidSignature(data, pubKey, signature)
}
