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
        const bigInt = new ScBigInt();
        bigInt.setUint64(value);
        return bigInt;
    }

    public add(rhs: ScBigInt): ScBigInt {
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen < rhsLen) {
            // always add shorter value to longer value
            return rhs.add(this);
        }

        const res = bigIntFromBytes(this.bytes)
        let carry: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            carry += res.bytes[i] as u16 + rhs.bytes[i] as u16;
            res.bytes[i] = carry as u8;
            carry >>= 8;
        }
        if (carry != 0) {
            for (let i = rhsLen; i < lhsLen; i++) {
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
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen != rhsLen) {
            if (lhsLen > rhsLen) {
                return 1;
            }
            return -1;
        }
        for (let i = lhsLen-1; i >= 0; i--) {
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
        if (rhs.bytes.length == 1 && rhs.bytes[0] == 1) {
            // divide by 1, quo = lhs, rem = 0
            return [this, new ScBigInt()];
        }
        panic("implement Div");
        return [this, rhs];
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
        return this.divMod(rhs)[1];
    }

    public mul(rhs: ScBigInt): ScBigInt {
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen < rhsLen) {
            // always multiply bigger value by smaller value
            return rhs.mul(this);
        }
        if (lhsLen+rhsLen <= wasmtypes.ScUint64Length) {
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
        panic("implement Mul");
        return rhs;
    }

    private normalize(): void {
        let bufLen = this.bytes.length;
        while (bufLen > 0 && this.bytes[bufLen - 1] == 0) {
            bufLen--;
        }
        this.bytes = this.bytes.slice(0, bufLen);
    }

    public setUint64(value: u64): void {
        this.bytes = wasmtypes.uint64ToBytes(value);
        this.normalize();
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

        const res = bigIntFromBytes(this.bytes)
        let borrow: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            borrow += res.bytes[i] as u16 - rhs.bytes[i] as u16;
            res.bytes[i] = borrow as u8;
            borrow >>= 8;
        }
        if (borrow != 0) {
            for (let i = rhsLen; i < lhsLen; i++) {
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
        const zeroes = wasmtypes.ScUint64Length-this.bytes.length;
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
    const zeroes =  wasmtypes.zeroes(18-digits.length);
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
