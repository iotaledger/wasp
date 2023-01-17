// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::*;
use std::{
    any::Any,
    sync::{mpsc, Arc, RwLock},
    thread::spawn,
};
use wasmlib::*;

// TODO to handle the request in parallel, WasmClientContext must be static now.
// We need to solve this problem. By copying the vector of event_handlers, we may solve this problem
pub struct WasmClientContext {
    pub chain_id: ScChainID,
    pub error: Arc<RwLock<errors::Result<()>>>,
    event_done: Arc<RwLock<bool>>, // Set `done` to true to close the ongoing `subscribe()`
    pub event_handlers: Vec<Box<dyn IEventHandlers>>,
    pub event_received: Arc<RwLock<bool>>,
    pub hrp: String,
    pub key_pair: Option<keypair::KeyPair>,
    pub nonce: Arc<RwLock<u64>>,
    pub req_id: Arc<RwLock<ScRequestID>>,
    pub sc_name: String,
    pub sc_hname: ScHname,
    pub svc_client: WasmClientService, //TODO Maybe  use 'dyn IClientService' for 'svc_client' instead of a struct
}

impl WasmClientContext {
    pub fn new(
        svc_client: &WasmClientService,
        chain_id: &wasmlib::ScChainID,
        sc_name: &str,
    ) -> WasmClientContext {
        let (hrp, _data, _v) = match bech32::decode(sc_name) {
            Ok(vals) => vals,
            Err(e) => {
                let ctx = WasmClientContext::default();
                ctx.err("WasmClientContext init err: ", &e.to_string());
                return ctx;
            }
        };

        WasmClientContext {
            chain_id: chain_id.clone(),
            error: Arc::new(RwLock::new(Ok(()))),
            event_done: Arc::new(RwLock::new(false)),
            event_handlers: Vec::new(),
            event_received: Arc::new(RwLock::new(false)),
            hrp: hrp.to_string(),
            key_pair: None,
            nonce: Arc::new(RwLock::new(0)),
            req_id: Arc::new(RwLock::new(request_id_from_bytes(&[]))),
            sc_name: sc_name.to_string(),
            sc_hname: ScHname::new(sc_name),
            svc_client: svc_client.clone(),
        }
    }

    pub fn current_chain_id(&self) -> ScChainID {
        return self.chain_id;
    }

    pub fn init_func_call_context(&'static self) {
        wasmlib::host::connect_host(self);
    }

    pub fn init_view_call_context(&'static self, _contract_hname: &ScHname) -> ScHname {
        wasmlib::host::connect_host(self);
        return self.sc_hname;
    }

    pub fn register(&'static mut self, handler: Box<dyn IEventHandlers>) -> errors::Result<()> {
        for h in self.event_handlers.iter() {
            if handler.type_id() == h.as_ref().type_id() {
                return Ok(());
            }
        }
        self.event_handlers.push(handler);
        if self.event_handlers.len() > 1 {
            return Ok(());
        }
        return self.start_event_handlers();
    }

    // overrides default contract name
    pub fn service_contract_name(&mut self, contract_name: &str) {
        self.sc_hname = wasmlib::ScHname::new(contract_name);
    }

    pub fn sign_requests(&mut self, key_pair: &keypair::KeyPair) {
        self.key_pair = Some(key_pair.clone());
    }

    pub fn unregister(&mut self, handler: Box<dyn IEventHandlers>) {
        self.event_handlers.retain(|h| {
            if handler.type_id() == h.as_ref().type_id() {
                return false;
            } else {
                return true;
            }
        });
        if self.event_handlers.len() == 0 {
            self.stop_event_handlers();
        }
    }

    pub fn wait_request(&mut self, req_id: Option<&ScRequestID>) {
        let r_id;
        let binding;
        match req_id {
            Some(id) => r_id = id,
            None => {
                binding = self.req_id.read().unwrap().to_owned();
                r_id = &binding;
            }
        };
        let res = self.svc_client.wait_until_request_processed(
            &self.chain_id,
            &r_id,
            std::time::Duration::new(60, 0),
        );

        if let Err(e) = res {
            self.err("WasmClientContext init err: ", &e)
        }
    }

    pub fn start_event_handlers(&'static self) -> errors::Result<()> {
        let (tx, rx): (mpsc::Sender<Vec<String>>, mpsc::Receiver<Vec<String>>) = mpsc::channel();
        let done = Arc::clone(&self.event_done);
        self.svc_client.subscribe_events(tx, done).unwrap();

        self.process_event(rx).unwrap();

        return Ok(());
    }

    pub fn stop_event_handlers(&self) {
        let mut done = self.event_done.write().unwrap();
        *done = true;
    }

