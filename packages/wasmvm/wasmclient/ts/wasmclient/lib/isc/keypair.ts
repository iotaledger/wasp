// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Blake2b, Ed25519} from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import {ScAddress, uint64ToBytes} from 'wasmlib';

export class KeyPair {
    publicKey: Uint8Array;
    privateKey: Uint8Array;

    public constructor(seed: Uint8Array) {
        if (seed.length == 0) {
            this.publicKey = seed;
            this.privateKey = seed;
            return this;
        }
        const keyPair = Ed25519.keyPairFromSeed(seed);
        this.privateKey = keyPair.privateKey;
        this.publicKey = keyPair.publicKey;
    }

    public static subSeed(seed: Uint8Array, n: u64): Uint8Array {
        const indexBytes = uint64ToBytes(n);
        const hashOfIndexBytes = Blake2b.sum256(indexBytes);
        for (let i = 0; i < seed.length; i++) {
            hashOfIndexBytes[i] ^= seed[i];
        }
        return hashOfIndexBytes;
    }

    public static fromSubSeed(seed: Uint8Array, n: u64): KeyPair {
        const subSeed = this.subSeed(seed, n);
        return new KeyPair(subSeed);
    }

    public address(): ScAddress {
        const addr = new Uint8Array(wasmlib.ScLengthEd25519);
        addr[0] = wasmlib.ScAddressEd25519;
        addr.set(Blake2b.sum256(this.publicKey), 1);
        return wasmlib.addressFromBytes(addr);
    }

    public sign(data: Uint8Array): Uint8Array {
        return Ed25519.sign(this.privateKey, data);
    }

    public verify(data: Uint8Array, sig: Uint8Array): bool {
        return Ed25519.verify(this.publicKey, data, sig);
    }
}