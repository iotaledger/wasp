// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as util from "./sandbox";
import * as wasmtypes from "./wasmtypes"
import {sandbox} from "./host";

export class ScSandboxUtils {
    // decodes the specified base58-encoded string value to its original bytes
    public base58Decode(value: string): u8[] {
        return sandbox(util.FnUtilsBase58Decode, wasmtypes.stringToBytes(value));
    }

    // encodes the specified bytes to a base-58-encoded string
    public base58Encode(bytes: u8[]): string {
        return wasmtypes.bytesToString(sandbox(util.FnUtilsBase58Encode, bytes));
    }

    public blsAddressFromPubKey(pubKey: u8[]): wasmtypes.ScAddress {
        return wasmtypes.addressFromBytes(sandbox(util.FnUtilsBlsAddress, pubKey));
    }

    public blsAggregateSignatures(pubKeys: u8[][], sigs: u8[][]): u8[][] {
        const enc = new wasmtypes.WasmEncoder();
        wasmtypes.uint32Encode(enc, pubKeys.length as u32);
        for (let i = 0; i < pubKeys.length; i++) {
            enc.bytes(pubKeys[i]);
        }
        wasmtypes.uint32Encode(enc, sigs.length as u32);
        for (let i = 0; i < sigs.length; i++) {
            enc.bytes(sigs[i]);
        }
        const result = sandbox(util.FnUtilsBlsAggregate, enc.buf());
        const decode = new wasmtypes.WasmDecoder(result);
        return [decode.bytes(), decode.bytes()];
    }

    public blsValidSignature(data: u8[], pubKey: u8[], signature: u8[]): bool {
        const enc = new wasmtypes.WasmEncoder().bytes(data).bytes(pubKey).bytes(signature);
        return wasmtypes.boolFromBytes(sandbox(util.FnUtilsBlsValid, enc.buf()));
    }

    public ed25519AddressFromPubKey(pubKey: u8[]): wasmtypes.ScAddress {
        return wasmtypes.addressFromBytes(sandbox(util.FnUtilsEd25519Address, pubKey));
    }

    public ed25519ValidSignature(data: u8[], pubKey: u8[], signature: u8[]): bool {
        const enc = new wasmtypes.WasmEncoder().bytes(data).bytes(pubKey).bytes(signature);
        return wasmtypes.boolFromBytes(sandbox(util.FnUtilsEd25519Valid, enc.buf()));
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    public hashBlake2b(value: u8[]): wasmtypes.ScHash {
        return wasmtypes.hashFromBytes(sandbox(util.FnUtilsHashBlake2b, value));
    }

    // hashes the specified value bytes using sha3 hashing and returns the resulting 32-byte hash
    public hashSha3(value: u8[]): wasmtypes.ScHash {
        return wasmtypes.hashFromBytes(sandbox(util.FnUtilsHashSha3, value));
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    public hname(value: string): wasmtypes.ScHname {
        return wasmtypes.hnameFromBytes(sandbox(util.FnUtilsHashName, wasmtypes.stringToBytes(value)));
    }
}
