// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScBigInt {
    bytes: u8[] = [];

    private static zero: ScBigInt = new ScBigInt();
    private static one: ScBigInt = ScBigInt.fromUint64(1);

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

        const buf = new Array<u8>(lhsLen);
        let carry: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            carry += (this.bytes[i] as u16) + (rhs.bytes[i] as u16);
            buf[i] = carry as u8;
            carry >>= 8;
        }
        for (let i = rhsLen; i < lhsLen; i++) {
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
                return [ScBigInt.zero, ScBigInt.normalize(this.bytes)];
            }
            // divide equal values, quo = 1, rem = 0
            return [ScBigInt.one, ScBigInt.zero];
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
                return [ScBigInt.normalize(this.bytes), ScBigInt.zero];
            }
            return this.divModSingleByte(rhs.bytes[0]);
        }
        return this.divModMultiByte(rhs);
    }

    // divModEstimate uses long division byte estimation and corrects on the remainder
    // by recursively estimating the next byte and discounting the result in the quotient
    public divModEstimate(rhs: ScBigInt): ScBigInt[] {
        // first filter out the simplest cases, when lengths are equal
        // that either results in a quotient of one or zero
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;
        if (lhsLen <= rhsLen) {
            if (this.cmp(rhs) < 0) {
                return [ScBigInt.zero, this];
            }
            return [ScBigInt.one, this.sub(rhs)];
        }

        // Now for the big estimation trick. Since we normalized the rhs to have the msb high bit
        // set, dividing two lhs MSBs by the rhs MSB will in over 80% of the cases be the correct
        // value for the entire lhs / rhs, almost 20% will be 1 too high, and about 0.5% will be
        // 2 too high. We could then use a simple compare and subtract loop to get the quotient
        // guess correct.
        // But we're going to correct the quotient with the estimate of the next byte instead.
        // When we overshoot the correct value we will subtract the estimate taken on the surplus
        // value, and otherwise we will add the estimate taken on the remainder value. This will
        // ultimately converge the quotient to the actual value of lhs / rhs

        // determine the initial guess for the quotient
        let bufLen = lhsLen - rhsLen;
        const buf = new Array<u8>(bufLen);
        const lhs16 = wasmtypes.uint16FromBytes(this.bytes.slice(lhsLen - 2));
        const rhs16 = rhs.bytes[rhsLen - 1] as u16;
        let res16 = lhs16 / rhs16;
        if (res16 > 0xff) {
            // res16 can be up to 0x0101, reduce guess to the nearest byte value
            // the estimate correction algorithm will take care of fixing this
            res16 = 0xff;
        }
        buf[bufLen - 1] = res16 as u8;
        const guess = ScBigInt.normalize(buf);

        // now see where this guess gets us when multiplying back
        const product = rhs.mul(guess);

        // compare the product to the original lhs to see where our guess get us
        const cmp = product.cmp(this);
        if (cmp == 0) {
            // lucky guess, exact match, and obviously exactly divisible by rhs without remainder
            return [guess, ScBigInt.zero];
        }

        if (cmp < 0) {
            // underestimated, correct our guess by adding the estimate on the remainder
            const guessRemainder = this.sub(product);
            const quoRem = guessRemainder.divModEstimate(rhs);
            return [guess.add(quoRem[0]), quoRem[1]];
        }

        // overestimated, correct guess by subtracting the estimate on the surplus
        const guessSurplus = product.sub(this);
        const quoRem = guessSurplus.divModEstimate(rhs);
        if (quoRem[1].isZero()) {
            // when the remainder is zero, we obviously have the exact match,
            // so we only subtract the surplus quotient
            return [guess.sub(quoRem[0]), ScBigInt.zero];
        }

        // When the remainder is nonzero, we subtract the surplus quotient, but since
        // we also need to subtract the remainder, we will need to correct the quotient
        // by subtracting one more. That will allow us to then subtract the surplus
        // remainder from the rhs to get the actual remainder.
        return [guess.sub(quoRem[0]).sub(ScBigInt.one), rhs.sub(quoRem[1])];
    }

    // divModMultiByte divides the lhs by a multi-byte rhs and returns quotient and remainder
    public divModMultiByte(rhs: ScBigInt): ScBigInt[] {
        // normalize first, shift divisor MSB until the high order bit is set
        // so that we get the best guess possible when dividing by MSB

        let msb = rhs.bytes[rhs.bytes.length - 1];
        if ((msb & 0x80) != 0) {
            // already normalized, no shifts necessary
            return this.divModEstimate(rhs);
        }

        // determine how far to shift
        let shift: u32 = 1;
        for (msb <<= 1; (msb & 0x80) == 0; msb <<= 1) {
            shift++;
        }

        // shift both lhs and rhs, that way the quotient will be the same
        // since lhs / rhs is equal to (lhs << shift) / (rhs << shift)
        const quoRem = this.shl(shift).divModEstimate(rhs.shl(shift));

        // note that remainder will be shifted, too, so shift back the remainder
        return [quoRem[0], quoRem[1].shr(shift)];
    }

    private divModSingleByte(value: u8): ScBigInt[] {
        const lhsLen = this.bytes.length;
        const buf = new Array<u8>(lhsLen);
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
            return ScBigInt.zero;
        }
        if (rhsLen == 1 && rhs.bytes[0] == 1) {
            // multiply by one, result lhs
            return ScBigInt.normalize(this.bytes);
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

    public shl(shift32: u32): ScBigInt {
        if (shift32 == 0) {
            return ScBigInt.normalize(this.bytes);
        }

        let whole_bytes = shift32 >> 3;
        let shift = (shift32 & 0x07) as u16;

        let lhs_len = this.bytes.length;
        let buf_len = lhs_len + whole_bytes + 1;
        let buf = new Array<u8>(buf_len);
        let word: u16 = 0;
        for (let i = lhs_len; i > 0; i--) {
            word = (word << 8) + (this.bytes[i - 1] as u16);
            buf_len -= 1;
            buf[buf_len] = (word >> (8 - shift)) as u8;
        }
        buf[buf_len - 1] = (word << shift) as u8;
        return ScBigInt.normalize(buf);
    }

    public shr(shift32: u32): ScBigInt {
        if (shift32 == 0) {
            return ScBigInt.normalize(this.bytes);
        }

        let whole_bytes = shift32 >> 3;
        let shift = (shift32 & 0x07) as u16;

        let lhs_len = this.bytes.length;
        if (whole_bytes >= (lhs_len as u32)) {
            return ScBigInt.zero;
        }

        let buf_len = lhs_len - whole_bytes;
        let buf = new Array<u8>(buf_len);
        let bytes = this.bytes.slice(whole_bytes);
        let word = (bytes[0] as u16) << 8;
        for (let i = 1; i < buf_len; i++) {
            word = (word >> 8) + ((bytes[i] as u16) << 8);
            buf[i - 1] = (word >> shift) as u8;
        }
        buf[buf_len - 1] = (word >> (8 + shift)) as u8;
        return ScBigInt.normalize(buf);
    }

    public sub(rhs: ScBigInt): ScBigInt {
        const cmp = this.cmp(rhs);
        if (cmp <= 0) {
            if (cmp < 0) {
                panic("subtraction underflow");
            }
            return ScBigInt.zero;
        }
        const lhsLen = this.bytes.length;
        const rhsLen = rhs.bytes.length;

        const buf = new Array<u8>(lhsLen);
        let borrow: u16 = 0;
        for (let i = 0; i < rhsLen; i++) {
            borrow += (this.bytes[i] as u16) - (rhs.bytes[i] as u16);
            buf[i] = borrow as u8;
            borrow = (borrow & 0xff00) | (borrow >> 8);
        }
        for (let i = rhsLen; i < lhsLen; i++) {
            borrow += this.bytes[i] as u16;
            buf[i] = borrow as u8;
            borrow = (borrow & 0xff00) | (borrow >> 8);
        }
        return ScBigInt.normalize(buf);
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return bigIntToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return bigIntToString(this);
    }

    public uint64(): u64 {
        const zeroes = wasmtypes.ScUint64Length - this.bytes.length;
        if (zeroes > wasmtypes.ScUint64Length) {
            panic("value exceeds Uint64");
        }
        const buf = this.bytes.concat(wasmtypes.zeroes(zeroes));
        return wasmtypes.uint64FromBytes(buf);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const quintillion = ScBigInt.fromUint64(1_000_000_000_000_000_000);

export function bigIntDecode(dec: wasmtypes.WasmDecoder): ScBigInt {
    const o = new ScBigInt();
    o.bytes = dec.bytes();
    return o;
}

export function bigIntEncode(enc: wasmtypes.WasmEncoder, value: ScBigInt): void {
    enc.bytes(value.bytes);
}

export function bigIntFromBytes(buf: u8[]): ScBigInt {
    const o = new ScBigInt();
    o.bytes = reverse(buf);
    return o;
}

export function bigIntToBytes(value: ScBigInt): u8[] {
    return reverse(value.bytes);
}

export function bigIntFromString(value: string): ScBigInt {
    // Uint64 fits 18 digits or 1 quintillion
    if (value.length <= 18) {
        return ScBigInt.fromUint64(wasmtypes.uint64FromString(value));
    }

    // build value 18 digits at a time
    const digits = value.length - 18;
    const lhs = bigIntFromString(value.slice(0, digits));
    const rhs = bigIntFromString(value.slice(digits));
    return lhs.mul(quintillion).add(rhs)
}

export function bigIntToString(value: ScBigInt): string {
    if (value.isUint64()) {
        return wasmtypes.uint64ToString(value.uint64());
    }
    const divMod = value.divMod(quintillion);
    const digits = wasmtypes.uint64ToString(divMod[1].uint64());
    const zeroes = "000000000000000000".slice(18 - digits.length);
    return bigIntToString(divMod[0]) + zeroes + digits;
}

// Stupid big.Int uses BigEndian byte encoding, so our external byte encoding should
// reflect this by reverse()-ing the byte order in BigIntFromBytes and BigIntToBytes
function reverse(bytes: u8[]): u8[] {
    let n = bytes.length;
    const buf = new Array<u8>(n);
    for (let i = 0; i < n; i++) {
        buf[n - 1 - i] = bytes[i];
    }
    return buf;
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
