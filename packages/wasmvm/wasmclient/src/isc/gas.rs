// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pub const MAX_GAS_PER_BLOCK: u64 = 1_000_000_000;
pub const MIN_GAS_PER_REQUEST: u64 = 10000;
pub const MAX_GAS_PER_REQUEST: u64 = MAX_GAS_PER_BLOCK / 20;
pub const MAX_GAS_EXTERNAL_VIEW_CALL: u64 = MIN_GAS_PER_REQUEST;
