// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::convert::TryInto;

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub const SC_COLOR_LENGTH: usize = 32;

#[derive(PartialEq, Clone, Copy, Eq, Hash)]
pub struct ScColor {
    id: [u8; SC_COLOR_LENGTH],
}

impl ScColor {
    pub const IOTA: ScColor = ScColor { id: [0x00; SC_COLOR_LENGTH] };
    pub const MINT: ScColor = ScColor { id: [0xff; SC_COLOR_LENGTH] };

    pub fn new(buf: &[u8]) -> ScColor {
        color_from_bytes(buf)
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        color_to_bytes(self)
    }

    pub fn to_string(&self) -> String {
        color_to_string(self)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub fn color_decode(dec: &mut WasmDecoder) -> ScColor {
    color_from_bytes_unchecked(&dec.fixed_bytes(SC_COLOR_LENGTH))
}

pub fn color_encode(enc: &mut WasmEncoder, value: &ScColor)  {
    enc.fixed_bytes(&value.to_bytes(), SC_COLOR_LENGTH);
}

pub fn color_from_bytes(buf: &[u8]) -> ScColor {
    if buf.len() == 0 {
        return ScColor { id: [0;SC_COLOR_LENGTH] };
    }
    if buf.len() != SC_COLOR_LENGTH {
        panic("invalid Color length");
    }
    ScColor { id: buf.try_into().expect("WTF?") }
}

pub fn color_to_bytes(value: &ScColor) -> Vec<u8> {
    value.id.to_vec()
}

pub fn color_to_string(value: &ScColor) -> String {
    // TODO standardize human readable string
    base58_encode(&value.id)
}

fn color_from_bytes_unchecked(buf: &[u8]) -> ScColor {
    ScColor { id: buf.try_into().expect("invalid Color length") }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

pub struct ScImmutableColor {
    proxy: Proxy
}

impl ScImmutableColor {
    pub fn new(proxy: Proxy) -> ScImmutableColor {
        ScImmutableColor { proxy }
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn to_string(&self) -> String {
        color_to_string(&self.value())
    }

    pub fn value(&self) -> ScColor {
        color_from_bytes(&self.proxy.get())
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScColor in host container
pub struct ScMutableColor {
    proxy: Proxy
}

impl ScMutableColor {
    pub fn new(proxy: Proxy) -> ScMutableColor {
        ScMutableColor { proxy }
    }

    pub fn delete(&self)  {
        self.proxy.delete();
    }

    pub fn exists(&self) -> bool {
        self.proxy.exists()
    }

    pub fn set_value(&self, value: &ScColor) {
        self.proxy.set(&color_to_bytes(&value));
    }

    pub fn to_string(&self) -> String {
        color_to_string(&self.value())
    }

    pub fn value(&self) -> ScColor {
        color_from_bytes(&self.proxy.get())
    }
}
