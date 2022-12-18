// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Blake2b, Ed25519} from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import {ScAddress, ScLengthEd25519, uint64ToBytes} from "wasmlib";

export class KeyPair {
    publicKey: Uint8Array;
    privateKey: Uint8Array;

    public constructor(seed: Uint8Array) {
        if (seed.length == 0) {
            this.publicKey = seed;
            this.privateKey = seed;
            return this;
        }
        const seedArray = wasmlib.bytesToUint8Array(seed);
        const keyPair = Ed25519.keyPairFromSeed(seedArray);
        this.privateKey = keyPair.privateKey;
        this.publicKey = keyPair.publicKey;
    }

    public address(): ScAddress {
        const addr = new Uint8Array(wasmlib.ScLengthEd25519);
        addr[0] = wasmlib.ScAddressEd25519
        addr.set(Blake2b.sum256(this.publicKey), 1);
        return wasmlib.addressFromBytes(addr);
    }

    public sign(data: Uint8Array): Uint8Array {
        return Ed25519.sign(this.privateKey, data);
    }

    public static fromSubSeed(seed: Uint8Array, n: u64): KeyPair {
        const indexBytes = uint64ToBytes(n);
        const hashOfIndexBytes = Blake2b.sum256(indexBytes);
        for (let i = 0; i < seed.length; i++) {
            hashOfIndexBytes[i] ^= seed[i];
        }
        return new KeyPair(hashOfIndexBytes);
    }
}