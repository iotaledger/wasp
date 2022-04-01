// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ZERO_U64: [u8;SC_UINT64_LENGTH] = [0;SC_UINT64_LENGTH];

#[derive(PartialEq, Clone, Eq, Hash)]
pub struct ScBigInt {
    bytes: Vec<u8>,
}

impl ScBigInt {
    pub fn new() -> ScBigInt {
        ScBigInt { bytes: Vec::new() }
    }

    pub fn from_uint64(value: u64) -> ScBigInt {
        let mut big_int = ScBigInt::new();
        big_int.set_uint64(value);
        big_int
    }

    pub fn add(&self, rhs: &ScBigInt) -> ScBigInt {
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();
        if lhs_len < rhs_len {
            // always add shorter value to longer value
            return rhs.add(self);
        }

        let mut res = self.clone();
        let mut carry: u16 = 0;
        for i in 0..rhs_len {
            carry += res.bytes[i] as u16 + rhs.bytes[i] as u16;
            res.bytes[i] = carry as u8;
            carry >>= 8;
        }
        if carry != 0 {
            for i in rhs_len..lhs_len {
                carry += res.bytes[i] as u16;
                res.bytes[i] = carry as u8;
                carry >>= 8;
                if carry == 0 {
                    return res;
                }
            }
            res.bytes.push(1);
        }
        res
    }

    pub fn cmp(&self, rhs: &ScBigInt) -> i8 {
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();
        if lhs_len != rhs_len {
            if lhs_len > rhs_len {
                return 1;
            }
            return -1;
        }
        for i in (0..lhs_len).rev() {
            let lhs_byte = self.bytes[i];
            let rhs_byte = rhs.bytes[i];
            if lhs_byte != rhs_byte {
                if lhs_byte > rhs_byte {
                    return 1;
                }
                return -1;
            }
        }
        0
    }

    pub fn div(&self, rhs: &ScBigInt) -> ScBigInt {
        let (d, _m) = self.div_mod(rhs);
        d
    }

    pub fn div_mod(&self, rhs: &ScBigInt) -> (ScBigInt, ScBigInt) {
        panic("implement DivMod");
        (self.clone(), rhs.clone())
    }

    pub fn is_uint64(&self) -> bool {
        self.bytes.len() <= SC_UINT64_LENGTH
    }

    pub fn is_zero(&self) -> bool {
        self.bytes.len() == 0
    }

    pub fn modulo(&self, rhs: &ScBigInt) -> ScBigInt {
        let (_d, m) = self.div_mod(rhs);
        m
    }

    pub fn mul(&self, rhs: &ScBigInt) -> ScBigInt {
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();
        if lhs_len < rhs_len {
            // always multiply bigger value by smaller value
            return rhs.mul(self);
        }
        panic("implement Mul");
        self.clone()
    }

    fn normalize(&mut self) {
        let mut buf_len = self.bytes.len();
        while buf_len > 0 && self.bytes[buf_len - 1] == 0 {
            buf_len -= 1;
        }
        self.bytes.truncate(buf_len);
    }

    pub fn set_uint64(&mut self, value: u64) {
        self.bytes = uint64_to_bytes(value);
        self.normalize();
    }

    pub fn sub(&self, rhs: &ScBigInt) -> ScBigInt {
        let cmp = self.cmp(rhs);
        if cmp <= 0 {
            if cmp < 0 {
                panic("subtraction underflow");
            }
            return ScBigInt::new();
        }
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();

        let mut res = self.clone();
        let mut borrow: u16 = 0;
        for i in 0..rhs_len {
            borrow += res.bytes[i] as u16 - rhs.bytes[i] as u16;
            res.bytes[i] = borrow as u8;
            borrow >>= 8;
        }
        if borrow != 0 {
            for i in rhs_len..lhs_len {
                borrow += res.bytes[i] as u16;
                res.bytes[i] = borrow as u8;
                borrow >>= 8;
                if borrow == 0 {
                    res.normalize();
                    return res;
                }
            }
        }
        res.normalize();
        res
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        big_int_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        big_int_to_string(self)
    }

    pub fn uint64(&self) -> u64 {
        let zeroes = SC_UINT64_LENGTH-self.bytes.len();
        if zeroes > SC_UINT64_LENGTH {
            panic("value exceeds Uint64");
        }
        let mut buf = big_int_to_bytes(self);
        buf.extend_from_slice(&ZERO_U64[..zeroes]);
        uint64_from_bytes(&buf)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn big_int_decode(dec: &mut WasmDecoder) -> ScBigInt {
    ScBigInt { bytes: dec.bytes() }
}

pub fn big_int_encode(enc: &mut WasmEncoder, value: &ScBigInt) {
    enc.bytes(&value.bytes);
}

pub fn big_int_from_bytes(buf: &[u8]) -> ScBigInt {
    ScBigInt { bytes: buf.to_vec() }
}

pub fn big_int_to_bytes(value: &ScBigInt) -> Vec<u8> {
    value.bytes.to_vec()
}

pub fn big_int_to_string(_value: &ScBigInt) -> String {
    // TODO standardize human readable string
    panic("implement BigIntToString");
    return "".to_string();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableBigInt {
    proxy: Proxy,
}

impl ScImmutableBigInt {
    pub fn new(proxy: Proxy) -> ScImmutableBigInt {
        ScImmutableBigInt { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        big_int_to_string(&self.value())
    }

    pub fn value(&self) -> ScBigInt {
        big_int_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScBigInt in host container
pub struct ScMutableBigInt {
    proxy: Proxy,
}

impl ScMutableBigInt {
    pub fn new(proxy: Proxy) -> ScMutableBigInt {
        ScMutableBigInt { proxy }
    }

    pub fn delete(&self) {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScBigInt) {
        self.proxy.set(&big_int_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        big_int_to_string(&self.value())
    }

    pub fn value(&self) -> ScBigInt {
        big_int_from_bytes(&self.proxy.get())
    }
}
