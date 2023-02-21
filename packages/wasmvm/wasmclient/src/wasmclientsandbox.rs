// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

impl ScViewCallContext for WasmClientContext {
    fn fn_call(&self, req: &wasmrequests::CallRequest) -> Vec<u8> {
        if req.contract != self.sc_hname {
            self.set_err("unknown contract: ", &req.contract.to_string());
            return Vec::new();
        }

        let res = self.svc_client.call_view_by_hname(
            &req.contract,
            &req.function,
            &req.params,
        );

        if let Err(e) = &res {
            self.set_err(&e, "");
            return Vec::new();
        }

        res.unwrap()
    }

    fn fn_chain_id(&self) -> ScChainID {
        self.current_chain_id()
    }

    fn init_view_call_context(&self, _contract_hname: ScHname) -> ScHname {
        self.sc_hname
    }
}

impl ScFuncCallContext for WasmClientContext {
    fn fn_post(&self, req: &wasmrequests::PostRequest) -> Vec<u8> {
        if self.key_pair.is_none() {
            self.set_err("fn_post: ", "missing key pair");
            return Vec::new();
        }

        if req.chain_id != self.current_chain_id() {
            self.set_err("unknown chain id: ", &req.chain_id.to_string());
            return Vec::new();
        }

        if req.contract != self.sc_hname {
            self.set_err("unknown contract: ", &req.contract.to_string());
            return Vec::new();
        }

        let sc_assets = ScAssets::new(&req.transfer);
        let mut nonce = self.nonce.lock().unwrap();
        *nonce += 1;
        let res = self.svc_client.post_request(
            &req.chain_id,
            &req.contract,
            &req.function,
            &req.params,
            &sc_assets,
            self.key_pair.as_ref().unwrap(),
            *nonce,
        );

        match res {
            Ok(req_id) => {
                let mut ctx_req_id = self.req_id.lock().unwrap();
                *ctx_req_id = req_id;
            }
            Err(e) => self.set_err(&e, ""),
        }
        Vec::new()
    }

    fn init_func_call_context(&self) {}
}
