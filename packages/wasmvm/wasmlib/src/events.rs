// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use core::any::Any;
use core::marker::{Send, Sync};

use crate::*;

pub trait IEventHandlers: Any + Sync + Send {
    fn call_handler(&self, topic: &str, dec: &mut WasmDecoder);
    fn id(&self) -> u32;
}

static mut NEXT_ID: u32 = 0;

pub struct EventHandlers {}

impl EventHandlers {
    pub fn generate_id() -> u32 {
        unsafe {
            NEXT_ID += 1;
            NEXT_ID
        }
    }
}

pub struct EventEncoder {}

impl EventEncoder {
    pub fn new() -> WasmEncoder {
        let mut enc = WasmEncoder::new();
        uint64_encode(&mut enc, ScFuncContext {}.timestamp());
        enc
    }

    pub fn emit(topic: &str, enc: &WasmEncoder) {
        ScFuncContext {}.event(&(topic.to_string() + "|" + &hex_encode(&enc.buf())));
    }
}
