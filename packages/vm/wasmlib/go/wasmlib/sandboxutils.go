// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

type ScSandboxUtils struct{}

func (u ScSandboxUtils) Base58Decode(value string) []byte {
	return Sandbox(FnUtilsBase58Decode, []byte(value))
}

func (u ScSandboxUtils) Base58Encode(bytes []byte) string {
	return string(Sandbox(FnUtilsBase58Encode, bytes))
}

func (u ScSandboxUtils) BlsAddress(pubKey []byte) ScAddress {
	return NewScAddressFromBytes(Sandbox(FnUtilsBlsAddress, pubKey))
}

func (u ScSandboxUtils) BlsAggregate(pubKeys, sigs [][]byte) ([]byte, []byte) {
	encode := NewBytesEncoder()
	encode.Int32(int32(len(pubKeys)))
	for _, pubKey := range pubKeys {
		encode.Bytes(pubKey)
	}
	encode.Int32(int32(len(sigs)))
	for _, sig := range sigs {
		encode.Bytes(sig)
	}
	result := Sandbox(FnUtilsBlsAggregate, encode.Data())
	decode := NewBytesDecoder(result)
	return decode.Bytes(), decode.Bytes()
}

func (u ScSandboxUtils) BlsValid(data, pubKey, signature []byte) bool {
	encode := NewBytesEncoder().Bytes(data).Bytes(pubKey).Bytes(signature)
	valid, _ := ExtractBool(Sandbox(FnUtilsBlsValid, encode.Data()))
	return valid
}

func (u ScSandboxUtils) Ed25519Address(pubKey []byte) ScAddress {
	return NewScAddressFromBytes(Sandbox(FnUtilsEd25519Address, pubKey))
}

func (u ScSandboxUtils) Ed25519Valid(data, pubKey, signature []byte) bool {
	encode := NewBytesEncoder().Bytes(data).Bytes(pubKey).Bytes(signature)
	valid, _ := ExtractBool(Sandbox(FnUtilsEd25519Valid, encode.Data()))
	return valid
}

func (u ScSandboxUtils) HashBlake2b(value []byte) ScHash {
	return NewScHashFromBytes(Sandbox(FnUtilsHashBlake2b, value))
}

func (u ScSandboxUtils) HashName(value string) ScHname {
	return NewScHnameFromBytes(Sandbox(FnUtilsHashName, []byte(value)))
}

func (u ScSandboxUtils) HashSha3(value []byte) ScHash {
	return NewScHashFromBytes(Sandbox(FnUtilsHashSha3, value))
}
