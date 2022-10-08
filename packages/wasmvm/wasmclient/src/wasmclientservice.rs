// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use crypto::ciphers::traits::consts::NonZero;
use hyper::{
    client::HttpConnector,
    {Body, Client},
};
use keypair::*;
use std::{borrow::Borrow, time::Duration};
use wasmlib::*;
use wasp_client::*;

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
    ) -> Result<(), String>;
    fn subscribe_events(&self, msg: [&str]) -> Result<(), String>;
    fn wait_until_request_processed(
        &self,
        chain_id: ScChainID,
        req_id: ScRequestID,
        timeout: Duration,
    ) -> Result<(), String>;
}

pub struct WasmClientService {
    // cvt        : wasmhost.WasmConvertor,
    client: wasp_client::WaspClient,
    event_port: String,
    nonce: u64,
}

impl WasmClientService {
    pub fn new(wasp_api: &str, event_port: &str) -> *mut WasmClientService {
        return &mut WasmClientService {
            client: WaspClient::new(wasp_api, None),
            event_port: event_port.to_string(),
            nonce: 0,
        };
    }

    pub fn default_wasm_client_service() -> *mut WasmClientService {
        return &mut WasmClientService {
            client: WaspClient::new("127.0.0.1:9090", None),
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

        let dict_res =
            self.client
                .CallViewByHname(&chain_id, contract_hname, function_hname, params, None)?;

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
    ) -> Result<(), String> {
        // let iscpChainID = self.cvt.IscpChainID(&chainID);
        // let iscpContract = self.cvt.IscpHname(contract_hname);
        // let iscpFunction = self.cvt.IscpHname(function_hname);
        let params = ScDict::from_bytes(args)?;
        self.nonce = self.nonce + 1;
        // let req = iscp.NewOffLedgerRequest(iscpChainID, iscpContract, iscpFunction, params, self.nonce);
        let req: wasp_client::OffLedgerRequestData = wasp_client::OffLedgerRequest::new(
            chain_id,
            contract_hname,
            function_hname,
            params,
            None,
            self.nonce,
        );
        req.with_allowance(&allowance);
        req.sign(key_pair);
        return self.client.PostOffLedgerRequest(&chain_id, req);
    }

    // FIXME the following implementation is a blocked version. It should be multithread
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
            .WaitUntilRequestProcessed(&chain_id, req_id, timeout)?;

        return Ok(());
    }
}
