// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::hashtypes::*;
use crate::host::*;

// decodes separate entities from a byte buffer
pub struct BytesDecoder<'a> {
    buf: &'a [u8],
}

impl BytesDecoder<'_> {
    // constructs a decoder
    pub fn new(data: &[u8]) -> BytesDecoder {
        if data.len() == 0 {
            panic("cannot decode empty byte array, use exist()");
        }
        BytesDecoder { buf: data }
    }

    // decodes an ScAddress from the byte buffer
    pub fn address(&mut self) -> ScAddress {
        ScAddress::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_ADDRESS as usize] as usize))
    }

    // decodes an ScAgentID from the byte buffer
    pub fn agent_id(&mut self) -> ScAgentID {
        ScAgentID::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_AGENT_ID as usize] as usize))
    }

    // decodes a bool from the byte buffer
    pub fn bool(&mut self) -> bool {
        self.uint8() != 0
    }

    // decodes the next variable length substring of bytes from the byte buffer
    pub fn bytes(&mut self) -> &[u8] {
        let length = self.uint32();
        self.fixed_bytes(length as usize)
    }

    // decodes an ScChainID from the byte buffer
    pub fn chain_id(&mut self) -> ScChainID {
        ScChainID::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_CHAIN_ID as usize] as usize))
    }

    // decodes an ScColor from the byte buffer
    pub fn color(&mut self) -> ScColor {
        ScColor::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_COLOR as usize] as usize))
    }

    // decodes the next fixed length substring of bytes from the byte buffer
    pub fn fixed_bytes(&mut self, size: usize) -> &[u8] {
        if self.buf.len() < size {
            panic("insufficient bytes");
        }
        let value = &self.buf[..size];
        self.buf = &self.buf[size..];
        value
    }

    // decodes an ScHash from the byte buffer
    pub fn hash(&mut self) -> ScHash {
        ScHash::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_HASH as usize] as usize))
    }

    // decodes an ScHname from the byte buffer
    pub fn hname(&mut self) -> ScHname {
        ScHname::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_HNAME as usize] as usize))
    }

    // decodes an int8 from the byte buffer
    pub fn int8(&mut self) -> i8 {
        self.uint8() as i8
    }

    // decodes an int16 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn int16(&mut self) -> i16 {
        self.vli_decode(16) as i16
    }

    // decodes an int32 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn int32(&mut self) -> i32 {
        self.vli_decode(32) as i32
    }

    // decodes an int64 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn int64(&mut self) -> i64 {
        self.vli_decode(64)
    }

    // decodes an ScRequestID from the byte buffer
    pub fn request_id(&mut self) -> ScRequestID {
        ScRequestID::from_bytes(self.fixed_bytes(TYPE_SIZES[TYPE_REQUEST_ID as usize] as usize))
    }

    // decodes an UTF-8 text string from the byte buffer
    pub fn string(&mut self) -> String {
        String::from_utf8_lossy(self.bytes()).to_string()
    }

    // decodes an uint8 from the byte buffer
    pub fn uint8(&mut self) -> u8 {
        if self.buf.len() == 0 {
            panic("insufficient bytes");
        }
        let value = self.buf[0];
        self.buf = &self.buf[1..];
        value
    }

    // decodes an uint16 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn uint16(&mut self) -> u16 {
        self.vlu_decode(16) as u16
    }

    // decodes an uint32 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn uint32(&mut self) -> u32 {
        self.vlu_decode(32) as u32
    }

    // decodes an uint64 from the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn uint64(&mut self) -> u64 {
        self.vlu_decode(64)
    }

    // vli (variable length integer) decoder
    fn vli_decode(&mut self, bits: i32) -> i64 {
        let mut b = self.uint8();
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
            b = self.uint8();
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
    fn vlu_decode(&mut self, bits: i32) -> u64 {
        // first group of 6 bits
        let mut b = self.uint8();
        let mut value = (b & 0x3f) as u64;
        let mut s = 7;

        // while continuation bit is set
        while (b & 0x80) != 0 {
            if s >= bits {
                panic("integer representation too long");
            }

            // next group of 7 bits
            b = self.uint8();
            value |= ((b & 0x7f) as u64) << s;
            s += 7;
        }

        value
    }
}

