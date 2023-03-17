// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use std::cell::Cell;
use std::collections::HashMap;
use std::sync::{Arc, mpsc, Mutex};
use std::sync::mpsc::channel;
use std::thread::JoinHandle;
use std::time::Duration;

use crypto::signatures::ed25519::PublicKey;
use reqwest::{blocking, StatusCode};
use serde::{Deserialize, Serialize};
use wasmlib::*;

use crate::*;
use crate::codec::*;
use crate::keypair::KeyPair;

const READ_TIMEOUT: Duration = Duration::from_millis(10000);

#[derive(Serialize, Deserialize, Debug, Clone)]
struct ChainInfoResponse {
    #[serde(rename = "chainID")]
    pub(crate) chain_id: String,
}

pub struct WasmClientService {
    chain_id: ScChainID,
    close_rx: Arc<Mutex<mpsc::Receiver<bool>>>,
    close_tx: mpsc::Sender<bool>,
    event_handlers: Arc<Mutex<Vec<WasmClientEvents>>>,
    handle: Cell<Option<JoinHandle<()>>>,
    nonces: Arc<Mutex<HashMap<PublicKey, u64>>>,
    wasp_api: String,
}

impl WasmClientService {
    pub fn new(wasp_api: &str) -> Self {
        let (tx, rx) = channel();
        WasmClientService {
            chain_id: chain_id_from_bytes(&[]),
            close_rx: Arc::new(Mutex::new(rx)),
            close_tx: tx,
            event_handlers: Arc::default(),
            handle: Cell::new(None),
            nonces: Arc::default(),
            wasp_api: String::from(wasp_api),
        }
    }

    pub(crate) fn call_view_by_hname(
        &self,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
    ) -> Result<Vec<u8>> {
        let url = self.wasp_api.clone() + "/requests/callview";
        let client = blocking::Client::builder()
            .timeout(READ_TIMEOUT)
            .build()
            .unwrap();
        let body = APICallViewRequest {
            arguments: json_encode(args),
            chain_id: self.chain_id.to_string(),
            contract_hname: contract_hname.to_string(),
            function_hname: function_hname.to_string(),
        };
        match client.post(url).json(&body).send() {
            Ok(v) => match v.status() {
                StatusCode::OK => {
                    match v.json::<JsonResponse>() {
                        Ok(json_obj) => Ok(json_decode(json_obj)),
                        Err(e) => Err(format!("call() response failed: {}", e.to_string())),
                    }
                }
                status => {
                    match v.json::<JsonError>() {
                        Ok(err_msg) => Err(Self::api_error(status, err_msg)),
                        Err(e) => Err(e.to_string()),
                    }
                }
            },
            Err(e) => Err(format!("call() request failed: {}", e.to_string())),
        }
    }

    pub fn current_chain_id(&self) -> ScChainID {
        self.chain_id
    }

    pub fn is_healthy(&self) -> bool {
        let url = format!("{}/health", self.wasp_api);
        let client = blocking::Client::builder().build().unwrap();
        let res = client.get(url).header("Content-Type", "application/json").send();
        !res.is_err() && res.unwrap().status() == StatusCode::OK
    }

    pub(crate) fn post_request(
        &self,
        chain_id: &ScChainID,
        h_contract: &ScHname,
        h_function: &ScHname,
        args: &[u8],
        allowance: &ScAssets,
        key_pair: &KeyPair,
    ) -> Result<ScRequestID> {
        let nonce: u64;
        match self.cache_nonce(key_pair) {
            Ok(n) => nonce = n,
            Err(e) => return Err(e),
        }
        let mut req =
            offledgerrequest::OffLedgerRequest::new(
                chain_id,
                h_contract,
                h_function,
                args,
                nonce,
            );
        req.with_allowance(&allowance);
        let signed = req.sign(key_pair);

        let url = self.wasp_api.clone() + "/requests/offledger";
        let client = blocking::Client::new();
        let body = APIOffLedgerRequest {
            chain_id: chain_id.to_string(),
            request: hex_encode(&signed.to_bytes()),
        };
        match client.post(url).json(&body).send() {
            Ok(v) => match v.status() {
                StatusCode::OK => Ok(signed.id()),
                StatusCode::ACCEPTED => Ok(signed.id()),
                status => {
                    match v.json::<JsonError>() {
                        Ok(err_msg) => Err(Self::api_error(status, err_msg)),
                        Err(e) => Err(e.to_string()),
                    }
                }
            },
            Err(e) => Err(format!("post() request failed: {}", e.to_string())),
        }
    }

