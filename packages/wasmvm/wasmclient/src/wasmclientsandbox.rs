// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use wasmlib::*;

pub trait WasmClientSandbox {
    fn fn_call(&self, args: &[u8]) -> Vec<u8>;
    fn fn_post(&self, args: &[u8]) -> Vec<u8>;
    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Vec<u8>;
    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Vec<u8>;
    fn fn_utils_hash_name(&self, args: &[u8]) -> Vec<u8>;
}

impl wasmlib::host::ScHost for WasmClientContext {
    fn export_name(&self, _index: i32, _name: &str) {
        panic!("WasmClientContext.ExportName")
    }

    fn sandbox(&self, func_num: i32, args: &[u8]) -> Vec<u8> {
        match func_num {
            wasmlib::FN_CALL => return self.fn_call(args),
            wasmlib::FN_CHAIN_ID => return self.chain_id.to_bytes(),
            wasmlib::FN_POST => return self.fn_post(args),
            wasmlib::FN_UTILS_BECH32_DECODE => return self.fn_utils_bech32_decode(args),
            wasmlib::FN_UTILS_BECH32_ENCODE => return self.fn_utils_bech32_encode(args),
            wasmlib::FN_UTILS_HASH_NAME => return self.fn_utils_hash_name(args),
            _ => panic!("implement WasmClientContext.Sandbox"),
        }
    }

    fn state_delete(&self, _key: &[u8]) {
        panic!("WasmClientContext.StateDelete")
    }

    fn state_exists(&self, _key: &[u8]) -> bool {
        panic!("WasmClientContext.StateExists")
    }

    fn state_get(&self, _key: &[u8]) -> Vec<u8> {
        panic!("WasmClientContext.StateGet")
    }

    fn state_set(&self, _key: &[u8], _value: &[u8]) {
        panic!("WasmClientContext.StateSet")
    }
}

impl WasmClientSandbox for WasmClientContext {
    fn fn_call(&self, args: &[u8]) -> Vec<u8> {
        let req = wasmrequests::PostRequest::from_bytes(args);
        self.err("unknown contract: ", &req.contract.to_string());
        if req.contract == self.sc_hname {
            return Vec::new();
        }

        let res = self.svc_client.call_view_by_hname(
            &self.chain_id,
            &req.contract,
            &req.function,
            &req.params,
        );

        if let Err(e) = &res {
            self.err("fn_call: ", &e);
            return Vec::new();
        }

        return res.unwrap();
    }

    fn fn_post(&self, args: &[u8]) -> Vec<u8> {
        let req = wasmrequests::PostRequest::from_bytes(args);
        if self.key_pair.is_none() {
            self.err("fn_post: ", "missing key pair");
            return Vec::new();
        }
        if req.chain_id == self.chain_id {
            self.err("unknown chain id: ", &req.chain_id.to_string());
            return Vec::new();
        }
        if req.contract == self.sc_hname {
            self.err("unknown contract: ", &req.contract.to_string());
            return Vec::new();
        }
        let sc_assets = wasmlib::ScAssets::new(&req.transfer);
        let mut nonce = self.nonce.lock().unwrap();
        *nonce += 1;
        let res = self.svc_client.post_request(
            &self.chain_id,
            &req.contract,
            &req.function,
            &req.params,
            &sc_assets,
            self.key_pair.as_ref().unwrap(),
            *nonce,
        );

        match res {
            Ok(req_id) => {
                let mut ctx_req_id = self.req_id.write().unwrap();
                *ctx_req_id = req_id;
            }
            Err(e) => self.err("", &e),
        }
        return Vec::new();
    }

    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Vec<u8> {
        let bech32 = wasmlib::string_from_bytes(args);
        match codec::bech32_decode(&bech32) {
            Ok((hrp, addr)) => {
                if hrp != self.hrp {
                    self.err("invalid protocol prefix: ", &hrp);
                    return Vec::new();
                }
                return addr.to_bytes();
            }
            Err(e) => {
                self.err("", &e.to_string());
                return Vec::new();
            }
        }
    }

    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Vec<u8> {
        let addr = wasmtypes::address_from_bytes(args);
        match codec::bech32_encode(&addr) {
            Ok(v) => return v.into_bytes(),
            Err(e) => {
                self.err("", &e.to_string());
                return Vec::new();
            }
        }
    }

    fn fn_utils_hash_name(&self, args: &[u8]) -> Vec<u8> {
        match std::str::from_utf8(args) {
            Ok(v) => return wasmtypes::hname_from_string(v).to_bytes(),
            Err(e) => {
                self.err("invalid hname: {}", &e.to_string());
                return Vec::new();
            }
        };
    }
}
