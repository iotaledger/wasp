// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::{
    any::Any,
    sync::{mpsc, Arc, Mutex, RwLock},
    thread::spawn,
};
use wasmclientsandbox::*;
use wasmlib::*;

use crate::*;

// TODO to handle the request in parallel, WasmClientContext must be static now.
// We need to solve this problem. By copying the vector of event_handlers, we may solve this problem
pub struct WasmClientContext {
    pub chain_id: ScChainID,
    pub error: Arc<RwLock<errors::Result<()>>>,
    event_done: Arc<RwLock<bool>>,
    pub event_handlers: Vec<Box<dyn IEventHandlers>>,
    pub event_received: Arc<RwLock<bool>>,
    pub key_pair: Option<keypair::KeyPair>,
    pub nonce: Mutex<u64>,
    pub req_id: Arc<RwLock<ScRequestID>>,
    pub sc_name: String,
    pub sc_hname: ScHname,
    pub svc_client: WasmClientService, //TODO Maybe  use 'dyn IClientService' for 'svc_client' instead of a struct
}

impl WasmClientContext {
    pub fn new(svc_client: &WasmClientService, chain_id: &str, sc_name: &str) -> WasmClientContext {
        unsafe {
            // local client implementations for sandboxed functions
            BECH32_DECODE = client_bech32_decode;
            BECH32_ENCODE = client_bech32_encode;
            HASH_NAME = client_hash_name;
        }

        // set the network prefix for the current network
        match codec::bech32_decode(chain_id) {
            Ok((hrp, _)) => unsafe {
                if HRP_FOR_CLIENT != hrp && HRP_FOR_CLIENT != "" {
                    panic!("WasmClient can only connect to one Tangle network per app");
                }
                HRP_FOR_CLIENT = hrp;
            },
            Err(e) => {
                let ctx = WasmClientContext::default();
                ctx.err("failed to init", e.as_str());
                return ctx;
            }
        };

        WasmClientContext {
            chain_id: chain_id_from_string(chain_id),
            error: Arc::new(RwLock::new(Ok(()))),
            event_done: Arc::new(RwLock::new(false)),
            event_handlers: Vec::new(),
            event_received: Arc::new(RwLock::new(false)),
            key_pair: None,
            nonce: Mutex::new(0),
            req_id: Arc::new(RwLock::new(request_id_from_bytes(&[]))),
            sc_name: sc_name.to_string(),
            sc_hname: hname_from_bytes(&codec::hname_bytes(&sc_name)),
            svc_client: svc_client.clone(),
        }
    }

    pub fn current_chain_id(&self) -> ScChainID {
        return self.chain_id;
    }

    pub fn current_keypair(&self) -> Option<keypair::KeyPair> {
        return self.key_pair.clone();
    }

    pub fn current_svc_client(&self) -> WasmClientService {
        return self.svc_client.clone();
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
        self.sc_hname = ScHname::new(contract_name);
    }

    pub fn sign_requests(&mut self, key_pair: &keypair::KeyPair) {
        self.key_pair = Some(key_pair.clone());
        // get last used nonce from accounts core contract
        let isc_agent = wasmlib::ScAgentID::from_address(&key_pair.address());
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

    pub fn wait_event(&self) -> errors::Result<()> {
        return self.wait_event_timeout(10000);
    }

    pub fn wait_event_timeout(&self, msec: u64) -> errors::Result<()> {
        for _ in 0..100 {
            if *self.event_received.read().unwrap() {
                return Ok(());
            }
            std::thread::sleep(std::time::Duration::from_millis(msec));
        }
        let err_msg = String::from("event wait timeout");
        self.err(&err_msg, "");
        return Err(err_msg);
    }

    pub fn wait_request(&self) {
        let req_id = self.req_id.read().unwrap().to_owned();
        self.wait_request_id(Some(&req_id));
    }

    pub fn wait_request_id(&self, req_id: Option<&ScRequestID>) {
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

    fn process_event(&'static self, rx: mpsc::Receiver<Vec<String>>) -> errors::Result<()> {
        for msg in rx {
            spawn(move || {
                let lock_received = self.event_received.clone();
                if msg[0] == isc::waspclient::ISC_EVENT_KIND_ERROR {
                    let mut received = lock_received.write().unwrap();
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

                let mut received = lock_received.write().unwrap();
                *received = true;
                return Ok(());
            });
        }

        return Ok(());
    }

    pub fn start_event_handlers(&'static self) -> errors::Result<()> {
        let (tx, rx): (mpsc::Sender<Vec<String>>, mpsc::Receiver<Vec<String>>) = mpsc::channel();
        let done = Arc::clone(&self.event_done);
        self.svc_client.subscribe_events(tx, done)?;
        self.process_event(rx)?;
        return Ok(());
    }

    pub fn stop_event_handlers(&self) {
        let mut done = self.event_done.write().unwrap();
        *done = true;
    }

    fn unescape(&self, param: &str) -> String {
        let idx = match param.find("~") {
            Some(idx) => idx,
            None => return String::from(param),
        };
        match param.chars().nth(idx + 1) {
            // escaped escape character
            Some('~') => param[0..idx].to_string() + "~" + &self.unescape(&param[idx + 2..]),
            // escaped vertical bar
            Some('/') => param[0..idx].to_string() + "|" + &self.unescape(&param[idx + 2..]),
            // escaped space
            Some('_') => param[0..idx].to_string() + " " + &self.unescape(&param[idx + 2..]),
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
            key_pair: None,
            nonce: Mutex::default(),
            req_id: Arc::new(RwLock::new(request_id_from_bytes(&[]))),
            sc_name: String::new(),
            sc_hname: ScHname(0),
            svc_client: WasmClientService::default(),
        }
    }
}

#[cfg(test)]
mod tests {
    use wasmlib::*;

    use crate::*;

    #[derive(Debug)]
    struct FakeEventHandler {}

    impl IEventHandlers for FakeEventHandler {
        fn call_handler(&self, _topic: &str, _params: &Vec<String>) {}
    }

    const MYCHAIN: &str = "atoi1prj5xunmvc8uka9qznnpu4yrhn3ftm3ya0wr2jvurwr209llw7xdyztcr6g";
    const MYSEED: &str = "0xa580555e5b84a4b72bbca829b4085a4725941f3b3702525f36862762d76c21f3";

    #[test]
    fn test_wasm_client_context_new() {
        let svc_client = wasmclientservice::WasmClientService::default();

        // FIXME use valid sc_name which meets the requirement of bech32
        let sc_name = "testwasmlib";
        let ctx = wasmclientcontext::WasmClientContext::new(&svc_client, MYCHAIN, sc_name);
        assert!(svc_client == ctx.svc_client);
        assert!(sc_name == ctx.sc_name);
        assert!(wasmlib::ScHname::new(sc_name) == ctx.sc_hname);
        assert!(MYCHAIN == ctx.chain_id.to_string());
        assert!(0 == ctx.event_handlers.len());
        assert!(None == ctx.key_pair);
        assert!(wasmlib::request_id_from_bytes(&[]) == *ctx.req_id.read().unwrap());
    }

    fn setup_client() -> WasmClientContext {
        let svc = WasmClientService::new("127.0.0.1:19090", "127.0.0.1:15550");
        let mut ctx = WasmClientContext::new(&svc, MYCHAIN, "testwasmlib");
        ctx.sign_requests(&keypair::KeyPair::from_sub_seed(
            &wasmlib::bytes_from_string(MYSEED),
            0,
        ));
        assert!(ctx.error.read().unwrap().is_ok());
        return ctx;
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
