// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScBigInt {
    bytes: u8[] = [];

    constructor() {
    }

    public static fromUint64(value: u64): ScBigInt {
        let bigInt = new ScBigInt();
        bigInt.setUint64(value);
        return bigInt;
    }

    public add(rhs: ScBigInt): ScBigInt {
        let lhs_len = this.bytes.length;
        let rhs_len = rhs.bytes.length;
        if (lhs_len < rhs_len) {
            // always add shorter value to longer value
            return rhs.add(this);
        }

        let res = bigIntFromBytes(this.bytes)
        let carry: u16 = 0;
        for (let i = 0; i < rhs_len; i++) {
            carry += res.bytes[i] as u16 + rhs.bytes[i] as u16;
            res.bytes[i] = carry as u8;
            carry >>= 8;
        }
        if (carry != 0) {
            for (let i = rhs_len; i < lhs_len; i++) {
                carry += res.bytes[i] as u16;
                res.bytes[i] = carry as u8;
                carry >>= 8;
                if (carry == 0) {
                    return res;
                }
            }
            res.bytes.push(1);
        }
        return res;
    }

    public cmp(rhs: ScBigInt): number {
        let lhs_len = this.bytes.length;
        let rhs_len = rhs.bytes.length;
        if (lhs_len != rhs_len) {
            if (lhs_len > rhs_len) {
                return 1;
            }
            return -1;
        }
        for (let i = lhs_len-1; i >= 0; i--) {
            let lhs_byte = this.bytes[i];
            let rhs_byte = rhs.bytes[i];
            if (lhs_byte != rhs_byte) {
                if (lhs_byte > rhs_byte) {
                    return 1;
                }
                return -1;
            }
        }
        return 0;
    }

    public div(rhs: ScBigInt): ScBigInt {
        panic("implement Div");
        return rhs;
    }

    public equals(other: ScBigInt): bool {
        return wasmtypes.bytesCompare(this.bytes, other.bytes) == 0;
    }

    public isUint64(): bool {
        return this.bytes.length <= wasmtypes.ScUint64Length;
    }

    public isZero(): bool {
        return this.bytes.length == 0;
    }

    public modulo(rhs: ScBigInt): ScBigInt {
        panic("implement Modulo");
        return rhs;
    }

    public mul(rhs: ScBigInt): ScBigInt {
        let lhs_len = this.bytes.length;
        let rhs_len = rhs.bytes.length;
        if (lhs_len < rhs_len) {
            // always multiply bigger value by smaller value
            return rhs.mul(this);
        }
        panic("implement Mul");
        return rhs;
    }

    private normalize(): void {
        let buf_len = this.bytes.length;
        while (buf_len > 0 && this.bytes[buf_len - 1] == 0) {
            buf_len--;
        }
        this.bytes = this.bytes.slice(0, buf_len);
    }

    public setUint64(value: u64): void {
        this.bytes = wasmtypes.uint64ToBytes(value);
        this.normalize();
    }

    public sub(rhs: ScBigInt): ScBigInt {
        let cmp = this.cmp(rhs);
        if (cmp <= 0) {
            if (cmp < 0) {
                panic("subtraction underflow");
            }
            return new ScBigInt();
        }
        let lhs_len = this.bytes.length;
        let rhs_len = rhs.bytes.length;

        let res = bigIntFromBytes(this.bytes)
        let borrow: u16 = 0;
        for (let i = 0; i < rhs_len; i++) {
            borrow += res.bytes[i] as u16 - rhs.bytes[i] as u16;
            res.bytes[i] = borrow as u8;
            borrow >>= 8;
        }
        if (borrow != 0) {
            for (let i = rhs_len; i < lhs_len; i++) {
                borrow += res.bytes[i] as u16;
                res.bytes[i] = borrow as u8;
                borrow >>= 8;
                if (borrow == 0) {
                    res.normalize();
                    return res;
                }
            }
        }
        res.normalize();
        return res;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return bigIntToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        // TODO standardize human readable string
        return bigIntToString(this);
    }

    public uint64(): u64 {
        let zeroes = wasmtypes.ScUint64Length-this.bytes.length;
        if (zeroes > wasmtypes.ScUint64Length) {
            panic("value exceeds Uint64");
        }
        let buf = bigIntToBytes(this).concat(wasmtypes.zeroes(zeroes));
        return wasmtypes.uint64FromBytes(buf);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function bigIntDecode(dec: wasmtypes.WasmDecoder): ScBigInt {
    return bigIntFromBytesUnchecked(dec.bytes());
}

export function bigIntEncode(enc: wasmtypes.WasmEncoder, value: ScBigInt): void {
    enc.bytes(value.bytes);
}

export function bigIntFromBytes(buf: u8[]): ScBigInt {
    if (buf.length == 0) {
        return new ScBigInt();
    }
    return bigIntFromBytesUnchecked(buf);
}

export function bigIntToBytes(value: ScBigInt): u8[] {
    return value.bytes;
}

export function bigIntToString(value: ScBigInt): string {
    // TODO standardize human readable string
    return wasmtypes.base58Encode(value.bytes);
}

function bigIntFromBytesUnchecked(buf: u8[]): ScBigInt {
    let o = new ScBigInt();
    o.bytes = buf.slice(0);
    return o;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableBigInt {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return bigIntToString(this.value());
    }

    value(): ScBigInt {
        return bigIntFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableBigInt extends ScImmutableBigInt {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScBigInt): void {
        this.proxy.set(bigIntToBytes(value));
    }
}
