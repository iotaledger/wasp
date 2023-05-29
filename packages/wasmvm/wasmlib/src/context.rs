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

impl ScFuncClientContext for ScFuncContext {}

impl ScViewClientContext for ScFuncContext {
    fn client_contract(&self, h_contract: ScHname) -> ScHname {
        h_contract
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// smart contract view interface which has only immutable access to state
#[derive(Clone, Copy)]
pub struct ScViewContext {}

// reuse shared part of interface
impl ScSandbox for ScViewContext {}

impl ScSandboxView for ScViewContext {}

impl ScViewClientContext for ScViewContext {
    fn client_contract(&self, h_contract: ScHname) -> ScHname {
        h_contract
    }
}
