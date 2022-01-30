// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ScSandboxUtils struct{}

// decodes the specified base58-encoded string value to its original bytes
func (u ScSandboxUtils) Base58Decode(value string) []byte {
	return Sandbox(FnUtilsBase58Decode, []byte(value))
}

// encodes the specified bytes to a base-58-encoded string
func (u ScSandboxUtils) Base58Encode(bytes []byte) string {
	return string(Sandbox(FnUtilsBase58Encode, bytes))
}

func (u ScSandboxUtils) BlsAddressFromPubKey(pubKey []byte) wasmtypes.ScAddress {
	return wasmtypes.AddressFromBytes(Sandbox(FnUtilsBlsAddress, pubKey))
}

func (u ScSandboxUtils) BlsAggregateSignatures(pubKeys, sigs [][]byte) ([]byte, []byte) {
	enc := wasmtypes.NewWasmEncoder()
	wasmtypes.Uint32Encode(enc, uint32(len(pubKeys)))
	for _, pubKey := range pubKeys {
		enc.Bytes(pubKey)
	}
	wasmtypes.Uint32Encode(enc, uint32(len(sigs)))
	for _, sig := range sigs {
		enc.Bytes(sig)
	}
	result := Sandbox(FnUtilsBlsAggregate, enc.Buf())
	decode := wasmtypes.NewWasmDecoder(result)
	return decode.Bytes(), decode.Bytes()
}

func (u ScSandboxUtils) BlsValidSignature(data, pubKey, signature []byte) bool {
	enc := wasmtypes.NewWasmEncoder().Bytes(data).Bytes(pubKey).Bytes(signature)
	return wasmtypes.BoolFromBytes(Sandbox(FnUtilsBlsValid, enc.Buf()))
}

func (u ScSandboxUtils) Ed25519AddressFromPubKey(pubKey []byte) wasmtypes.ScAddress {
	return wasmtypes.AddressFromBytes(Sandbox(FnUtilsEd25519Address, pubKey))
}

func (u ScSandboxUtils) Ed25519ValidSignature(data, pubKey, signature []byte) bool {
	enc := wasmtypes.NewWasmEncoder().Bytes(data).Bytes(pubKey).Bytes(signature)
	return wasmtypes.BoolFromBytes(Sandbox(FnUtilsEd25519Valid, enc.Buf()))
}

// hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
func (u ScSandboxUtils) HashBlake2b(value []byte) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(Sandbox(FnUtilsHashBlake2b, value))
}

// hashes the specified value bytes using sha3 hashing and returns the resulting 32-byte hash
func (u ScSandboxUtils) HashSha3(value []byte) wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(Sandbox(FnUtilsHashSha3, value))
}

// hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
func (u ScSandboxUtils) Hname(value string) wasmtypes.ScHname {
	return wasmtypes.HnameFromBytes(Sandbox(FnUtilsHashName, []byte(value)))
}

// converts an integer to its string representation
func (u ScSandboxUtils) String(value int64) string {
	return wasmtypes.Int64ToString(value)
}
