// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const ZERO_U64: [u8; SC_UINT64_LENGTH] = [0; SC_UINT64_LENGTH];

#[derive(PartialEq, Clone, Eq, Hash)]
pub struct ScBigInt {
    bytes: Vec<u8>,
}

impl ScBigInt {
    pub fn zero() -> ScBigInt {
        ScBigInt::new()
    }

    pub fn one() -> ScBigInt {
        ScBigInt::from_uint64(1)
    }

    pub fn new() -> ScBigInt {
        ScBigInt { bytes: Vec::new() }
    }

    pub fn from_uint64(value: u64) -> ScBigInt {
        ScBigInt::normalize(&uint64_to_bytes(value))
    }

    fn normalize(buf: &[u8]) -> ScBigInt {
        let mut buf_len = buf.len();
        while buf_len > 0 && buf[buf_len - 1] == 0 {
            buf_len -= 1;
        }
        ScBigInt { bytes: buf[..buf_len].to_vec() }
    }

    pub fn add(&self, rhs: &ScBigInt) -> ScBigInt {
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();
        if lhs_len < rhs_len {
            // always add shorter value to longer value
            return rhs.add(self);
        }

        let mut buf: Vec<u8> = vec![0; lhs_len];
        let mut carry: u16 = 0;
        for i in 0..rhs_len {
            carry += (self.bytes[i] as u16) + (rhs.bytes[i] as u16);
            buf[i] = carry as u8;
            carry >>= 8;
        }
        for i in rhs_len..lhs_len {
            carry += self.bytes[i] as u16;
            buf[i] = carry as u8;
            carry >>= 8;
        }
        if carry != 0 {
            buf.push(1);
        }
        ScBigInt::normalize(&buf)
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
        if rhs.is_zero() {
            panic("divide by zero");
        }
        let cmp = self.cmp(rhs);
        if cmp <= 0 {
            if cmp < 0 {
                // divide by larger value, quo = 0, rem = lhs
                return (ScBigInt::new(), self.clone());
            }
            // divide equal values, quo = 1, rem = 0
            return (ScBigInt::from_uint64(1), ScBigInt::new());
        }
        if self.is_uint64() {
            // let standard uint64 type do the heavy lifting
            let lhs64 = self.uint64();
            let rhs64 = rhs.uint64();
            let div = ScBigInt::from_uint64(lhs64 / rhs64);
            return (div, ScBigInt::from_uint64(lhs64 % rhs64));
        }
        if rhs.bytes.len() == 1 {
            if rhs.bytes[0] == 1 {
                // divide by 1, quo = lhs, rem = 0
                return (self.clone(), ScBigInt::new());
            }
            return self.div_mod_single_byte(rhs.bytes[0]);
        }
        self.div_mod_multi_byte(&rhs)
    }

