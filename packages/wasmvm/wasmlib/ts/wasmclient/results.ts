// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index";
import {Buffer} from "./buffer";

export class Results extends wasmclient.Decoder {
    protected keys = new Map<string, Buffer>();
    protected res = new Map<string, Buffer>();

    protected exists(key: string): boolean {
        return this.res.has(key);
    }

    protected forEach(keyValue: (key: Buffer, val: Buffer) => void): void {
        this.res.forEach((val, key) => {
            const keyBuf = this.keys.get(key);
            if (keyBuf === undefined) {
                wasmclient.panic("missing key");
                return;
            }
            keyValue(keyBuf, val);
        });
    }

    protected get(key: string): Buffer | undefined {
        return this.res.get(key);
    }

    public set(key: Buffer, val: Buffer) {
        const stringKey = key.toString();
        this.keys.set(stringKey, key);
        this.res.set(stringKey, val);
    }
}
