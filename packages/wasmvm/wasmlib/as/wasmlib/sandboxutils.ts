// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {sandbox} from "./host";
import {
    FnUtilsBech32Decode,
    FnUtilsBech32Encode,
    FnUtilsBlsAddress,
    FnUtilsBlsAggregate,
    FnUtilsBlsValid,
    FnUtilsEd25519Address,
    FnUtilsEd25519Valid,
    FnUtilsHashBlake2b,
    FnUtilsHashName,
    FnUtilsHashSha3
} from "./sandbox";
import {boolFromBytes} from "./wasmtypes/scbool";
import {hashFromBytes, ScHash} from "./wasmtypes/schash";
import {WasmDecoder, WasmEncoder} from "./wasmtypes/codec";
import {addressFromBytes, addressToBytes, ScAddress} from "./wasmtypes/scaddress";
import {stringFromBytes, stringToBytes} from "./wasmtypes/scstring";
import {uint32Encode} from "./wasmtypes/scuint32";
import {hnameFromBytes, ScHname} from "./wasmtypes/schname";

export class ScSandboxUtils {
    // decodes the specified bech32-encoded string value to its original bytes
    public bech32Decode(value: string): ScAddress {
        return addressFromBytes(sandbox(FnUtilsBech32Decode, stringToBytes(value)));
    }

    // encodes the specified bytes to a bech32-encoded string
    public bech32Encode(addr: ScAddress): string {
        return stringFromBytes(sandbox(FnUtilsBech32Encode, addressToBytes(addr)));
    }

    public blsAddressFromPubKey(pubKey: Uint8Array): ScAddress {
        return addressFromBytes(sandbox(FnUtilsBlsAddress, pubKey));
    }

    public blsAggregateSignatures(pubKeys: Uint8Array[], sigs: Uint8Array[]): Uint8Array[] {
        const enc = new WasmEncoder();
        uint32Encode(enc, pubKeys.length as u32);
        for (let i = 0; i < pubKeys.length; i++) {
            enc.bytes(pubKeys[i]);
        }
        uint32Encode(enc, sigs.length as u32);
        for (let i = 0; i < sigs.length; i++) {
            enc.bytes(sigs[i]);
        }
        const result = sandbox(FnUtilsBlsAggregate, enc.buf());
        const decode = new WasmDecoder(result);
        return [decode.bytes(), decode.bytes()];
    }

    public blsValidSignature(data: Uint8Array, pubKey: Uint8Array, signature: Uint8Array): bool {
        const enc = new WasmEncoder().bytes(data).bytes(pubKey).bytes(signature);
        return boolFromBytes(sandbox(FnUtilsBlsValid, enc.buf()));
    }

    public ed25519AddressFromPubKey(pubKey: Uint8Array): ScAddress {
        return addressFromBytes(sandbox(FnUtilsEd25519Address, pubKey));
    }

    public ed25519ValidSignature(data: Uint8Array, pubKey: Uint8Array, signature: Uint8Array): bool {
        const enc = new WasmEncoder().bytes(data).bytes(pubKey).bytes(signature);
        return boolFromBytes(sandbox(FnUtilsEd25519Valid, enc.buf()));
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    public hashBlake2b(value: Uint8Array): ScHash {
        return hashFromBytes(sandbox(FnUtilsHashBlake2b, value));
    }

    // hashes the specified value bytes using sha3 hashing and returns the resulting 32-byte hash
    public hashSha3(value: Uint8Array): ScHash {
        return hashFromBytes(sandbox(FnUtilsHashSha3, value));
    }

    // hashes the specified value bytes using blake2b hashing and returns the resulting 32-byte hash
    public hname(value: string): ScHname {
        return hnameFromBytes(sandbox(FnUtilsHashName, stringToBytes(value)));
    }
}
