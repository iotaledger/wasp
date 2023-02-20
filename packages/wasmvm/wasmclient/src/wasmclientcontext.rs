// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::sync::{Arc, Mutex};

use wasmlib::*;

use crate::*;
use crate::codec::*;
use crate::keypair::KeyPair;

pub struct WasmClientContext {
    pub(crate) chain_id: ScChainID,
    pub(crate) error: Arc<Mutex<Result<()>>>,
    pub(crate) event_done: Arc<Mutex<bool>>,
    pub(crate) event_handlers: Arc<Mutex<Vec<Box<dyn IEventHandlers>>>>,
    pub(crate) key_pair: Option<KeyPair>,
    pub(crate) nonce: Arc<Mutex<u64>>,
    pub(crate) req_id: Arc<Mutex<ScRequestID>>,
    pub(crate) sc_name: String,
    pub(crate) sc_hname: ScHname,
    pub(crate) svc_client: WasmClientService,
}

impl WasmClientContext {
    pub fn new(svc_client: &WasmClientService, chain_id: &str, sc_name: &str) -> WasmClientContext {
        let mut ctx = WasmClientContext {
            chain_id: chain_id_from_bytes(&[]),
            error: Arc::new(Mutex::new(Ok(()))),
            event_done: Arc::default(),
            event_handlers: Arc::default(),
            key_pair: None,
            nonce: Arc::default(),
            req_id: Arc::new(Mutex::new(request_id_from_bytes(&[]))),
            sc_name: String::from(sc_name),
            sc_hname: hname_from_bytes(&hname_bytes(&sc_name)),
            svc_client: svc_client.clone(),
        };

        if let Err(e) = set_sandbox_wrappers(chain_id) {
            ctx.set_err(&e, "");
            return ctx;
        }

        ctx.chain_id = chain_id_from_string(chain_id);
        ctx
    }

    pub fn current_chain_id(&self) -> ScChainID {
        return self.chain_id;
    }

    pub fn current_keypair(&self) -> Option<KeyPair> {
        return self.key_pair.clone();
    }

    pub fn current_svc_client(&self) -> WasmClientService {
        return self.svc_client.clone();
    }

    pub fn register(&mut self, handler: Box<dyn IEventHandlers>) {
        {
            let target = handler.id();
            let mut event_handlers = self.event_handlers.lock().unwrap();
            for h in event_handlers.iter() {
                if h.id() == target {
                    return;
                }
            }
            event_handlers.push(handler);
            if event_handlers.len() > 1 {
                return;
            }
        }
        let res = self.svc_client.subscribe_events(self.event_handlers.clone(), self.event_done.clone());
        if let Err(e) = res {
            self.set_err(&e, "")
        }
    }

    // overrides default contract name
    pub fn service_contract_name(&mut self, contract_name: &str) {
        self.sc_hname = ScHname::new(contract_name);
    }

    pub fn sign_requests(&mut self, key_pair: &KeyPair) {
        self.key_pair = Some(key_pair.clone());
        // get last used nonce from accounts core contract
        let isc_agent = ScAgentID::from_address(&key_pair.address());
        let ctx = WasmClientContext::new(
            &self.svc_client,
            &self.chain_id.to_string(),
            coreaccounts::SC_NAME,
        );
        let n = coreaccounts::ScFuncs::get_account_nonce(&ctx);
        n.params.agent_id().set_value(&isc_agent);
        n.func.call();
        let mut nonce = self.nonce.lock().unwrap();
        *nonce = n.results.account_nonce().value();
    }

    pub fn unregister(&mut self, id: &str) {
        let mut event_handlers = self.event_handlers.lock().unwrap();
        event_handlers.retain(|h| {
            h.id() != id
        });
        if event_handlers.len() == 0 {
            self.stop_event_handlers();
        }
    }

    pub fn wait_request(&self) {
        let req_id = self.req_id.lock().unwrap();
        self.wait_request_id(&*req_id);
    }

    pub fn wait_request_id(&self, req_id: &ScRequestID) {
        let res = self.svc_client.wait_until_request_processed(
            &self.chain_id,
            req_id,
            std::time::Duration::new(60, 0),
        );

        if let Err(e) = res {
            self.set_err(&e, "")
        }
    }

    pub(crate) fn process_event(event_handlers: &Arc<Mutex<Vec<Box<dyn IEventHandlers>>>>, event: &ContractEvent) {
        let mut params: Vec<String> = event.data.split("|").map(|s| s.into()).collect();
        for i in 0..params.len() {
            params[i] = Self::unescape(&params[i]);
        }
        let topic = params.remove(0);

        let event_handlers = event_handlers.lock().unwrap();
        for handler in event_handlers.iter() {
            handler.as_ref().call_handler(&topic, &params);
        }
    }

    pub fn stop_event_handlers(&self) {
        let mut done = self.event_done.lock().unwrap();
        *done = true;
    }

    fn unescape(param: &str) -> String {
        let idx = match param.find("~") {
            Some(idx) => idx,
            None => return String::from(param),
        };
        match param.chars().nth(idx + 1) {
            // escaped escape character
            Some('~') => param[0..idx].to_string() + "~" + &Self::unescape(&param[idx + 2..]),
            // escaped vertical bar
            Some('/') => param[0..idx].to_string() + "|" + &Self::unescape(&param[idx + 2..]),
            // escaped space
            Some('_') => param[0..idx].to_string() + " " + &Self::unescape(&param[idx + 2..]),
            _ => panic!("invalid event encoding"),
        }
    }

    pub fn set_err(&self, e1: &str, e2: &str) {
        let mut err = self.error.lock().unwrap();
        *err = Err(e1.to_string() + e2);
    }

    pub fn err(&self) -> Result<()> {
        let err = self.error.lock().unwrap();
        return err.clone();
    }
}
