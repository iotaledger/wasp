// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Blake2b, Ed25519} from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import {createHmac} from 'crypto';
import {Codec} from "./codec";

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

    public static subSeed(seed: Uint8Array, index: u32): Uint8Array {
        const zero = new Uint8Array(1);
        const buf = new Uint8Array(4);

        let h = createHmac("sha512", wasmlib.stringToBytes('ed25519 seed'));
        h.update(seed);
        let hash = h.digest();
        let key = hash.subarray(0, 32);
        let chainCode = hash.subarray(32);

        let coinType: u32 = 1;
        switch (Codec.hrpForClient()) {
            case 'iota':
                coinType = 4218;
                break
            case 'smr':
                coinType = 4219;
                break
        }

        const path: u32[] = [44, coinType, index, 0, 0];
        for (let i = 0; i < path.length; i++) {
            const element = path[i] | 0x80000000;
            // big-endian u32
            buf[3] = element as u8;
            buf[2] = (element >> 8) as u8;
            buf[1] = (element >> 16) as u8;
            buf[0] = (element >> 24) as u8;
            h = createHmac("sha512", chainCode);
            h.update(zero);
            h.update(key);
            h.update(buf);
            hash = h.digest();
            key = hash.subarray(0, 32);
            chainCode = hash.subarray(32);
        }
        return key;
    }

    public static fromSubSeed(seed: Uint8Array, index: u32): KeyPair {
        const subSeed = this.subSeed(seed, index);
        return new KeyPair(subSeed);
    }

    public address(): wasmlib.ScAddress {
        const address = Blake2b.sum256(this.publicKey);
        const addr = new Uint8Array(wasmlib.ScLengthEd25519);
        addr[0] = wasmlib.ScAddressEd25519;
        addr.set(address, 1);
        return wasmlib.addressFromBytes(addr);
    }

    public sign(data: Uint8Array): Uint8Array {
        return Ed25519.sign(this.privateKey, data);
    }

    public verify(data: Uint8Array, sig: Uint8Array): bool {
        return Ed25519.verify(this.publicKey, data, sig);
    }
}