// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmsolo

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

func (s *SoloSandbox) fnUtilsBech32Decode(args []byte) []byte {
	hrp, addr, err := iotago.ParseBech32(string(args))
	s.checkErr(err)
	if hrp != parameters.L1.Protocol.Bech32HRP {
		s.Panicf("Invalid protocol prefix: %s", string(hrp))
	}
	return s.cvt.ScAddress(addr).Bytes()
}

func (s *SoloSandbox) fnUtilsBech32Encode(args []byte) []byte {
	scAddress := wasmtypes.AddressFromBytes(args)
	addr := s.cvt.IscpAddress(&scAddress)
	return []byte(addr.Bech32(parameters.L1.Protocol.Bech32HRP))
}

//func (s *SoloSandbox) fnUtilsBlsAddress(args []byte) []byte {
//	address, err := s.utils.BLS().AddressFromPublicKey(args)
//	s.checkErr(err)
//	return iscp.BytesFromAddress(address)
//}
//
//func (s *SoloSandbox) fnUtilsBlsAggregate(args []byte) []byte {
//	dec := wasmtypes.NewWasmDecoder(args)
//	count := int(wasmtypes.Uint32Decode(dec))
//	pubKeysBin := make([][]byte, count)
//	for i := 0; i < count; i++ {
//		pubKeysBin[i] = dec.Bytes()
//	}
//	count = int(wasmtypes.Uint32Decode(dec))
//	sigsBin := make([][]byte, count)
//	for i := 0; i < count; i++ {
//		sigsBin[i] = dec.Bytes()
//	}
//	pubKeyBin, sigBin, err := s.utils.BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
//	s.checkErr(err)
//	return wasmtypes.NewWasmEncoder().Bytes(pubKeyBin).Bytes(sigBin).Buf()
//}
//
//func (s *SoloSandbox) fnUtilsBlsValid(args []byte) []byte {
//	dec := wasmtypes.NewWasmDecoder(args)
//	data := dec.Bytes()
//	pubKey := dec.Bytes()
//	signature := dec.Bytes()
//	valid := s.utils.BLS().ValidSignature(data, pubKey, signature)
//	return codec.EncodeBool(valid)
//}
//
//func (s *SoloSandbox) fnUtilsEd25519Address(args []byte) []byte {
//	address, err := s.utils.ED25519().AddressFromPublicKey(args)
//	s.checkErr(err)
//	return iscp.BytesFromAddress(address)
//}
//
//func (s *SoloSandbox) fnUtilsEd25519Valid(args []byte) []byte {
//	dec := wasmtypes.NewWasmDecoder(args)
//	data := dec.Bytes()
//	pubKey := dec.Bytes()
//	signature := dec.Bytes()
//	valid := s.utils.ED25519().ValidSignature(data, pubKey, signature)
//	return codec.EncodeBool(valid)
//}
//
//func (s *SoloSandbox) fnUtilsHashBlake2b(args []byte) []byte {
//	return s.utils.Hashing().Blake2b(args).Bytes()
//}

func (s *SoloSandbox) fnUtilsHashName(args []byte) []byte {
	return codec.EncodeHname(s.utils.Hashing().Hname(string(args)))
}

//
//func (s *SoloSandbox) fnUtilsHashSha3(args []byte) []byte {
//	return s.utils.Hashing().Sha3(args).Bytes()
//}
