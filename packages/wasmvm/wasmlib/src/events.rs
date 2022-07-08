// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::*;

// encodes separate entities into a byte buffer
pub struct EventEncoder {
    event: String,
}

impl EventEncoder {
    pub fn new(event_name: &str) -> EventEncoder {
        let mut e = EventEncoder {
            event: event_name.to_string(),
        };
        let timestamp = ScFuncContext {}.timestamp();
        e.encode(&uint64_to_string(timestamp / 1_000_000_000));
        e
    }

    pub fn emit(&self) {
        ScFuncContext {}.event(&self.event);
    }

    pub fn encode(&mut self, value: &str) -> &EventEncoder {
        let mut replaced_value = str::replace(value, "~", "~~");
        replaced_value = str::replace(&replaced_value, "|", "~/");
        replaced_value = str::replace(&replaced_value, " ", "~_");
        self.event += "|";
        self.event += &replaced_value;
        self
    }
}
