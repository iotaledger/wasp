// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Bytes, Int32, Item, Items, panic} from "."
import {Encoder} from "./encoder"
import {Buffer} from "./buffer";

// The Arguments struct is used to gather all arguments for this smart
// contract function call and encode it into this deterministic byte array
export class Arguments extends Encoder {
    private args = new Map<string, Bytes>();
    // get(key: string): wasmclient.Bytes {
    //     const bytes = this.args.get(key);
    //     return bytes ?? Buffer.alloc(0);
    // }

    public set(key: string, val: Bytes): void {
        this.args.set(key, val);
    }

    public indexedKey(key: string, index: Int32): string {
        return key + "." + index.toString();
    }

    public mandatory(key: string): void {
        if (!this.args.has(key)) {
            panic("missing mandatory " + key)
        }
    }

    // Encode returns a byte array that encodes the Arguments as follows:
    // Sort all keys in ascending order (very important, because this data
    // will be part of the data that will be signed, so the order needs to
    // be 100% deterministic). Then emit the 4-byte argument count.
    // Next for each argument emit the 2-byte key length, the key prepended
    // with the minus sign, the 4-byte value length, and then the value bytes.
    public encode(): Bytes {
        const keys = new Array<string>();
        for (const key of this.args.keys()) {
            keys.push(key);
        }
        keys.sort((lhs, rhs) => lhs.localeCompare(rhs));

        let buf = Buffer.alloc(4);
        buf.writeUInt32LE(keys.length, 0);
        for (const key of keys) {
            const keyBuf = Buffer.from("-" + key);
            const keyLen = Buffer.alloc(2);
            keyLen.writeUInt16LE(keyBuf.length, 0);
            const valBuf = this.args.get(key);
            if (!valBuf) {
                throw new Error("Arguments.encode: missing value");
            }
            const valLen = Buffer.alloc(4);
            valLen.writeUInt32LE(valBuf.length, 0);
            buf = Buffer.concat([buf, keyLen, keyBuf, valLen, valBuf]);
        }
        return buf;
    }

    public encodeCall(): Items {
        const items = new Items()
        for (const [key, val] of this.args) {
            const k = Buffer.from(key).toString("base64");
            const v = val.toString("base64");
            items.Items.push(new Item(k, v))
        }
        return items;
    }
}
