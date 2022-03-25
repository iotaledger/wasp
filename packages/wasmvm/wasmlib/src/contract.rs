// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// base contract objects

use std::rc::Rc;

use crate::*;
use crate::host::*;

pub trait ScFuncCallContext {
    fn can_call_func(&self);
}

pub trait ScViewCallContext {
    fn can_call_view(&self);
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

#[derive(Clone)]
pub struct ScView {
    h_contract: ScHname,
    h_function: ScHname,
    params: Rc<ScDict>,
    results: Rc<ScDict>,
}

impl ScView {
    pub fn new(h_contract: ScHname, h_function: ScHname) -> ScView {
        ScView {
            h_contract: h_contract,
            h_function: h_function,
            params: Rc::new(ScDict::new(&[])),
            results: Rc::new(ScDict::new(&[])),
        }
    }

    pub fn link_params(proxy: &mut Proxy, view: &ScView) {
        Proxy::link(proxy, &view.params);
    }

    pub fn link_results(proxy: &mut Proxy, view: &ScView) {
        Proxy::link(proxy, &view.results);
    }

    pub fn call(&self) {
        self.call_with_transfer(None);
    }

    fn call_with_transfer(&self, transfer: Option<ScAssets>) {
        let mut req = wasmrequests::CallRequest {
            contract: self.h_contract,
            function: self.h_function,
            params: self.params.to_bytes(),
            transfer: vec![0; SC_UINT32_LENGTH],
        };
        if let Some(transfer) = transfer {
            req.transfer = transfer.to_bytes();
        }
        let res = sandbox(FN_CALL, &req.to_bytes());
        self.results.copy(&res);
    }

    pub fn of_contract(&self, h_contract: ScHname) -> ScView {
        let mut ret = self.clone();
        ret.h_contract = h_contract;
        ret
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

#[derive(Clone)]
pub struct ScInitFunc {
    view: ScView,
}

impl ScInitFunc {
    pub fn new(h_contract: ScHname, h_function: ScHname) -> ScInitFunc {
        ScInitFunc {
            view: ScView::new(h_contract, h_function),
        }
    }

    pub fn link_params(proxy: &mut Proxy, func: &ScInitFunc) {
        ScView::link_params(proxy, &func.view);
    }

    pub fn link_results(proxy: &mut Proxy, func: &ScInitFunc) {
        ScView::link_results(proxy, &func.view);
    }

    pub fn call(&self) {
        panic("cannot call init")
    }

    pub fn of_contract(&self, h_contract: ScHname) -> ScInitFunc {
        let mut ret = self.clone();
        ret.view.h_contract = h_contract;
        ret
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

#[derive(Clone)]
pub struct ScFunc {
    pub view: ScView,
    delay: u32,
    transfer: ScAssets,
}

impl ScFunc {
    pub fn new(h_contract: ScHname, h_function: ScHname) -> ScFunc {
        ScFunc {
            view: ScView::new(h_contract, h_function),
            delay: 0,
            transfer: ScAssets::new(&[]),
        }
    }

    pub fn link_params(proxy: &mut Proxy, func: &ScFunc) {
        ScView::link_params(proxy, &func.view);
    }

    pub fn link_results(proxy: &mut Proxy, func: &ScFunc) {
        ScView::link_results(proxy, &func.view);
    }

    pub fn call(&self) {
        if self.delay != 0 {
            panic("cannot delay a call")
        }
        self.view.call_with_transfer(Some(self.transfer.clone()));
    }

    pub fn delay(&self, seconds: u32) -> ScFunc {
        let mut ret = self.clone();
        ret.delay = seconds;
        ret
    }

    pub fn of_contract(&self, h_contract: ScHname) -> ScFunc {
        let mut ret = self.clone();
        ret.view.h_contract = h_contract;
        ret
    }

    pub fn post(&self) {
        self.post_to_chain(ScFuncContext {}.chain_id())
    }

    pub(crate) fn post_request(&self, chain_id: ScChainID) -> wasmrequests::PostRequest {
        wasmrequests::PostRequest {
            chain_id: chain_id,
            contract: self.view.h_contract,
            function: self.view.h_function,
            params: self.view.params.to_bytes(),
            transfer: self.transfer.to_bytes(),
            delay: self.delay,
        }
    }

    pub fn post_to_chain(&self, chain_id: ScChainID) {
        let req = self.post_request(chain_id);
        let res = sandbox(FN_POST, &req.to_bytes());
        self.view.results.copy(&res);
    }

    pub fn transfer(&self, transfer: ScTransfers) -> ScFunc {
        let mut ret = self.clone();
        ret.transfer = transfer.as_assets();
        ret
    }

    pub fn transfer_iotas(&self, amount: u64) -> ScFunc {
        self.transfer(ScTransfers::iotas(amount))
    }
}
