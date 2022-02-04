// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// encapsulates standard host entities into a simple interface

use crate::*;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract interface with mutable access to state
#[derive(Clone, Copy)]
pub struct ScFuncContext {}

// reuse shared part of interface
impl ScSandbox for ScFuncContext {}
impl ScSandboxFunc for ScFuncContext {}

impl ScFuncCallContext for ScFuncContext {
    fn can_call_func(&self) {
        panic!("can_call_func");
    }
}

impl ScViewCallContext for ScFuncContext {
    fn can_call_view(&self) {
        panic!("can_call_view");
    }
}

impl ScFuncContext {
    pub fn incoming(&self) -> ScBalances {
        self.incoming_transfer()
    }

    pub fn transfer_to_address(&self, address: &ScAddress, transfer: ScTransfers) {
        self.send(address, &transfer)
    }
}
// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
#[derive(Clone, Copy)]
pub struct ScViewContext {}

// reuse shared part of interface
impl ScSandbox for ScViewContext {}
impl ScSandboxView for ScViewContext {}

impl ScViewCallContext for ScViewContext {
    fn can_call_view(&self) {
        panic!("can_call_view");
    }
}
