// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::sync::{Arc, Mutex};

use wasmlib::*;

use crate::*;
use crate::codec::*;
use crate::keypair::KeyPair;

pub struct WasmClientContext {
    pub(crate) error: Arc<Mutex<Result<()>>>,
    pub(crate) key_pair: Option<KeyPair>,
    pub(crate) req_id: Arc<Mutex<ScRequestID>>,
    pub(crate) sc_name: String,
    pub(crate) sc_hname: ScHname,
    pub(crate) svc_client: Arc<WasmClientService>,
}

impl WasmClientContext {
    pub fn new(svc_client: Arc<WasmClientService>, sc_name: &str) -> WasmClientContext {
        WasmClientContext {
            error: Arc::new(Mutex::new(Ok(()))),
            key_pair: None,
            req_id: Arc::new(Mutex::new(request_id_from_bytes(&[]))),
            sc_name: String::from(sc_name),
            sc_hname: hname_from_bytes(&hname_bytes(&sc_name)),
            svc_client,
        }
    }

    pub fn current_chain_id(&self) -> ScChainID {
        self.svc_client.current_chain_id()
    }

    pub fn current_keypair(&self) -> Option<KeyPair> {
        self.key_pair.clone()
    }

    pub fn current_svc_client(&self) -> Arc<WasmClientService> {
        self.svc_client.clone()
    }

    pub fn register(&self, handler: Box<dyn IEventHandlers>) {
        self.svc_client.subscribe_events(EventProcessor {
            chain_id: self.svc_client.current_chain_id(),
            contract_id: self.sc_hname,
            handler,
        });
    }

    // overrides default contract name
    pub fn service_contract_name(&mut self, contract_name: &str) {
        self.sc_hname = ScHname::new(contract_name);
    }

    pub fn sign_requests(&mut self, key_pair: &KeyPair) {
        self.key_pair = Some(key_pair.clone());
    }

    pub fn unregister(&self, events_id: u32) {
        self.svc_client.unsubscribe_events(events_id);
    }

    pub fn wait_request(&self) {
        let req_id = self.req_id.lock().unwrap();
        self.wait_request_id(&*req_id);
    }

    pub fn wait_request_id(&self, req_id: &ScRequestID) {
        let res = self.svc_client.wait_until_request_processed(
            req_id,
            std::time::Duration::new(60, 0),
        );

        if let Err(e) = res {
            self.set_err(&e, "")
        }
    }

    pub fn set_err(&self, e1: &str, e2: &str) {
        let mut err = self.error.lock().unwrap();
        *err = Err(e1.to_string() + e2);
    }

    pub fn err(&self) -> Result<()> {
        let err = self.error.lock().unwrap();
        err.clone()
    }
}