impl Drop for BytesDecoder<'_> {
    fn drop(&mut self) {
        if self.buf.len() != 0 {
            panic("extra bytes");
        }
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// encodes separate entities into a byte buffer
pub struct BytesEncoder {
    buf: Vec<u8>,
}

impl BytesEncoder {
    // constructs an encoder
    pub fn new() -> BytesEncoder {
        BytesEncoder { buf: Vec::new() }
    }

    // encodes an ScAddress into the byte buffer
    pub fn address(&mut self, value: &ScAddress) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_ADDRESS as usize] as usize)
    }

    // encodes an ScAgentID into the byte buffer
    pub fn agent_id(&mut self, value: &ScAgentID) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_AGENT_ID as usize] as usize)
    }

    // encodes a bool into the byte buffer
    pub fn bool(&mut self, value: bool) -> &BytesEncoder {
        self.uint8(value as u8)
    }

    // encodes a variable sized substring of bytes into the byte buffer
    pub fn bytes(&mut self, value: &[u8]) -> &BytesEncoder {
        let length = value.len();
        self.uint32(length as u32);
        self.fixed_bytes(value, length)
    }

    // encodes an ScChainID into the byte buffer
    pub fn chain_id(&mut self, value: &ScChainID) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_CHAIN_ID as usize] as usize)
    }

    // encodes an ScColor into the byte buffer
    pub fn color(&mut self, value: &ScColor) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_COLOR as usize] as usize)
    }

    // retrieve the encoded byte buffer
    pub fn data(&self) -> Vec<u8> {
        self.buf.clone()
    }

    // encodes a fixed sized substring of bytes into the byte buffer
    pub fn fixed_bytes(&mut self, value: &[u8], length: usize) -> &BytesEncoder {
        if value.len() != length as usize {
            panic("invalid fixed bytes length");
        }
        self.buf.extend_from_slice(value);
        self
    }

    // encodes an ScHash into the byte buffer
    pub fn hash(&mut self, value: &ScHash) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_HASH as usize] as usize)
    }

    // encodes an ScHname into the byte buffer
    pub fn hname(&mut self, value: ScHname) -> &BytesEncoder {
        self.fixed_bytes(&value.to_bytes(), TYPE_SIZES[TYPE_HNAME as usize] as usize)
    }

    // encodes an int8 into the byte buffer
    pub fn int8(&mut self, value: i8) -> &BytesEncoder {
        self.uint8(value as u8)
    }

    // encodes an int16 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn int16(&mut self, value: i16) -> &BytesEncoder {
        self.int64(value as i64)
    }

    // encodes an int32 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    pub fn int32(&mut self, value: i32) -> &BytesEncoder {
        self.int64(value as i64)
    }

    // encodes an int64 into the byte buffer
    // note that these are encoded using vli encoding to conserve space
    // vli (variable length integer) encoder
    pub fn int64(&mut self, mut value: i64) -> &BytesEncoder {
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

    // encodes an ScRequestID into the byte buffer
    pub fn request_id(&mut self, value: &ScRequestID) -> &BytesEncoder {
        self.fixed_bytes(value.to_bytes(), TYPE_SIZES[TYPE_REQUEST_ID as usize] as usize)
    }

    // encodes an UTF-8 text string into the byte buffer
    pub fn string(&mut self, value: &str) -> &BytesEncoder {
        self.bytes(value.as_bytes())
    }

    // encodes an uint8 into the byte buffer
    pub fn uint8(&mut self, value: u8) -> &BytesEncoder {
        self.buf.push(value);
        self
    }

    // encodes an uint16 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    pub fn uint16(&mut self, value: u16) -> &BytesEncoder {
        self.uint64(value as u64)
    }

    // encodes an uint32 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    pub fn uint32(&mut self, value: u32) -> &BytesEncoder {
        self.uint64(value as u64)
    }

    // encodes an uint64 into the byte buffer
    // note that these are encoded using vlu encoding to conserve space
    // vlu (variable length unsigned) encoder
    pub fn uint64(&mut self, mut value: u64) -> &BytesEncoder {
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