    fn process_event(&'static self, rx: mpsc::Receiver<Vec<String>>) -> errors::Result<()> {
        for msg in rx {
            spawn(move || {
                let l = self.event_received.clone();
                if msg[0] == isc::waspclient::ISC_EVENT_KIND_ERROR {
                    let mut received = l.write().unwrap();
                    *received = true;
                    return Err(msg[1].clone());
                }

                if msg[0] != isc::waspclient::ISC_EVENT_KIND_SMART_CONTRACT
                    && msg[1] != self.chain_id.to_string()
                {
                    // not intended for us
                    return Ok(());
                }
                let mut params: Vec<String> = msg[6].split("|").map(|s| s.into()).collect();
                for i in 0..params.len() {
                    params[i] = self.unescape(&params[i]);
                }
                let topic = params.remove(0);

                for handler in self.event_handlers.iter() {
                    handler.as_ref().call_handler(&topic, &params);
                }

                let mut received = l.write().unwrap();
                *received = true;
                return Ok(());
            });
        }

        return Ok(());
    }

    pub fn wait_event(&self) -> errors::Result<()> {
        for _ in 0..100 {
            if *self.event_received.read().unwrap() {
                return Ok(());
            }
            std::thread::sleep(std::time::Duration::from_millis(100));
        }
        let err_msg = String::from("event wait timeout");
        self.err(&err_msg, "");
        return Err(err_msg);
    }

    fn unescape(&self, param: &str) -> String {
        let idx = match param.find("~") {
            Some(idx) => idx,
            None => return String::from(param),
        };
        match param.chars().nth(idx + 1).unwrap() {
            // escaped escape character
            '~' => param[0..idx].to_string() + "~" + &self.unescape(&param[idx + 2..]),
            // escaped vertical bar
            '/' => param[0..idx].to_string() + "|" + &self.unescape(&param[idx + 2..]),
            // escaped space
            '_' => param[0..idx].to_string() + " " + &self.unescape(&param[idx + 2..]),
            _ => panic!("invalid event encoding"),
        }
    }
    pub fn err(&self, current_layer_msg: &str, e: &str) {
        let mut err = self.error.write().unwrap();
        *err = Err(current_layer_msg.to_string() + e);
        drop(err);
    }
}

impl Default for WasmClientContext {
    fn default() -> WasmClientContext {
        WasmClientContext {
            chain_id: chain_id_from_bytes(&[]),
            error: Arc::new(RwLock::new(Ok(()))),
            event_done: Arc::default(),
            event_handlers: Vec::new(),
            event_received: Arc::default(),
            hrp: String::from(""),
            key_pair: None,
            nonce: Arc::new(RwLock::new(0)),
            req_id: Arc::new(RwLock::new(request_id_from_bytes(&[]))),
            sc_name: String::new(),
            sc_hname: ScHname(0),
            svc_client: WasmClientService::default(),
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::*;
    use wasmlib::*;

    struct FakeEventHandler {}
    impl IEventHandlers for FakeEventHandler {
        fn call_handler(&self, _topic: &str, _params: &Vec<String>) {}
    }

    #[test]
    fn test_wasm_client_context_new() {
        let svc_client = wasmclientservice::WasmClientService::default();
        let chain_id = wasmlib::chain_id_from_bytes(&vec![
            41, 180, 220, 182, 186, 38, 166, 60, 91, 105, 181, 183, 219, 243, 200, 162, 131, 181,
            57, 142, 41, 30, 236, 92, 178, 1, 116, 229, 174, 86, 156, 210,
        ]);
        let sc_name = "sc_name";
        let ctx = wasmclientcontext::WasmClientContext::new(&svc_client, &chain_id, sc_name);
        assert!(ctx.svc_client == svc_client);
        assert!(ctx.sc_name == sc_name);
        assert!(ctx.sc_hname == wasmlib::ScHname::new(sc_name));
        assert!(ctx.chain_id == chain_id);
        assert!(ctx.event_handlers.len() == 0);
        // assert!(ctx.key_pair == None);
        assert!(*ctx.req_id.read().unwrap() == wasmlib::request_id_from_bytes(&[]));
    }

    #[test]
    fn test_register() {}

    #[test]
    fn test_call_view_by_hname() {}

    #[test]
    fn test_unescape() {
        let ctx = WasmClientContext::default();
        let res = ctx.unescape(r"before~~/after");
        println!("res: {}", res);
        assert!(res == "before~/after");
        let res = ctx.unescape(r"before~/after");
        assert!(res == "before|after");
        let res = ctx.unescape(r"before~_after");
        assert!(res == "before after");
    }
}
