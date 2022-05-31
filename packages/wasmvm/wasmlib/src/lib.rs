// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

#![allow(dead_code)]

pub use assets::*;
pub use context::*;
pub use contract::*;
pub use dict::*;
pub use events::*;
pub use exports::*;
pub use sandbox::*;
pub use sandboxutils::*;
pub use wasmtypes::*;
pub use wasmvmhost::*;

pub mod assets;
pub mod context;
pub mod contract;
pub mod coreaccounts;
pub mod coreblob;
pub mod coreblocklog;
pub mod coregovernance;
pub mod coreroot;
pub mod dict;
pub mod events;
pub mod exports;
pub mod host;
pub mod sandbox;
pub mod sandboxutils;
pub mod wasmrequests;
pub mod wasmtypes;
pub mod wasmvmhost;

// When the `wee_alloc` feature is enabled,
// use `wee_alloc` as the global allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

// When the `console_error_panic_hook` feature is enabled, we can call the
// `set_panic_hook` function at least once during initialization, and then
// we will get better error messages if our code ever panics.
//
// For more details see
// https://github.com/rustwasm/console_error_panic_hook#readme
pub fn set_panic_hook() {
    #[cfg(feature = "console_error_panic_hook")]
        console_error_panic_hook::set_once();
}
