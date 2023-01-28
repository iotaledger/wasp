// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use wasmlib::*;

pub trait WasmClientSandbox {
    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Vec<u8>;
    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Vec<u8>;
    fn fn_utils_hash_name(&self, args: &[u8]) -> Vec<u8>;
}

impl host::ScHost for WasmClientContext {
    fn export_name(&self, _index: i32, _name: &str) {
        panic!("WasmClientContext.ExportName")
    }

    fn sandbox(&self, func_num: i32, args: &[u8]) -> Vec<u8> {
        match func_num {
            FN_CALL => return self.fn_call(&wasmrequests::CallRequest::from_bytes(args)),
            FN_CHAIN_ID => return self.fn_chain_id().to_bytes(),
            FN_POST => return self.fn_post(&wasmrequests::PostRequest::from_bytes(args)),
            FN_UTILS_BECH32_DECODE => return self.fn_utils_bech32_decode(args),
            FN_UTILS_BECH32_ENCODE => return self.fn_utils_bech32_encode(args),
            FN_UTILS_HASH_NAME => return self.fn_utils_hash_name(args),
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

impl ScViewCallContext for WasmClientContext {
    fn fn_call(&self, req: &wasmrequests::CallRequest) -> Vec<u8> {
        let lock_received = self.event_received.clone();
        let mut received = lock_received.write().unwrap();
        *received = false;

        if req.contract != self.sc_hname {
            self.err("unknown contract: ", &req.contract.to_string());
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

        res.unwrap()
    }

    fn fn_chain_id(&self) -> ScChainID {
        self.chain_id
    }

    fn init_view_call_context(&self, _contract_hname: ScHname) -> ScHname {
        self.sc_hname
    }
}

impl ScFuncCallContext for WasmClientContext {
    fn fn_post(&self, req: &wasmrequests::PostRequest) -> Vec<u8> {
        let lock_received = self.event_received.clone();
        let mut received = lock_received.write().unwrap();
        *received = false;

        if self.key_pair.is_none() {
            self.err("fn_post: ", "missing key pair");
            return Vec::new();
        }
        if req.chain_id != self.chain_id {
            self.err("unknown chain id: ", &req.chain_id.to_string());
            return Vec::new();
        }
        if req.contract != self.sc_hname {
            self.err("unknown contract: ", &req.contract.to_string());
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
                let mut ctx_req_id = self.req_id.write().unwrap();
                *ctx_req_id = req_id;
            }
            Err(e) => self.err("", &e),
        }
        Vec::new()
    }

    fn init_func_call_context(&self) {}
}

impl WasmClientSandbox for WasmClientContext {
    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Vec<u8> {
        let bech32 = string_from_bytes(args);
        client_bech32_decode(&bech32).to_bytes()
    }

    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Vec<u8> {
        let addr = address_from_bytes(args);
        string_to_bytes(&client_bech32_encode(&addr))
    }

    fn fn_utils_hash_name(&self, args: &[u8]) -> Vec<u8> {
        let name = string_from_bytes(args);
        client_hash_name(&name).to_bytes()
    }
}

pub(crate) static mut HRP_FOR_CLIENT: String = String::new();

pub(crate) fn client_bech32_decode(bech32: &str) -> ScAddress {
    match codec::bech32_decode(&bech32) {
        Ok((hrp, addr)) => unsafe {
            if hrp != HRP_FOR_CLIENT {
                panic(&("invalid protocol prefix: ".to_owned() + &hrp));
                return address_from_bytes(&[]);
            }
            return addr;
        }
        Err(e) => {
            panic(&e.to_string());
            return address_from_bytes(&[]);
        }
    }
}

pub(crate) fn client_bech32_encode(addr: &ScAddress) -> String {
    unsafe {
        match codec::bech32_encode(&HRP_FOR_CLIENT, &addr) {
            Ok(v) => return v,
            Err(e) => {
                panic(&e.to_string());
                return String::new();
            }
        }
    }
}

pub(crate) fn client_hash_name(name: &str) -> ScHname {
    hname_from_bytes(&codec::hname_bytes(name))
}
