// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {sandbox} from "../host";
import {FnUtilsBase58Encode, panic} from "../sandbox";
import * as wasmtypes from "./index";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// WasmDecoder decodes separate entities from a byte buffer
export class WasmDecoder {
    buf: u8[];

    constructor(buf: u8[]) {
        if (buf.length == 0) {
            panic("empty decode buffer");
        }
        this.buf = buf;
    }

    // decodes the next byte from the byte buffer
    byte(): u8 {
        if (this.buf.length == 0) {
            panic("insufficient bytes");
        }
        const value = this.buf[0];
        this.buf = this.buf.slice(1);
        return value;
    }

    // decodes the next variable sized slice of bytes from the byte buffer
    bytes(): u8[] {
        const length = this.vluDecode(32) as u32;
        return this.fixedBytes(length);
    }

    // finalizes decoding by panicking if any bytes remain in the byte buffer
    close(): void {
        if (this.buf.length != 0) {
            panic("extra bytes");
        }
    }

    // decodes the next fixed size slice of bytes from the byte buffer
    fixedBytes(size: u32): u8[] {
        if ((this.buf.length as u32) < size) {
            panic("insufficient fixed bytes");
        }
        let value = this.buf.slice(0, size);
        this.buf = this.buf.slice(size);
        return value;
    }

    // peeks at the next byte in the byte buffer
    peek(): u8 {
        if (this.buf.length == 0) {
            panic("insufficient peek bytes");
        }
        return this.buf[0];
    }

    // Variable Length Integer decoder, uses modified LEB128
    vliDecode(bits: i32): i64 {
        let b = this.byte();
        const sign = b & 0x40;

        // first group of 6 bits
        let value = (b & 0x3f) as i64;
        let s = 6;

        // while continuation bit is set
        for (; (b & 0x80) != 0; s += 7) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.byte();
            value |= ((b & 0x7f) as i64) << s;
        }

        if (sign == 0) {
            // positive, sign bits are already zero
            return value;
        }

        // negative, extend sign bits
        return value | ((-1 as i64) << s);
    }

    // Variable Length Unsigned decoder, uses ULEB128
    vluDecode(bits: i32): u64 {
        // first group of 7 bits
        let b = this.byte();
        let value = (b & 0x7f) as u64;
        let s = 7;

        // while continuation bit is set
        for (; (b & 0x80) != 0; s += 7) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.byte();
            value |= ((b & 0x7f) as u64) << s;
        }

        return value;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// WasmEncoder encodes separate entities into a byte buffer
export class WasmEncoder {
    data: u8[];

    // constructs an encoder
    constructor() {
        this.data = [];
    }

    // retrieves the encoded byte buffer
    buf(): u8[] {
        return this.data;
    }

    // encodes a single byte into the byte buffer
    byte(value: u8): wasmtypes.WasmEncoder {
        this.data.push(value);
        return this;
    }

    // encodes a variable sized slice of bytes into the byte buffer
    bytes(value: u8[]): wasmtypes.WasmEncoder {
        const length = value.length;
        this.vluEncode(length as u64);
        return this.fixedBytes(value, length as u32);
    }

    // encodes a fixed size slice of bytes into the byte buffer
    fixedBytes(value: u8[], length: u32): wasmtypes.WasmEncoder {
        if ((value.length as u32) != length) {
            panic("invalid fixed bytes length");
        }
        this.data = this.data.concat(value);
        return this;
    }

    // Variable Length Integer encoder, uses modified LEB128
    vliEncode(value: i64): wasmtypes.WasmEncoder {
        // bit 7 is always continuation bit

        // first group: 6 bits of data plus sign bit
        // bit 6 encodes 0 as positive and 1 as negative
        let b = (value as u8) & 0x3f;
        value >>= 6;

        let finalValue: i64 = 0;
        if (value < 0) {
            // 1st byte encodes 1 as negative in bit 6
            b |= 0x40;
            // negative value, start with all high bits set to 1
            finalValue = -1;
        }

        // keep shifting until all bits are done
        while (value != finalValue) {
            // emit with continuation bit
            this.data.push(b | 0x80);

            // next group of 7 data bits
            b = (value as u8) & 0x7f;
            value >>= 7;
        }

        // emit without continuation bit to signal end
        this.data.push(b);
        return this;
    }

    // Variable Length Unsigned encoder, uses ULEB128
    vluEncode(value: u64): wasmtypes.WasmEncoder {
        // bit 7 is always continuation bit

        // first group of 7 data bits
        let b = (value as u8) & 0x7f;
        value >>= 7;

        // keep shifting until all bits are done
        while (value != 0) {
            // emit with continuation bit
            this.data.push(b | 0x80);

            // next group of 7 data bits
            b = (value as u8) & 0x7f;
            value >>= 7;
        }

        // emit without continuation bit to signal end
        this.data.push(b);
        return this;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// wrapper for simplified use by hashtypes
export function base58Encode(buf: u8[]): string {
    return wasmtypes.stringFromBytes(sandbox(FnUtilsBase58Encode, buf));
}

export function zeroes(count: u32): u8[] {
    const buf: u8[] = new Array(count);
    buf.fill(0);
    return buf;
}
