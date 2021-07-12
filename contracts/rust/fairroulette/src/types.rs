// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

//@formatter:off
pub struct Bet {
    pub amount: i64,
    pub better: ScAgentID,
    pub number: i64,
}
//@formatter:on

impl Bet {
    pub fn from_bytes(bytes: &[u8]) -> Bet {
        let mut decode = BytesDecoder::new(bytes);
        Bet {
            amount: decode.int64(),
            better: decode.agent_id(),
            number: decode.int64(),
        }
    }

    pub fn to_bytes(&self) -> Vec<u8> {
        let mut encode = BytesEncoder::new();
        encode.int64(self.amount);
        encode.agent_id(&self.better);
        encode.int64(self.number);
        return encode.data();
    }
}
