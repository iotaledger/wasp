// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func (f WasmToSandbox) fnUtilsBase58Decode(args []byte) []byte {
	bytes, err := f.common.Utils().Base58().Decode(string(args))
	f.checkErr(err)
	return bytes
}

func (f WasmToSandbox) fnUtilsBase58Encode(args []byte) []byte {
	return []byte(f.common.Utils().Base58().Encode(args))
}

func (f WasmToSandbox) fnUtilsBlsAddress(args []byte) []byte {
	address, err := f.common.Utils().BLS().AddressFromPublicKey(args)
	f.checkErr(err)
	return address.Bytes()
}

func (f WasmToSandbox) fnUtilsBlsAggregate(args []byte) []byte {
	decode := NewBytesDecoder(args)
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
	pubKeyBin, sigBin, err := f.common.Utils().BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	f.checkErr(err)
	return NewBytesEncoder().Bytes(pubKeyBin).Bytes(sigBin).Data()
}

func (f WasmToSandbox) fnUtilsBlsValid(args []byte) []byte {
	decode := NewBytesDecoder(args)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	valid := f.common.Utils().BLS().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (f WasmToSandbox) fnUtilsEd25519Address(args []byte) []byte {
	address, err := f.common.Utils().ED25519().AddressFromPublicKey(args)
	f.checkErr(err)
	return address.Bytes()
}

func (f WasmToSandbox) fnUtilsEd25519Valid(args []byte) []byte {
	decode := NewBytesDecoder(args)
	data := decode.Bytes()
	pubKey := decode.Bytes()
	signature := decode.Bytes()
	valid := f.common.Utils().ED25519().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (f WasmToSandbox) fnUtilsHashBlake2b(args []byte) []byte {
	return f.common.Utils().Hashing().Blake2b(args).Bytes()
}

func (f WasmToSandbox) fnUtilsHashName(args []byte) []byte {
	return codec.EncodeHname(f.common.Utils().Hashing().Hname(string(args)))
}

func (f WasmToSandbox) fnUtilsHashSha3(args []byte) []byte {
	return f.common.Utils().Hashing().Sha3(args).Bytes()
}
