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
    fn init_func_call_context(&self) {
        panic!("init_view_func_context");
    }
}

impl ScViewCallContext for ScFuncContext {
    fn init_view_call_context(&self, _h_contract: ScHname) -> ScHname {
        panic!("init_view_call_context");
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
    fn init_view_call_context(&self, _h_contract: ScHname) -> ScHname {
        panic!("can_call_view");
    }
}
