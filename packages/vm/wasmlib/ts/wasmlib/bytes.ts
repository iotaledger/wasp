// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Convert} from "./convert";
import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import {panic} from "./host";

// decodes separate entities from a byte buffer
export class BytesDecoder {
    buf: u8[];

    // constructs a decoder
    constructor(data: u8[]) {
        this.buf = data;
    }

    // decodes an ScAddress from the byte buffer
    address(): ScAddress {
        return ScAddress.fromBytes(this.bytes());
    }

    // decodes an ScAgentID from the byte buffer
    agentID(): ScAgentID {
        return ScAgentID.fromBytes(this.bytes());
    }

    // decodes the next substring of bytes from the byte buffer
    bytes(): u8[] {
        let size = this.int32();
        if (this.buf.length < size) {
            panic("insufficient bytes");
        }
        let value = this.buf.slice(0, size);
        this.buf = this.buf.slice(size);
        return value;
    }

    // decodes an ScChainID from the byte buffer
    chainID(): ScChainID {
        return ScChainID.fromBytes(this.bytes());
    }

    // decodes an ScColor from the byte buffer
    color(): ScColor {
        return ScColor.fromBytes(this.bytes());
    }

    // decodes an ScHash from the byte buffer
    hash(): ScHash {
        return ScHash.fromBytes(this.bytes());
    }

    // decodes an ScHname from the byte buffer
    hname(): ScHname {
        return ScHname.fromBytes(this.bytes());
    }

    // decodes an int16 from the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int16(): i16 {
        return this.leb128Decode(16) as i16;
    }

    // decodes an int32 from the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int32(): i32 {
        return this.leb128Decode(32) as i32;
    }

    // decodes an int64 from the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int64(): i64 {
        return this.leb128Decode(64);
    }

    // leb128 decoder
    leb128Decode(bits: i32): i64 {
        let val: i64 = 0;
        let s = 0;
        for (; ;) {
            if (this.buf.length == 0) {
                panic("leb128Decode: insufficient bytes");
            }
            let b = this.buf.shift() as i8;
            val |= ((b & 0x7f) as i64) << s;

            // termination bit set?
            if ((b & -0x80) == 0) {
                if ((((val >> s) as i8) & 0x7f) != (b & 0x7f)) {
                    panic("integer too large");
                }

                // extend int7 sign to int8
                b |= (b & 0x40) << 1;

                // extend int8 sign to int64
                val |= ((b as i64) << s);
                break;
            }
            s += 7;
            if (s >= bits) {
                panic("integer representation too long");
            }
        }
        return val;
    }

    // decodes an ScRequestID from the byte buffer
    requestID(): ScRequestID {
        return ScRequestID.fromBytes(this.bytes());
    }

    // decodes an UTF-8 text string from the byte buffer
    string(): string {
        return Convert.toString(this.bytes());
    }

    close(): void {
        if (this.buf.length != 0) {
            panic("extra bytes");
        }
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
        this.bytes(value.toBytes());
        return this;
    }

    // encodes an ScAgentID into the byte buffer
    agentID(value: ScAgentID): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // encodes a substring of bytes into the byte buffer
    bytes(value: u8[]): BytesEncoder {
        this.int32(value.length);
        for (let i = 0; i < value.length; i++) {
            this.buf.push(value[i]);
        }
        return this;
    }

    // encodes an ScChainID into the byte buffer
    chainID(value: ScChainID): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // encodes an ScColor into the byte buffer
    color(value: ScColor): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // retrieve the encoded byte buffer
    data(): u8[] {
        return this.buf;
    }

    // encodes an ScHash into the byte buffer
    hash(value: ScHash): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // encodes an ScHname into the byte buffer
    hname(value: ScHname): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // encodes an int16 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int16(val: i16): BytesEncoder {
        return this.leb128Encode(val as i64);
    }

    // encodes an int32 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int32(val: i32): BytesEncoder {
        return this.leb128Encode(val as i64);
    }

    // encodes an int64 into the byte buffer
    // note that these are encoded using leb128 encoding to conserve space
    int64(val: i64): BytesEncoder {
        return this.leb128Encode(val);
    }

    // leb128 encoder
    leb128Encode(val: i64): BytesEncoder {
        for (; ;) {
            let b = val as u8;
            let s = b & 0x40;
            val >>= 7;
            if ((val == 0 && s == 0) || (val == -1 && s != 0)) {
                this.buf.push(b & 0x7f);
                break;
            }
            this.buf.push(b | 0x80);
        }
        return this;
    }

    // encodes an ScRequestID into the byte buffer
    requestID(value: ScRequestID): BytesEncoder {
        this.bytes(value.toBytes());
        return this;
    }

    // encodes an UTF-8 text string into the byte buffer
    string(value: string): BytesEncoder {
        this.bytes(Convert.fromString(value));
        return this;
    }
}
