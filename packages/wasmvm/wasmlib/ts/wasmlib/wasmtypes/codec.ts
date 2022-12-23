// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {stringFromBytes} from "./scstring";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// WasmDecoder decodes separate entities from a byte buffer
export class WasmDecoder {
    buf: Uint8Array;

    constructor(buf: Uint8Array) {
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
        this.buf = this.buf.subarray(1);
        return value;
    }

    // decodes the next variable sized slice of bytes from the byte buffer
    bytes(): Uint8Array {
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
    fixedBytes(size: u32): Uint8Array {
        if ((this.buf.length as u32) < size) {
            panic("insufficient fixed bytes");
        }
        let value = this.buf.slice(0, size);
        this.buf = this.buf.subarray(size);
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
    vliDecode(bits: i32): i32 {
        if (bits > 32) {
            panic("vliDecode");
        }
        return Number(this.vliDecode64(bits)) as i32;
    }

    // Variable Length Integer decoder, uses modified LEB128
    vliDecode64(bits: i32): i64 {
        let b = this.byte();
        const sign = b & 0x40;

        // first group of 6 bits
        let value = BigInt(b & 0x3f) as i64;
        let s = 6n;

        // while continuation bit is set
        for (; (b & 0x80) != 0; s += 7n) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.byte();
            value += (BigInt(b & 0x7f) << s) as i64;
        }

        // value was encoded as absolute value
        return (sign != 0) ? -value : value;
    }

    // Variable Length Unsigned decoder, uses ULEB128
    vluDecode(bits: i32): u32 {
        if (bits > 32) {
            panic("vluDecode");
        }
        return Number(this.vluDecode64(bits)) as u32;
    }

    // Variable Length Unsigned decoder, uses ULEB128
    vluDecode64(bits: i32): u64 {
        // first group of 7 bits
        let b = this.byte();
        let value = BigInt(b & 0x7f) as u64;
        let s = 7n;

        // while continuation bit is set
        for (; (b & 0x80) != 0; s += 7n) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.byte();
            value += (BigInt(b & 0x7f) << s) as u64;
        }

        return value;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// WasmEncoder encodes separate entities into a byte buffer
export class WasmEncoder {
    data: Uint8Array;

    // constructs an encoder
    constructor() {
        this.data = new Uint8Array(new ArrayBuffer(128), 0, 0);
    }

    // retrieves the encoded byte buffer
    buf(): Uint8Array {
        return this.data;
    }

    // encodes a single byte into the byte buffer
    byte(value: u8): WasmEncoder {
        const len = this.data.length;
        if (len == this.data.buffer.byteLength){
            const data = this.data;
            this.data = new Uint8Array(new ArrayBuffer(len * 2), 0, len);
            this.data.set(data);
        }
        this.data = new Uint8Array(this.data.buffer, 0, len+1);
        this.data[len] = value;
        return this;
    }

    // encodes a variable sized slice of bytes into the byte buffer
    bytes(value: Uint8Array): WasmEncoder {
        const length = value.length;
        this.vluEncode(length as u32);
        return this.fixedBytes(value, length as u32);
    }

    // encodes a fixed size slice of bytes into the byte buffer
    fixedBytes(value: Uint8Array, length: u32): WasmEncoder {
        if ((value.length as u32) != length) {
            panic("invalid fixed bytes length");
        }
        const len = this.data.length;
        if (len + value.length > this.data.buffer.byteLength){
            const data = this.data;
            this.data = new Uint8Array(new ArrayBuffer(len * 2 + value.length), 0, len);
            this.data.set(data);
        }
        this.data = new Uint8Array(this.data.buffer, 0, len+value.length);
        this.data.set(value, len);
        return this;
    }

    // Variable Length Integer encoder, uses modified LEB128
    vliEncode(value: i32): WasmEncoder {
        return this.vliEncode64(BigInt(value))
    }

    // Variable Length Integer encoder, uses modified LEB128
    vliEncode64(value: i64): WasmEncoder {
        // bit 7 is always continuation bit

        // first group: 6 bits of data plus sign bit
        // bit 6 encodes 0 as positive and 1 as negative
        let sign: u8 = 0x00;
        if (value < 0) {
            sign = 0x40;
            // encode absolute value
            value = -value;
        }

        let b = (Number(value & 0x3fn) | sign) as u8;
        value >>= 6n;

        // keep shifting until all bits are done
        while (value != 0n) {
            // emit with continuation bit
            this.byte(b | 0x80);

            // next group of 7 data bits
            b = Number(value & 0x7fn) as u8;
            value >>= 7n;
        }

        // emit without continuation bit to signal end
        this.byte(b);
        return this;
    }

    // Variable Length Unsigned encoder, uses ULEB128
    vluEncode(value: u32): WasmEncoder {
        return this.vluEncode64(BigInt(value))
    }

    // Variable Length Unsigned encoder, uses ULEB128
    vluEncode64(value: u64): WasmEncoder {
        // bit 7 is always continuation bit

        // first group of 7 data bits
        let b = Number(value & 0x7fn) as u8;
        value >>= 7n;

        // keep shifting until all bits are done
        while (value != 0n) {
            // emit with continuation bit
            this.byte(b | 0x80);

            // next group of 7 data bits
            b = Number(value & 0x7fn) as u8;
            value >>= 7n;
        }

        // emit without continuation bit to signal end
        this.byte(b);
        return this;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function concat(lhs: Uint8Array, rhs: Uint8Array): Uint8Array {
    const buf = new Uint8Array(lhs.length + rhs.length);
    buf.set(lhs);
    buf.set(rhs, lhs.length);
    return buf;
}

function has0xPrefix(s: string): boolean {
    return s.length >= 2 && s.charAt(0) == '0' && (s.charAt(1) == 'x' || s.charAt(1) == 'X')
}

export function hexDecode(hex: string): Uint8Array {
    if (!has0xPrefix(hex)) {
        panic("hex string missing 0x prefix")
    }
    const digits = hex.length - 2;
    if ((digits & 1) != 0) {
        panic("odd hex string length");
    }
    const buf = new Uint8Array(digits / 2);
    for (let i = 0; i < digits; i += 2) {
        buf[i / 2] = (hexer(hex.charCodeAt(i + 2) as u8) << 4) | hexer(hex.charCodeAt(i + 3) as u8)
    }
    return buf
}

export function hexEncode(buf: Uint8Array): string {
    const bytes = buf.length;
    const hex = new Uint8Array(bytes * 2);
    const alpha = (0x61 - 10) as u8;
    const digit = 0x30 as u8;

    for (let i = 0; i < bytes; i++) {
        const b: u8 = buf[i];
        const b1: u8 = b >> 4;
        hex[i * 2] = b1 + ((b1 > 9) ? alpha : digit);
        const b2: u8 = b & 0x0f;
        hex[i * 2 + 1] = b2 + ((b2 > 9) ? alpha : digit);
    }
    return "0x" + stringFromBytes(hex);
}

function hexer(hexDigit: u8): u8 {
    // '0' to '9'
    if (hexDigit >= 0x30 && hexDigit <= 0x39) {
        return hexDigit - 0x30;
    }
    // 'a' to 'f'
    if (hexDigit >= 0x61 && hexDigit <= 0x66) {
        return hexDigit - 0x61 + 10;
    }
    // 'A' to 'F'
    if (hexDigit >= 0x41 && hexDigit <= 0x46) {
        return hexDigit - 0x41 + 10;
    }
    panic("invalid hex digit");
    return 0;
}

export function intFromString(value: string, bits: u32): i32 {
    if (bits > 32) {
        panic("intFromString");
    }
    return Number(intFromString64(value, bits)) as i32
}

export function intFromString64(value: string, bits: u32): i64 {
    if (value.length == 0) {
        panic("intFromString: empty string");
    }
    let neg = false
    switch (value.charCodeAt(0)) {
        case 0x2b: // '+'
            value = value.substring(1);
            break;
        case 0x2d: // '-'
            neg = true;
            value = value.substring(1);
            break;
    }
    const uns = uintFromString64(value, bits);
    const cutoff = 2n**BigInt(bits - 1);
    if (neg) {
        if (neg && uns > cutoff) {
            panic("intFromString: min overflow");
        }
        return -uns as i64;
    }
    if (uns >= cutoff) {
        panic("intFromString: max overflow");
    }
    return uns as i64;
}

export function uintFromString(value: string, bits: u32): u32 {
    if (bits > 32) {
        panic("uintFromString");
    }
    return Number(uintFromString64(value, bits)) as u32
}

export function uintFromString64(value: string, bits: u32): u64 {
    if (value.length == 0) {
        panic("uintFromString: empty string");
    }
    const cutoff = 0xffffffffffffffffn / 10n + 1n;

    const maxVal = 2n**BigInt(bits) - 1n;

    let n = 0n as u64;
    for (let i = 0; i < value.length; i++) {
        const c = value.charCodeAt(i) as u32;
        if (c < 0x30 || c > 0x39) {
            panic("uintFromString: invalid digit");
        }
        if (n >= cutoff) {
            panic("uintFromString: cutoff overflow");
        }
        const n1 = n * 10n;
        n = n1 + BigInt(c - 0x30);
        if (n < n1 || n > maxVal) {
            panic("uintFromString: range overflow");
        }
    }
    return n;
}

export function zeroes(count: u32): Uint8Array {
    return new Uint8Array(count);
}
