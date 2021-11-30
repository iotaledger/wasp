// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::context::*;
use crate::hashtypes::*;
use crate::keys::*;

// encodes separate entities into a byte buffer
pub struct EventEncoder {
    event: String,
}

impl EventEncoder {
    // constructs an encoder
    pub fn new(event_name: &str) -> EventEncoder {
        let mut e = EventEncoder { event: event_name.to_string() };
        let timestamp = ROOT.get_int64(&KEY_TIMESTAMP).value();
        e.int64(timestamp / 1_000_000_000);
        e
    }

    // encodes an ScAddress into the byte buffer
    pub fn address(&mut self, value: &ScAddress) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an ScAgentID into the byte buffer
    pub fn agent_id(&mut self, value: &ScAgentID) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes a Bool as 0/1 into the byte buffer
    pub fn bool(&mut self, value: bool) -> &EventEncoder {
        self.uint8(value as u8)
    }

    // encodes a substring of bytes into the byte buffer
    pub fn bytes(&mut self, value: &[u8]) -> &EventEncoder {
        self.string(&base58_encode(value))
    }

    // encodes an ScChainID into the byte buffer
    pub fn chain_id(&mut self, value: &ScChainID) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an ScColor into the byte buffer
    pub fn color(&mut self, value: &ScColor) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // retrieve the encoded byte buffer
    pub fn emit(&self) {
        ROOT.get_string(&KEY_EVENT).set_value(&self.event);
    }

    // encodes an ScHash into the byte buffer
    pub fn hash(&mut self, value: &ScHash) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an ScHname into the byte buffer
    pub fn hname(&mut self, value: ScHname) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Int8 into the byte buffer
    pub fn int8(&mut self, value: i8) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Int16 into the byte buffer
    pub fn int16(&mut self, value: i16) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Int32 into the byte buffer
    pub fn int32(&mut self, value: i32) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Int64 into the byte buffer
    pub fn int64(&mut self, value: i64) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an ScRequestID into the byte buffer
    pub fn request_id(&mut self, value: &ScRequestID) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an UTF-8 text string into the byte buffer
    pub fn string(&mut self, value: &str) -> &EventEncoder {
        self.event += &("|".to_owned() + value);
        self
    }

    // encodes an Uint8 into the byte buffer
    pub fn uint8(&mut self, value: u8) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Uint16 into the byte buffer
    pub fn uint16(&mut self, value: u16) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Uint32 into the byte buffer
    pub fn uint32(&mut self, value: u32) -> &EventEncoder {
        self.string(&value.to_string())
    }

    // encodes an Uint64 into the byte buffer
    pub fn uint64(&mut self, value: u64) -> &EventEncoder {
        self.string(&value.to_string())
    }
}