    pub fn set_current_chain_id(&mut self, chain_id: &str) -> Result<()> {
        set_sandbox_wrappers(chain_id).unwrap();
        self.chain_id = chain_id_from_string(chain_id);
        Ok(())
    }

    pub fn set_default_chain_id(&mut self) -> Result<()> {
        let url = format!("{}/chains", self.wasp_api);
        let client = blocking::Client::builder()
            .build()
            .unwrap();
        let res = client.get(url).header("Content-Type", "application/json").send();
        match res {
            Ok(v) => match v.status() {
                StatusCode::OK =>
                    match v.json::<Vec<ChainInfoResponse>>() {
                        Ok(chains) => {
                            if chains.len() != 1 {
                                return Err(String::from("expected a single chain for default chain ID"));
                            }
                            let chain_id = &chains[0].chain_id;
                            println!("default chain ID: {}", chain_id);
                            self.set_current_chain_id(chain_id)
                        }
                        Err(e) => Err(format!("response failed: {}", e.to_string())),
                    },
                status => {
                    match v.text() {
                        Ok(err_msg) => Err(format!("{}: {}", status, err_msg)),
                        Err(e) => Err(e.to_string()),
                    }
                }
            },
            Err(e) => Err(format!("request failed: {}", e.to_string())),
        }
    }

    pub(crate) fn subscribe_events(&self, event_handler: WasmClientEvents) {
        {
            let mut event_handlers = self.event_handlers.lock().unwrap();
            event_handlers.push(event_handler);
            if event_handlers.len() != 1 {
                return;
            }
        }

        let socket_url = self.wasp_api.replace("http:", "ws:") + "/ws";
        let close_rx = self.close_rx.clone();
        let event_handlers = self.event_handlers.clone();
        let handle = WasmClientEvents::start_event_loop(socket_url, close_rx, event_handlers);
        self.handle.set(Some(handle));
    }

    pub(crate) fn unsubscribe_events(&self, events_id: u32) {
        let mut event_handlers = self.event_handlers.lock().unwrap();
        event_handlers.retain(|h| {
            h.handler.id() != events_id
        });
        if event_handlers.len() == 0 {
            // stop event loop
            if let Some(handle) = self.handle.take() {
                self.close_tx.send(true).unwrap();
                handle.join().unwrap();
                self.handle.set(None);
            }
        }
    }

    pub(crate) fn wait_until_request_processed(
        &self,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> Result<()> {
        let url = format!(
            "{}/chains/{}/requests/{}/wait",
            self.wasp_api,
            self.chain_id.to_string(),
            req_id.to_string()
        );
        let client = blocking::Client::builder()
            .timeout(timeout)
            .build()
            .unwrap();
        let res = client.get(url).header("Content-Type", "application/json").send();
        match res {
            Ok(v) => match v.status() {
                StatusCode::OK => Ok(()),
                status => {
                    match v.text() {
                        Ok(err_msg) => Err(format!("{}: {}", status, err_msg)),
                        Err(e) => Err(e.to_string()),
                    }
                }
            },
            Err(e) => Err(format!("request failed: {}", e.to_string())),
        }
    }

    fn api_error(status: StatusCode, err_msg: JsonError) -> String {
        format!("{}: {}: {}", status, err_msg.message, err_msg.error)
    }

    fn cache_nonce(&self, key_pair: &KeyPair) -> Result<u64> {
        let key = key_pair.public_key;
        let mut nonces = self.nonces.lock().unwrap();
        let mut nonce: u64;
        match nonces.get(&key) {
            None => {
                // get last used nonce from accounts core contract
                let isc_agent = ScAgentID::from_address(&key_pair.address());
                let mut svc = WasmClientService::new(&self.wasp_api);
                svc.set_current_chain_id(&self.chain_id.to_string()).unwrap();
                let ctx = WasmClientContext::new(Arc::new(svc), coreaccounts::SC_NAME);
                let n = coreaccounts::ScFuncs::get_account_nonce(&ctx);
                n.params.agent_id().set_value(&isc_agent);
                n.func.call();
                match ctx.err() {
                    Ok(_) => nonce = n.results.account_nonce().value(),
                    Err(e) => return Err(e),
                }
            }
            Some(n) => nonce = *n,
        }
        nonce += 1;
        nonces.insert(key, nonce);
        Ok(nonce)
    }
}
