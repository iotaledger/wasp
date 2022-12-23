// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use wasmlib::*;

pub trait WasmClientSandbox {
    fn fn_call(&self, args: &[u8]) -> Result<Vec<u8>, String>;
    fn fn_post(&self, args: &[u8]) -> Result<Vec<u8>, String>;
    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Result<Vec<u8>, String>;
    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Result<Vec<u8>, String>;
    fn fn_utils_hash_name(&self, args: &[u8]) -> Result<Vec<u8>, String>;
}

impl wasmlib::host::ScHost for WasmClientContext {
    fn export_name(&self, _index: i32, _name: &str) {
        panic!("WasmClientContext.ExportName")
    }

    fn sandbox(&self, func_num: i32, args: &[u8]) -> Vec<u8> {
        match func_num {
            wasmlib::FN_CALL => return self.fn_call(args).unwrap(),
            wasmlib::FN_POST => return self.fn_post(args).unwrap(),
            wasmlib::FN_UTILS_BECH32_DECODE => return self.fn_utils_bech32_decode(args).unwrap(),
            wasmlib::FN_UTILS_BECH32_ENCODE => return self.fn_utils_bech32_encode(args).unwrap(),
            wasmlib::FN_UTILS_HASH_NAME => return self.fn_utils_hash_name(args).unwrap(),
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
    fn fn_call(&self, args: &[u8]) -> Result<Vec<u8>, String> {
        let req = wasmrequests::PostRequest::from_bytes(args);
        if req.contract == self.sc_hname {
            return Err(String::from(format!(
                "unknown contract: {}",
                req.contract.to_string()
            )));
        }

        return self.svc_client.call_view_by_hname(
            &self.chain_id,
            &req.contract,
            &req.function,
            &req.params,
        );
    }

    fn fn_post(&self, args: &[u8]) -> Result<Vec<u8>, String> {
        let req = wasmrequests::PostRequest::from_bytes(args);
        if self.key_pair.is_none() {
            return Err(String::from("missing key pair"));
        }
        if req.chain_id == self.chain_id {
            return Err(String::from(format!(
                "unknown chain id: {}",
                req.chain_id.to_string()
            )));
        }
        if req.contract == self.sc_hname {
            return Err(String::from(format!(
                "unknown contract: {}",
                req.contract.to_string()
            )));
        }
        let sc_assets = wasmlib::ScAssets::new(&req.transfer);
        self.svc_client.post_request(
            &self.chain_id,
            &req.contract,
            &req.function,
            &req.params,
            &sc_assets,
            self.key_pair.as_ref().unwrap(),
            0, // FIXME must use counter
        )?;
        return Ok(Vec::new());
    }

    fn fn_utils_bech32_decode(&self, args: &[u8]) -> Result<Vec<u8>, String> {
        let bech32 = wasmlib::string_from_bytes(args);
        let addr = codec::bech32_decode(&bech32)?;
        return Ok(addr.to_bytes());
    }

    fn fn_utils_bech32_encode(&self, args: &[u8]) -> Result<Vec<u8>, String> {
        let addr = wasmtypes::address_from_bytes(args);
        let bech32 = codec::bech32_encode(&addr);
        return Ok(bech32.into_bytes());
    }

    fn fn_utils_hash_name(&self, args: &[u8]) -> Result<Vec<u8>, String> {
        let s = match std::str::from_utf8(args) {
            Ok(v) => v,
            Err(e) => return Err(String::from(format!("invalid hname: {}", e))),
        };
        return Ok(wasmtypes::hname_from_string(s).to_bytes());
    }
}