    // divModEstimate uses long division byte estimation and corrects on the remainder
    // by recursively estimating the next byte and discounting the result in the quotient
    fn div_mod_estimate(&self, rhs: &ScBigInt) -> (ScBigInt, ScBigInt) {
        // first filter out the simplest cases, when lengths are equal
        // that either results in a quotient of one or zero
        let lhs_len = self.bytes.len();
        let rhs_len = rhs.bytes.len();
        if lhs_len <= rhs_len {
            if self.cmp(rhs) < 0 {
                return (ScBigInt::zero(), self.clone());
            }
            return (ScBigInt::one(), self.sub(rhs));
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
        let buf_len = lhs_len - rhs_len;
        let mut buf: Vec<u8> = vec![0; buf_len];
        let lhs16 = uint16_from_bytes(&self.bytes[lhs_len - 2..]);
        let rhs16 = rhs.bytes[rhs_len - 1] as u16;
        let mut res16 = lhs16 / rhs16;
        if res16 > 0xff {
            // res16 can be up to 0x0101, reduce guess to the nearest byte value
            // the estimate correction algorithm will take care of fixing self
            res16 = 0xff;
        }
        buf[buf_len - 1] = res16 as u8;
        let guess = ScBigInt::normalize(&buf);

        // now see where self guess gets us when multiplying back
        let product = rhs.mul(&guess);

        // compare the product to the original lhs to see where our guess get us
        let cmp = product.cmp(self);
        if cmp == 0 {
            // lucky guess, exact match, and obviously exactly divisible by rhs without remainder
            return (guess, ScBigInt::zero());
        }

        if cmp < 0 {
            // underestimated, correct our guess by adding the estimate on the remainder
            let guess_remainder = self.sub(&product);
            let (quo, rem) = guess_remainder.div_mod_estimate(rhs);
            return (guess.add(&quo), rem);
        }

        // overestimated, correct guess by subtracting the estimate on the surplus
        let guess_surplus = product.sub(self);
        let (quo, rem) = guess_surplus.div_mod_estimate(rhs);
        if rem.is_zero() {
            // when the remainder is zero, we obviously have the exact match,
            // so we only subtract the surplus quotient
            return (guess.sub(&quo), ScBigInt::zero());
        }

        // When the remainder is nonzero, we subtract the surplus quotient, but since
        // we also need to subtract the remainder, we will need to correct the quotient
        // by subtracting one more. That will allow us to then subtract the surplus
        // remainder from the rhs to get the actual remainder.
        return (guess.sub(&quo).sub(&ScBigInt::one()), rhs.sub(&rem));
    }

    // divModMultiByte divides the lhs by a multi-byte rhs and returns quotient and remainder
    fn div_mod_multi_byte(&self, rhs: &ScBigInt) -> (ScBigInt, ScBigInt) {
        // normalize first, shift divisor MSB until the high order bit is set
        // so that we get the best guess possible when dividing by MSB

        let mut msb = rhs.bytes[rhs.bytes.len() - 1];
        if (msb & 0x80) != 0 {
            // already normalized, no shifts necessary
            return self.div_mod_estimate(rhs);
        }

        // determine how far to shift
        let mut shift: u32 = 1;
        msb <<= 1;
        while (msb & 0x80) == 0 {
            shift += 1;
            msb <<= 1;
        }

        // shift both lhs and rhs, that way the quotient will be the same
        // since lhs / rhs is equal to (lhs << shift) / (rhs << shift)
        let (quo, rem) = self.shl(shift).div_mod_estimate(&rhs.shl(shift));

        // note that remainder will be shifted, too, so shift back the remainder
        return (quo, rem.shr(shift));
    }

    fn div_mod_single_byte(&self, value: u8) -> (ScBigInt, ScBigInt) {
        let lhs_len = self.bytes.len();
        let mut buf: Vec<u8> = vec![0; lhs_len];
        let mut remain: u16 = 0;
        let rhs = value as u16;
        for i in (0..lhs_len).rev() {
            remain = (remain << 8) + (self.bytes[i] as u16);
            buf[i] = (remain / rhs) as u8;
            remain %= rhs;
        }
        (ScBigInt::normalize(&buf), ScBigInt::normalize(&[remain as u8]))
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
        if lhs_len + rhs_len <= SC_UINT64_LENGTH {
            return ScBigInt::from_uint64(self.uint64() * rhs.uint64());
        }
        if rhs_len == 0 {
            // multiply by zero, result zero
            return ScBigInt::new();
        }
        if rhs_len == 1 && rhs.bytes[0] == 1 {
            // multiply by one, result lhs
            return self.clone();
        }

        //TODO optimize by using u32 words instead of u8 words
        let mut buf: Vec<u8> = vec![0; lhs_len + rhs_len];
        for r in 0..rhs_len {
            let mut carry: u16 = 0;
            for l in 0..lhs_len {
                carry += (buf[l + r] as u16) + (self.bytes[l] as u16) * (rhs.bytes[r] as u16);
                buf[l + r] = carry as u8;
                carry >>= 8;
            }
            buf[r + lhs_len] = carry as u8;
        }
        ScBigInt::normalize(&buf)
    }

    pub fn shl(&self, shift: u32) -> ScBigInt {
        if shift == 0 {
            return self.clone();
        }

        let whole_bytes = (shift >> 3) as usize;
        let shift = shift & 0x07;

        let lhs_len = self.bytes.len();
        let mut buf_len = lhs_len + whole_bytes + 1;
        let mut buf: Vec<u8> = vec![0; buf_len];
        let mut word: u16 = 0;
        for i in (0..lhs_len).rev() {
            word = (word << 8) + (self.bytes[i] as u16);
            buf_len -= 1;
            buf[buf_len] = (word >> (8 - shift)) as u8;
        }
        buf[buf_len - 1] = (word << shift) as u8;
        ScBigInt::normalize(&buf)
    }

    pub fn shr(&self, shift: u32) -> ScBigInt {
        if shift == 0 {
            return self.clone();
        }

        let whole_bytes = (shift >> 3) as usize;
        let shift = shift & 0x07;

        let lhs_len = self.bytes.len();
        if whole_bytes >= lhs_len {
            return ScBigInt::new();
        }

        let buf_len = lhs_len - whole_bytes;
        let mut buf: Vec<u8> = vec![0; buf_len];
        let bytes = self.bytes[whole_bytes..].to_vec();
        let mut word = (bytes[0] as u16) << 8;
        for i in 1..buf_len {
            word = (word >> 8) + ((bytes[i] as u16) << 8);
            buf[i - 1] = (word >> shift) as u8;
        }
        buf[buf_len - 1] = (word >> (8 + shift)) as u8;
        ScBigInt::normalize(&buf)
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

        let mut buf: Vec<u8> = vec![0; lhs_len];
        let mut borrow: u16 = 0;
        for i in 0..rhs_len {
            borrow += (self.bytes[i] as u16) - (rhs.bytes[i] as u16);
            buf[i] = borrow as u8;
            borrow = (borrow & 0xff00) | (borrow >> 8);
        }
        for i in rhs_len..lhs_len {
            borrow += self.bytes[i] as u16;
            buf[i] = borrow as u8;
            borrow = (borrow & 0xff00) | (borrow >> 8);
        }
        ScBigInt::normalize(&buf)
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        big_int_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        big_int_to_string(self)
    }

    pub fn uint64(&self) -> u64 {
        let zeroes = SC_UINT64_LENGTH - self.bytes.len();
        if zeroes > SC_UINT64_LENGTH {
            panic("value exceeds Uint64");
        }
        let mut buf = self.bytes.clone();
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
    ScBigInt { bytes: reverse(buf) }
}

pub fn big_int_to_bytes(value: &ScBigInt) -> Vec<u8> {
    reverse(&value.bytes)
}

pub fn big_int_from_string(value: &str) -> ScBigInt {
    // Uint64 fits 18 digits or 1 quintillion
    if value.len() <= 18 {
        return ScBigInt::from_uint64(uint64_from_string(value));
    }

    // build value 18 digits at a time
    let digits = value.len() - 18;
    let quintillion = ScBigInt::from_uint64(1_000_000_000_000_000_000);
    let lhs = big_int_from_string(&value[..digits]);
    let rhs = big_int_from_string(&value[digits..]);
    lhs.mul(&quintillion).add(&rhs)
}

pub fn big_int_to_string(value: &ScBigInt) -> String {
    if value.is_uint64() {
        return uint64_to_string(value.uint64());
    }
    let (div, modulo) = value.div_mod(&ScBigInt::from_uint64(1_000_000_000_000_000_000));
    let digits = uint64_to_string(modulo.uint64());
    let zeroes = &"000000000000000000"[..18 - digits.len()];
    return big_int_to_string(&div) + zeroes + &digits;
}

// Stupid big.Int uses BigEndian byte encoding, so our external byte encoding should
// reflect this by reverse()-ing the byte order in BigIntFromBytes and BigIntToBytes
fn reverse(bytes: &[u8]) -> Vec<u8> {
    let n = bytes.len();
    let mut buf: Vec<u8> = vec![0; n];
    for i in 0..n {
        buf[n - 1 - i] = bytes[i];
    }
    buf
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
