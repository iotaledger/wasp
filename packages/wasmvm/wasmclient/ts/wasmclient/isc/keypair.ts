// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Ed25519} from "@iota/crypto.js"
import * as wasmlib from "wasmlib"

export class KeyPair {
    publicKey: Uint8Array;
    privateKey: Uint8Array;

    public constructor(seed: u8[]) {
        const seedArray = wasmlib.bytesToUint8Array(seed);
        const keyPair = Ed25519.keyPairFromSeed(seedArray);
        this.privateKey = keyPair.privateKey;
        this.publicKey = keyPair.publicKey;
    }

    public sign(data: u8[]): u8[] {
        const message = wasmlib.bytesToUint8Array(data);
        const signed = Ed25519.sign(this.privateKey, message);
        return wasmlib.bytesFromUint8Array(signed);
    }
}