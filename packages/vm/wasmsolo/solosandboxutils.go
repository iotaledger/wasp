// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmcodec"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

func (s *SoloSandbox) fnUtilsBase58Decode(args []byte) []byte {
	bytes, err := s.utils.Base58().Decode(string(args))
	s.checkErr(err)
	return bytes
}

func (s *SoloSandbox) fnUtilsBase58Encode(args []byte) []byte {
	return []byte(s.utils.Base58().Encode(args))
}

func (s *SoloSandbox) fnUtilsBlsAddress(args []byte) []byte {
	address, err := s.utils.BLS().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s *SoloSandbox) fnUtilsBlsAggregate(args []byte) []byte {
	dec := wasmcodec.NewWasmDecoder(args)
	count := int(wasmtypes.DecodeUint32(dec))
	pubKeysBin := make([][]byte, count)
	for i := 0; i < count; i++ {
		pubKeysBin[i] = dec.Bytes()
	}
	count = int(wasmtypes.DecodeUint32(dec))
	sigsBin := make([][]byte, count)
	for i := 0; i < count; i++ {
		sigsBin[i] = dec.Bytes()
	}
	pubKeyBin, sigBin, err := s.utils.BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	s.checkErr(err)
	return wasmcodec.NewWasmEncoder().Bytes(pubKeyBin).Bytes(sigBin).Buf()
}

func (s *SoloSandbox) fnUtilsBlsValid(args []byte) []byte {
	dec := wasmcodec.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.utils.BLS().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s *SoloSandbox) fnUtilsEd25519Address(args []byte) []byte {
	address, err := s.utils.ED25519().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s *SoloSandbox) fnUtilsEd25519Valid(args []byte) []byte {
	dec := wasmcodec.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.utils.ED25519().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s *SoloSandbox) fnUtilsHashBlake2b(args []byte) []byte {
	return s.utils.Hashing().Blake2b(args).Bytes()
}

func (s *SoloSandbox) fnUtilsHashName(args []byte) []byte {
	return codec.EncodeHname(s.utils.Hashing().Hname(string(args)))
}

func (s *SoloSandbox) fnUtilsHashSha3(args []byte) []byte {
	return s.utils.Hashing().Sha3(args).Bytes()
}
