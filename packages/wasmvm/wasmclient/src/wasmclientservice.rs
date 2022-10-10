// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use isc::{offledger::*, waspclient};
use keypair::*;
use std::time::Duration;
use wasmlib::*;

pub trait IClientService {
    fn call_view_by_hname(
        &self,
        chain_id: ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: &[u8],
    ) -> Result<Vec<u8>, String>;
    fn post_request(
        &self,
        chain_id: ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: &[u8],
        allowance: ScAssets,
        key_pair: KeyPair,
    ) -> Result<ScRequestID, String>;
    fn subscribe_events(&self, msg: [&str]) -> Result<(), String>;
    fn wait_until_request_processed(
        &self,
        chain_id: ScChainID,
        req_id: ScRequestID,
        timeout: Duration,
    ) -> Result<(), String>;
}

pub struct WasmClientService {
    client: waspclient::WaspClient,
    event_port: String,
    nonce: u64,
}

impl WasmClientService {
    pub fn new(wasp_api: &str, event_port: &str) -> *mut WasmClientService {
        return &mut WasmClientService {
            client: waspclient::WaspClient::new(wasp_api, None),
            event_port: event_port.to_string(),
            nonce: 0,
        };
    }

    pub fn default_wasm_client_service() -> *mut WasmClientService {
        return &mut WasmClientService {
            client: waspclient::WaspClient::new("127.0.0.1:9090", None),
            event_port: "127.0.0.1:5550".to_string(),
            nonce: 0,
        };
    }

    pub fn call_view_by_hname(
        &self,
        chain_id: ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: &[u8],
    ) -> Result<Vec<u8>, String> {
        let params = ScDict::from_bytes(args)?;

        let dict_res = self.client.call_view_by_hname(
            &chain_id,
            contract_hname,
            function_hname,
            params,
            None,
        )?;

        return Ok(dict_res.to_bytes());
    }

    pub fn post_request(
        &mut self,
        chain_id: ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: &[u8],
        allowance: ScAssets,
        key_pair: KeyPair,
    ) -> Result<ScRequestID, String> {
        let params = ScDict::from_bytes(args)?;
        self.nonce += 1;
        let req: offledger::OffLedgerRequestData = offledger::OffLedgerRequest::new(
            chain_id,
            contract_hname,
            function_hname,
            params,
            None,
            self.nonce,
        );
        req.with_allowance(&allowance);
        req.sign(key_pair);

        self.client.post_offledger_request(&chain_id, &req)?;
        return Ok(req.id());
    }

    // FIXME the following implementation is a blocked version. It should be multithread
    // To impl channels, see https://doc.rust-lang.org/rust-by-example/std_misc/channels.html
    pub fn subscribe_events(&self, msg: &Vec<String>) -> Result<(), String> {
        return Err("not impl".to_string());
    }

    pub fn wait_until_request_processed(
        &self,
        chain_id: ScChainID,
        req_id: ScRequestID,
        timeout: Duration,
    ) -> Result<(), String> {
        let _ = self
            .client
            .wait_until_request_processed(&chain_id, req_id, timeout)?;

        return Ok(());
    }
}
