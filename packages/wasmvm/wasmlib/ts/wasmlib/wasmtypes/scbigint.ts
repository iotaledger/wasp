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
        return ScBigInt.normalize(wasmtypes.uint64ToBytes(value));
    }

    private static normalize(buf: u8[]): ScBigInt {
        let bufLen = buf.length;
        while (bufLen > 0 && buf[bufLen - 1] == 0) {
            bufLen--;
        }
        const res = new ScBigInt();
        res.bytes = buf.slice(0, bufLen);
        return res;
    }

    public add(rhs: ScBigInt): ScBigInt {
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen < rhsLen) {
            // always add shorter value to longer value
            return rhs.add(this);
        }

        const buf: u8[] = new Array(lhsLen);
        let carry: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            carry += (this.bytes[i] as u16) + (rhs.bytes[i] as u16);
            buf[i] = carry as u8;
            carry >>= 8;
        }
        for (let i = rhsLen; carry != 0 && i < lhsLen; i++) {
            carry += this.bytes[i] as u16;
            buf[i] = carry as u8;
            carry >>= 8;
        }
        if (carry != 0) {
            buf.push(1);
        }
        return ScBigInt.normalize(buf);
    }

    public cmp(rhs: ScBigInt): number {
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen != rhsLen) {
            if (lhsLen > rhsLen) {
                return 1;
            }
            return -1;
        }
        for (let i = lhsLen - 1; i >= 0; i--) {
            const lhsByte = this.bytes[i];
            const rhsByte = rhs.bytes[i];
            if (lhsByte != rhsByte) {
                if (lhsByte > rhsByte) {
                    return 1;
                }
                return -1;
            }
        }
        return 0;
    }

    public div(rhs: ScBigInt): ScBigInt {
        return this.divMod(rhs)[0];
    }

    public divMod(rhs: ScBigInt): ScBigInt[] {
        if (rhs.isZero()) {
            panic("divide by zero");
        }
        const cmp = this.cmp(rhs);
        if (cmp <= 0) {
            if (cmp < 0) {
                // divide by larger value, quo = 0, rem = lhs
                return [new ScBigInt(), this];
            }
            // divide equal values, quo = 1, rem = 0
            return [ScBigInt.fromUint64(1), new ScBigInt()];
        }
        if (this.isUint64()) {
            // let standard uint64 type do the heavy lifting
            const lhs64 = this.uint64();
            const rhs64 = rhs.uint64();
            const div = ScBigInt.fromUint64(lhs64 / rhs64);
            return [div, ScBigInt.fromUint64(lhs64 % rhs64)];
        }
        if (rhs.bytes.length == 1) {
            if (rhs.bytes[0] == 1) {
                // divide by 1, quo = lhs, rem = 0
                return [this, new ScBigInt()];
            }
            return this.divModSimple(rhs.bytes[0]);
        }
        //TODO
        panic("implement rest of DivMod");
        return [this, rhs];
    }

    private divModSimple(value: u8): ScBigInt[] {
        const lhsLen = this.bytes.length;
        const buf: u8[] = new Array(lhsLen);
        let remain: u16 = 0;
        const rhs = value as u16;
        for (let i = lhsLen - 1; i >= 0; i--) {
            remain = (remain << 8) + (this.bytes[i] as u16);
            buf[i] = (remain / rhs) as u8;
            remain %= rhs;
        }
        return [ScBigInt.normalize(buf), ScBigInt.normalize([remain as u8])];
    }

    public equals(rhs: ScBigInt): bool {
        return this.cmp(rhs) == 0;
    }

    public isUint64(): bool {
        return this.bytes.length <= wasmtypes.ScUint64Length;
    }

    public isZero(): bool {
        return this.bytes.length == 0;
    }

    public modulo(rhs: ScBigInt): ScBigInt {
        return this.divMod(rhs)[1];
    }

    public mul(rhs: ScBigInt): ScBigInt {
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen < rhsLen) {
            // always multiply bigger value by smaller value
            return rhs.mul(this);
        }
        if (lhsLen + rhsLen <= wasmtypes.ScUint64Length) {
            return ScBigInt.fromUint64(this.uint64() * rhs.uint64());
        }
        if (rhsLen == 0) {
            // multiply by zero, result zero
            return new ScBigInt();
        }
        if (rhsLen == 1 && rhs.bytes[0] == 1) {
            // multiply by one, result lhs
            return this;
        }

        //TODO optimize by using u32 words instead of u8 words
        const buf = wasmtypes.zeroes(lhsLen + rhsLen);
        for (let r = 0; r < rhsLen; r++) {
            let carry: u16 = 0;
            for (let l = 0; l < lhsLen; l++) {
                carry += (buf[l + r] as u16) + (this.bytes[l] as u16) * (rhs.bytes[r] as u16);
                buf[l + r] = carry as u8;
                carry >>= 8;
            }
            buf[r + lhsLen] = carry as u8;
        }
        return ScBigInt.normalize(buf);
    }

    public sub(rhs: ScBigInt): ScBigInt {
        const cmp = this.cmp(rhs);
        if (cmp <= 0) {
            if (cmp < 0) {
                panic("subtraction underflow");
            }
            return new ScBigInt();
        }
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;

        const buf: u8[] = new Array(lhsLen);
        let borrow: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            borrow += (this.bytes[i] as u16) - (rhs.bytes[i] as u16);
            buf[i] = borrow as u8;
            borrow >>= 8;
        }
        for (let i = rhsLen; i < lhsLen; i++) {
            borrow += this.bytes[i] as u16;
            buf[i] = borrow as u8;
            borrow >>= 8;
        }
        return ScBigInt.normalize(buf);
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
        const zeroes = wasmtypes.ScUint64Length - this.bytes.length;
        if (zeroes > wasmtypes.ScUint64Length) {
            panic("value exceeds Uint64");
        }
        const buf = bigIntToBytes(this).concat(wasmtypes.zeroes(zeroes));
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
    if (value.isUint64()) {
        return wasmtypes.uint64ToString(value.uint64());
    }
    const divMod = value.divMod(ScBigInt.fromUint64(1_000_000_000_000_000_000));
    const digits = wasmtypes.uint64ToString(divMod[1].uint64());
    const zeroes = wasmtypes.zeroes(18 - digits.length);
    return bigIntToString(divMod[0]) + zeroes + digits;
}

function bigIntFromBytesUnchecked(buf: u8[]): ScBigInt {
    const o = new ScBigInt();
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
