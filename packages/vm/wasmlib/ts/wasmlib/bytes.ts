// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Convert} from "./convert";
import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import * as host from "./host";
import {panic} from "./host";

// decodes separate entities from a byte buffer
export class BytesDecoder {
    buf: u8[];

    // constructs a decoder
    constructor(data: u8[]) {
        if (data.length == 0) {
            panic("cannot decode empty byte array, use exist()");
        }
        this.buf = data;
    }

    // decodes an ScAddress from the byte buffer
    address(): ScAddress {
        return ScAddress.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_ADDRESS]));
    }

    // decodes an ScAgentID from the byte buffer
    agentID(): ScAgentID {
        return ScAgentID.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_AGENT_ID]));
    }

    // decodes a bool from the byte buffer
    bool(): boolean {
        return this.uint8() != 0;
    }

    // decodes the next substring of bytes from the byte buffer
    bytes(): u8[] {
        const length = this.uint32();
        return this.fixedBytes(length);
    }

    // decodes an ScChainID from the byte buffer
    chainID(): ScChainID {
        return ScChainID.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_CHAIN_ID]));
    }

    close(): void {
        if (this.buf.length != 0) {
            panic("extra bytes");
        }
    }

    // decodes an ScColor from the byte buffer
    color(): ScColor {
        return ScColor.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_COLOR]));
    }

    // decodes the next substring of bytes from the byte buffer
    fixedBytes(length: u32): u8[] {
        if (u32(this.buf.length) < length) {
            panic("insufficient bytes");
        }
        let value = this.buf.slice(0, length);
        this.buf = this.buf.slice(length);
        return value;
    }

    // decodes an ScHash from the byte buffer
    hash(): ScHash {
        return ScHash.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_HASH]));
    }

    // decodes an ScHname from the byte buffer
    hname(): ScHname {
        return ScHname.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_HNAME]));
    }

    // decodes an int8 from the byte buffer
    int8(): i8 {
        return this.uint8() as i8;
    }

    // decodes an int16 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    int16(): i16 {
        return this.vliDecode(16) as i16;
    }

    // decodes an int32 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    int32(): i32 {
        return this.vliDecode(32) as i32;
    }

    // decodes an int64 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    int64(): i64 {
        return this.vliDecode(64);
    }

    // decodes an ScRequestID from the byte buffer
    requestID(): ScRequestID {
        return ScRequestID.fromBytes(this.fixedBytes(host.TYPE_SIZES[host.TYPE_REQUEST_ID]));
    }

    // decodes an UTF-8 text string from the byte buffer
    string(): string {
        return Convert.toString(this.bytes());
    }

    // decodes an uint8 from the byte buffer
    uint8(): u8 {
        if (this.buf.length == 0) {
            panic("insufficient bytes");
        }
        return this.buf.shift();
    }

    // decodes an uint16 from the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    uint16(): u16 {
        return this.vluDecode(16) as u16;
    }

    // decodes an uint32 from the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    uint32(): u32 {
        return this.vluDecode(32) as u32;
    }

    // decodes an uint64 from the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    uint64(): u64 {
        return this.vluDecode(64);
    }

    // vli (variable length integer) decoder
    vliDecode(bits: i32): i64 {
        let b = this.uint8();
        const sign = b & 0x40;

        // first group of 6 bits
        let value = (b & 0x3f) as i64;
        let s = 6;

        // while continuation bit is set
        while ((b & 0x80) != 0) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.uint8();
            value |= ((b & 0x7f) as i64) << s;
            s += 7;
        }

        if (sign == 0) {
            // positive, sign bits are already zero
            return value;
        }

        // negative, extend sign bits
        return value | ((-1 as i64) << s);
    }

    // vlu decoder
    vluDecode(bits: i32): u64 {
        // first group of 7 bits
        let b = this.uint8();
        let value = (b & 0x7f) as u64;
        let s = 7;

        // while continuation bit is set
        while ((b & 0x80) != 0) {
            if (s >= bits) {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = this.uint8();
            value |= ((b & 0x7f) as u64) << s;
            s += 7;
        }

        return value;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// encodes separate entities into a byte buffer
export class BytesEncoder {
    buf: u8[];

    // constructs an encoder
    constructor() {
        this.buf = []
    }

    // encodes an ScAddress into the byte buffer
    address(value: ScAddress): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_ADDRESS]);
    }

    // encodes an ScAgentID into the byte buffer
    agentID(value: ScAgentID): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_AGENT_ID]);
    }

    // encodes a bool into the byte buffer
    bool(value: boolean): BytesEncoder {
        return this.uint8(value ? 1 : 0);
    }

    // encodes a substring of bytes into the byte buffer
    bytes(value: u8[]): BytesEncoder {
        const length = value.length;
        this.uint32(length);
        return this.fixedBytes(value, length);
    }

    // encodes an ScChainID into the byte buffer
    chainID(value: ScChainID): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_CHAIN_ID]);
    }

    // encodes an ScColor into the byte buffer
    color(value: ScColor): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_COLOR]);
    }

    // retrieve the encoded byte buffer
    data(): u8[] {
        return this.buf;
    }

    // encodes a substring of bytes into the byte buffer
    fixedBytes(value: u8[], length: u32): BytesEncoder {
        if (value.length != length) {
            panic("invalid fixed bytes length");
        }
        for (let i: u32 = 0; i < length; i++) {
            this.buf.push(value[i]);
        }
        return this;
    }

    // encodes an ScHash into the byte buffer
    hash(value: ScHash): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_HASH]);
    }

    // encodes an ScHname into the byte buffer
    hname(value: ScHname): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_HNAME]);
    }

    // encodes an int8 into the byte buffer
    int8(value: i8): BytesEncoder {
        return this.uint8(value as u8);
    }

    // encodes an int16 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    int16(value: i16): BytesEncoder {
        return this.int64(value as i64);
    }

    // encodes an int32 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    int32(value: i32): BytesEncoder {
        return this.int64(value as i64);
    }

    // encodes an int64 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    // vli (variable length integer) encoder
    int64(value: i64): BytesEncoder {
        // first group of 6 bits
        // 1st byte encodes 0 as positive in bit 6
        let b = (value as u8) & 0x3f;
        value >>= 6;

        let finalValue: i64 = 0;
        if (value < 0) {
            // encode negative value
            // 1st byte encodes 1 as negative in bit 6
            b |= 0x40;
            finalValue = -1;
        }

        // keep shifting until all bits are done
        while (value != finalValue) {
            // emit with continuation bit
            this.buf.push(b | 0x80);

            // next group of 7 bits
            b = (value as u8) & 0x7f;
            value >>= 7;
        }

        // emit without continuation bit
        this.buf.push(b);
        return this;
    }

    // encodes an ScRequestID into the byte buffer
    requestID(value: ScRequestID): BytesEncoder {
        return this.fixedBytes(value.toBytes(), host.TYPE_SIZES[host.TYPE_REQUEST_ID]);
    }

    // encodes an UTF-8 text string into the byte buffer
    string(value: string): BytesEncoder {
        return this.bytes(Convert.fromString(value));
    }

    // encodes an uint8 into the byte buffer
    uint8(value: u8): BytesEncoder {
        this.buf.push(value);
        return this;
    }

    // encodes an uint16 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    uint16(value: u16): BytesEncoder {
        return this.uint64(value as u64);
    }

    // encodes an uint32 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    uint32(value: u32): BytesEncoder {
        return this.uint64(value as u64);
    }

    // encodes an uint64 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    // vlu (variable length unsigned) encoder
    uint64(value: u64): BytesEncoder {
        // first group of 7 bits
        let b = (value as u8) & 0x7f;
        value >>= 7;

        // keep shifting until all bits are done
        while (value != 0) {
            // emit with continuation bit
            this.buf.push(b | 0x80);

            // next group of 7 bits
            b = (value as u8) & 0x7f;
            value >>= 7;
        }

        // emit without continuation bit
        this.buf.push(b);
        return this;
    }
}
