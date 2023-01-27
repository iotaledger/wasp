// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// base contract objects

use std::rc::Rc;

use crate::*;
use crate::host::*;
use crate::wasmrequests::*;

pub trait ScViewCallContext {
    fn fn_call(&self, req: &CallRequest) -> Vec<u8> {
        sandbox(FN_CALL, &req.to_bytes())
    }

    fn fn_chain_id(&self) -> ScChainID {
        chain_id_from_bytes(&sandbox(FN_CHAIN_ID, &[]))
    }

    fn init_view_call_context(&self, h_contract: ScHname) -> ScHname;
}

pub trait ScFuncCallContext: ScViewCallContext {
    fn fn_post(&self, req: &PostRequest) -> Vec<u8> {
        sandbox(FN_POST, &req.to_bytes())
    }

    fn init_func_call_context(&self);
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

#[derive(Clone)]
pub struct ScView<'a> {
    ctx: &'a dyn ScViewCallContext,
    h_contract: ScHname,
    h_function: ScHname,
    params: Rc<ScDict>,
    results: Rc<ScDict>,
}

impl<'a> ScView<'_> {
    pub fn new(ctx: &'a impl ScViewCallContext, h_contract: ScHname, h_function: ScHname) -> ScView {
        ScView {
            // allow context to override default hContract
            ctx: ctx,
            h_contract: ctx.init_view_call_context(h_contract),
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
        self.call_with_allowance(None);
    }

    fn call_with_allowance(&self, allowance: Option<ScTransfer>) {
        let mut req = wasmrequests::CallRequest {
            contract: self.h_contract,
            function: self.h_function,
            params: self.params.to_bytes(),
            allowance: vec![0; SC_UINT32_LENGTH],
        };
        if let Some(allowance) = allowance {
            req.allowance = allowance.to_bytes();
        }
        let res = self.ctx.fn_call(&req);
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
pub struct ScInitFunc<'a> {
    view: ScView<'a>,
}

impl<'a> ScInitFunc<'_> {
    pub fn new(ctx: &'a impl ScFuncCallContext, h_contract: ScHname, h_function: ScHname) -> ScInitFunc {
        ScInitFunc {
            view: ScView::new(ctx, h_contract, h_function),
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
pub struct ScFunc<'a> {
    pub view: ScView<'a>,
    allowance: ScTransfer,
    delay: u32,
    fctx: &'a dyn ScFuncCallContext,
    transfer: ScTransfer,
}

impl<'a> ScFunc<'_> {
    pub fn new(ctx: &'a impl ScFuncCallContext, h_contract: ScHname, h_function: ScHname) -> ScFunc {
        ScFunc {
            view: ScView::new(ctx, h_contract, h_function),
            allowance: ScTransfer::new(),
            delay: 0,
            fctx: ctx,
            transfer: ScTransfer::new(),
        }
    }

    pub fn link_params(proxy: &mut Proxy, func: &ScFunc) {
        ScView::link_params(proxy, &func.view);
    }

    pub fn link_results(proxy: &mut Proxy, func: &ScFunc) {
        ScView::link_results(proxy, &func.view);
    }

    pub fn allowance(&self, allowance: ScTransfer) -> ScFunc {
        let mut ret = self.clone();
        ret.allowance = allowance.clone();
        ret
    }

    pub fn allowance_base_tokens(&self, amount: u64) -> ScFunc {
        self.allowance(ScTransfer::base_tokens(amount))
    }

    pub fn call(&self) {
        if !self.transfer.is_empty() {
            panic("cannot transfer assets in a call");
        }
        if self.delay != 0 {
            panic("cannot delay a call");
        }
        self.view.call_with_allowance(Some(self.transfer.clone()));
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
        self.post_to_chain(self.fctx.fn_chain_id())
    }

    pub(crate) fn post_request(&self, chain_id: ScChainID) -> wasmrequests::PostRequest {
        wasmrequests::PostRequest {
            chain_id: chain_id,
            contract: self.view.h_contract,
            function: self.view.h_function,
            params: self.view.params.to_bytes(),
            allowance: self.allowance.to_bytes(),
            transfer: self.transfer.to_bytes(),
            delay: self.delay,
        }
    }

    pub fn post_to_chain(&self, chain_id: ScChainID) {
        let req = self.post_request(chain_id);
        let res = self.fctx.fn_post(&req);
        self.view.results.copy(&res);
    }

    pub fn transfer(&self, transfer: ScTransfer) -> ScFunc {
        let mut ret = self.clone();
        ret.transfer = transfer.clone();
        ret
    }

    pub fn transfer_base_tokens(&self, amount: u64) -> ScFunc {
        self.transfer(ScTransfer::base_tokens(amount))
    }
}
