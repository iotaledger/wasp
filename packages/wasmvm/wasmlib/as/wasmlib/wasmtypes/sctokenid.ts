// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from '../sandbox';
import {hexDecode, hexEncode, WasmDecoder, WasmEncoder, zeroes} from './codec';
import {Proxy} from './proxy';
import {bytesCompare} from './scbytes';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScTokenIDLength = 38;

export class ScTokenID {
    id: Uint8Array = zeroes(ScTokenIDLength);

    public equals(other: ScTokenID): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
        return tokenIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return tokenIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function tokenIDDecode(dec: WasmDecoder): ScTokenID {
    return tokenIDFromBytesUnchecked(dec.fixedBytes(ScTokenIDLength));
}

export function tokenIDEncode(enc: WasmEncoder, value: ScTokenID): void {
    enc.fixedBytes(value.id, ScTokenIDLength);
}

export function tokenIDFromBytes(buf: Uint8Array | null): ScTokenID {
    if (buf === null || buf.length == 0) {
        return new ScTokenID();
    }
    if (buf.length != ScTokenIDLength) {
        panic('invalid TokenID length');
    }
    return tokenIDFromBytesUnchecked(buf);
}

export function tokenIDToBytes(value: ScTokenID): Uint8Array {
    return value.id;
}

export function tokenIDFromString(value: string): ScTokenID {
    return tokenIDFromBytes(hexDecode(value));
}

export function tokenIDToString(value: ScTokenID): string {
    return hexEncode(value.id);
}

function tokenIDFromBytesUnchecked(buf: Uint8Array): ScTokenID {
    const o = new ScTokenID();
    o.id = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableTokenID {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return tokenIDToString(this.value());
    }

    value(): ScTokenID {
        return tokenIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableTokenID extends ScImmutableTokenID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScTokenID): void {
        this.proxy.set(tokenIDToBytes(value));
    }
}
