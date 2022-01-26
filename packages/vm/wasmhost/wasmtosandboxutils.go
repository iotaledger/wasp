// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

func (s WasmToSandbox) fnUtilsBase58Decode(args []byte) []byte {
	bytes, err := s.common.Utils().Base58().Decode(string(args))
	s.checkErr(err)
	return bytes
}

func (s WasmToSandbox) fnUtilsBase58Encode(args []byte) []byte {
	return []byte(s.common.Utils().Base58().Encode(args))
}

func (s WasmToSandbox) fnUtilsBlsAddress(args []byte) []byte {
	address, err := s.common.Utils().BLS().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s WasmToSandbox) fnUtilsBlsAggregate(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	count := wasmtypes.DecodeUint32(dec)
	pubKeysBin := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		pubKeysBin[i] = dec.Bytes()
	}
	count = wasmtypes.DecodeUint32(dec)
	sigsBin := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		sigsBin[i] = dec.Bytes()
	}
	pubKeyBin, sigBin, err := s.common.Utils().BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	s.checkErr(err)
	return wasmtypes.NewWasmEncoder().Bytes(pubKeyBin).Bytes(sigBin).Buf()
}

func (s WasmToSandbox) fnUtilsBlsValid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().BLS().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmToSandbox) fnUtilsEd25519Address(args []byte) []byte {
	address, err := s.common.Utils().ED25519().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s WasmToSandbox) fnUtilsEd25519Valid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().ED25519().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmToSandbox) fnUtilsHashBlake2b(args []byte) []byte {
	return s.common.Utils().Hashing().Blake2b(args).Bytes()
}

func (s WasmToSandbox) fnUtilsHashName(args []byte) []byte {
	return codec.EncodeHname(s.common.Utils().Hashing().Hname(string(args)))
}

func (s WasmToSandbox) fnUtilsHashSha3(args []byte) []byte {
	return s.common.Utils().Hashing().Sha3(args).Bytes()
}
