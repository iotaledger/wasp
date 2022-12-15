// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use crate::*;
use isc::{offledgerrequest::*, waspclient::*};
use std::sync::{mpsc, Arc, RwLock};
use std::time::Duration;

pub trait IClientService {
    fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
    ) -> errors::Result<Vec<u8>>;
    fn post_request(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
        allowance: &ScAssets,
        key_pair: &keypair::KeyPair,
        nonce: u64,
    ) -> errors::Result<ScRequestID>;
    fn subscribe_events(
        &self,
        tx: mpsc::Sender<String>,
        done: Arc<RwLock<bool>>,
    ) -> errors::Result<()>;
    fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> errors::Result<()>;
}

#[derive(Clone, PartialEq)]
pub struct WasmClientService {
    client: waspclient::WaspClient,
    websocket: Option<websocket::Client>,
    event_port: String,
    last_err: errors::Result<()>,
}

impl IClientService for WasmClientService {
    fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
    ) -> errors::Result<Vec<u8>> {
        let params = ScDict::from_bytes(args)?;

        let _ = self.client.call_view_by_hname(
            chain_id,
            contract_hname,
            function_hname,
            &params,
            None,
        )?;

        return Ok(Vec::new());
    }

    fn post_request(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
        allowance: &ScAssets,
        key_pair: &keypair::KeyPair,
        nonce: u64,
    ) -> errors::Result<ScRequestID> {
        let params = ScDict::from_bytes(args)?;
        let mut req: offledgerrequest::OffLedgerRequestData =
            offledgerrequest::OffLedgerRequest::new(
                chain_id,
                contract_hname,
                function_hname,
                &params,
                nonce,
            );
        req.with_allowance(&allowance);
        req.sign(key_pair);
        self.client.post_offledger_request(&chain_id, &req)?;
        return Ok(req.id());
    }

    fn subscribe_events(
        &self,
        tx: mpsc::Sender<String>,
        done: Arc<RwLock<bool>>,
    ) -> errors::Result<()> {
        self.websocket.clone().unwrap().subscribe(tx, done); // TODO remove clone
        return Ok(());
    }

    fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> errors::Result<()> {
        let _ = self
            .client
            .wait_until_request_processed(&chain_id, req_id, timeout)?;

        return Ok(());
    }
}

impl WasmClientService {
    pub fn new(wasp_api: &str, event_port: &str, websocket_url: &str) -> Self {
        return WasmClientService {
            client: waspclient::WaspClient::new(wasp_api),
            websocket: Some(websocket::Client::new(websocket_url).unwrap()),
            event_port: event_port.to_string(),
            last_err: Ok(()),
        };
    }
}

impl Default for WasmClientService {
    fn default() -> Self {
        return WasmClientService {
            client: waspclient::WaspClient::new("127.0.0.1:19090"),
            event_port: "127.0.0.1:15550".to_string(),
            websocket: None, // TODO set an empty object
            last_err: Ok(()),
        };
    }
}

#[cfg(test)]
mod tests {
    use crate::isc::waspclient;
    use crate::WasmClientService;

    #[test]
    fn service_default() {
        let service = WasmClientService::default();
        let default_service = WasmClientService {
            client: waspclient::WaspClient::new("127.0.0.1:19090"),
            websocket: None,
            event_port: "127.0.0.1:15550".to_string(),
            last_err: Ok(()),
        };
        assert!(default_service.event_port == service.event_port);
        assert!(default_service.last_err == Ok(()));
    }
}
