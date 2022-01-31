// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::hashtypes::*;
use crate::host::*;

pub struct WasmDecoder<'a> {
    buf: &'a [u8],
}

impl WasmDecoder<'_> {
    // constructs a decoder
    pub fn new(buf: &[u8]) -> WasmDecoder {
        if buf.len() == 0 {
            panic("empty decode buffer");
        }
        WasmDecoder { buf }
    }

    // decodes the next variable length substring of bytes from the byte buffer
    pub fn bytes(&mut self) -> &[u8] {
        let length = self.vlu_decode(32);
        self.fixed_bytes(length as usize)
    }

    // decodes an uint8 from the byte buffer
    pub fn byte(&mut self) -> u8 {
        if self.buf.len() == 0 {
            panic("insufficient bytes");
        }
        let value = self.buf[0];
        self.buf = &self.buf[1..];
        value
    }

    // decodes the next fixed length substring of bytes from the byte buffer
    pub fn fixed_bytes(&mut self, size: usize) -> &[u8] {
        if self.buf.len() < size {
            panic("insufficient fixed bytes");
        }
        let value = &self.buf[..size];
        self.buf = &self.buf[size..];
        value
    }

    // vli (variable length integer) decoder
    pub fn vli_decode(&mut self, bits: i32) -> i64 {
        let mut b = self.byte();
        let sign = b & 0x40;

        // first group of 6 bits
        let mut value = (b & 0x3f) as i64;
        let mut s = 6;

        // while continuation bit is set
        while (b & 0x80) != 0 {
            if s >= bits {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = self.byte();
            value |= ((b & 0x7f) as i64) << s;
            s += 7;
        }

        if sign == 0 {
            // positive, sign bits are already zero
            return value;
        }

        // negative, extend sign bits
        value | (-1_i64 << s)
    }

    // vlu (variable length unsigned) decoder
    pub fn vlu_decode(&mut self, bits: i32) -> u64 {
        // first group of 6 bits
        let mut b = self.byte();
        let mut value = (b & 0x7f) as u64;
        let mut s = 7;

        // while continuation bit is set
        while (b & 0x80) != 0 {
            if s >= bits {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = self.byte();
            value |= ((b & 0x7f) as u64) << s;
            s += 7;
        }

        value
    }
}

impl Drop for WasmDecoder<'_> {
    fn drop(&mut self) {
        if self.buf.len() != 0 {
            panic("extra bytes");
        }
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// encodes separate entities into a byte buffer
pub struct WasmEncoder {
    buf: Vec<u8>,
}

impl WasmEncoder {
    // constructs an encoder
    pub fn new() -> WasmEncoder {
        WasmEncoder { buf: Vec::new() }
    }

    // encodes an uint8 into the byte buffer
    pub fn byte(&mut self, value: u8) -> &WasmEncoder {
        self.buf.push(value);
        self
    }

    // encodes a variable sized substring of bytes into the byte buffer
    pub fn bytes(&mut self, value: &[u8]) -> &WasmEncoder {
        let length = value.len();
        self.vlu_encode(length as u64);
        self.fixed_bytes(value, length)
    }

    // retrieve the encoded byte buffer
    pub fn buf(&self) -> Vec<u8> {
        self.buf.clone()
    }

    // encodes a fixed sized substring of bytes into the byte buffer
    pub fn fixed_bytes(&mut self, value: &[u8], length: usize) -> &WasmEncoder {
        if value.len() != length as usize {
            panic("invalid fixed bytes length");
        }
        self.buf.extend_from_slice(value);
        self
    }

    // encodes an int64 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    // vli (variable length integer) encoder
    pub fn vli_encode(&mut self, mut value: i64) -> &WasmEncoder {
        // first group of 6 bits
        // 1st byte encodes 0 as positive in bit 6
        let mut b = value as u8 & 0x3f;
        value >>= 6;

        let mut final_value = 0_i64;
        if value < 0 {
            // encode negative value
            // 1st byte encodes 1 as negative in bit 6
            b |= 0x40;
            final_value = -1_i64;
        }

        // keep shifting until all bits are done
        while value != final_value {
            // emit with continuation bit
            self.buf.push(b|0x80);

            // next group of 7 bits
            b = value as u8 & 0x7f;
            value >>= 7;
        }

        // emit without continuation bit
        self.buf.push(b);
        self
    }

    // encodes an uint64 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    // vlu (variable length unsigned) encoder
    pub fn vlu_encode(&mut self, mut value: u64) -> &WasmEncoder {
        // first group of 7 bits
        // 1st byte encodes 0 as positive in bit 6
        let mut b = value as u8;
        value >>= 7;

        // keep shifting until all bits are done
        while value != 0 {
            // emit with continuation bit
            self.buf.push(b|0x80);

            // next group of 7 bits
            b = value as u8;
            value >>= 7;
        }

        // emit without continuation bit
        self.buf.push(b);
        self
    }
}

pub fn base58_encode(buf: &[u8]) -> String {
    hex(buf)
}

pub fn hex(buf: &[u8]) -> String {
    //TODO
    "".to_string()
}
