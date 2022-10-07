// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Ed25519 } from "@iota/crypto.js"
import {Buffer} from "../../old/wasmclient/buffer";

//TODO
export class KeyPair {
    publicKey: Uint8Array;
    privateKey: Uint8Array;

    public constructor(seed: u8[]) {
        const seedArray = Buffer.from(seed);
        const keyPair = Ed25519.keyPairFromSeed(seedArray);
        this.privateKey = keyPair.privateKey;
        this.publicKey = keyPair.publicKey;
    }

    public sign(data: u8[]): u8[] {
        const message = Buffer.from(data);
        const signed = Ed25519.sign(this.privateKey, message);
        const ret = new Array<u8>(signed.length);
        for (let i = 0; i < signed.length; i++) {
            ret[i] = signed[i];
        }
        return ret;
    }
}