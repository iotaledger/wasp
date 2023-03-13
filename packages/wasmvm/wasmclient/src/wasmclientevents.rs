// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use std::sync::{Arc, mpsc, Mutex};
use std::thread::{JoinHandle, spawn};

use serde::{Deserialize, Serialize};
use wasmlib::*;
use ws::{CloseCode, connect, Message, Sender};

pub const ISC_EVENT_KIND_NEW_BLOCK: &str = "new_block";
pub const ISC_EVENT_KIND_RECEIPT: &str = "receipt";
pub const ISC_EVENT_KIND_SMART_CONTRACT: &str = "contract";
pub const ISC_EVENT_KIND_ERROR: &str = "error";

#[derive(Serialize, Deserialize)]
pub struct SubscriptionCommand {
    pub command: String,
    pub topic: String,
}

#[derive(Serialize, Deserialize)]
pub struct EventMessage {
    #[serde(rename = "kind")]
    pub kind: String,
    #[serde(rename = "issuer")]
    pub issuer: String,
    #[serde(rename = "requestID")]
    pub request_id: String,
    #[serde(rename = "chainID")]
    pub chain_id: String,
    #[serde(rename = "payload")]
    pub payload: Vec<String>,
}

pub struct ContractEvent {
    pub chain_id: ScChainID,
    pub contract_id: ScHname,
    pub data: String,
}

pub struct WasmClientEvents {
    pub(crate) chain_id: ScChainID,
    pub(crate) contract_id: ScHname,
    pub(crate) handler: Box<dyn IEventHandlers>,
}

impl WasmClientEvents {
    pub(crate) fn start_event_loop(socket_url: String, close_rx: Arc<Mutex<mpsc::Receiver<bool>>>, event_handlers: Arc<Mutex<Vec<WasmClientEvents>>>) -> JoinHandle<()> {
        spawn(move || {
            connect(socket_url, |out| {
                // on connect start the thread that allows interrupting the message handler thread
                // note that we did not know the `out` websocket until this point, so we use an
                // external channel to this thread to signal that the websocket can be closed
                let close_rx = close_rx.clone();
                let socket = out.clone();
                spawn(move || {
                    close_rx.lock().unwrap().recv().unwrap();
                    println!("Closing websocket");
                    socket.close(CloseCode::Normal).unwrap();
                });

                // tell API to send us block events for all chains
                Self::subscribe(&out, "chains");
                Self::subscribe(&out, "block_events");

                // return the message handler closure that will be called from the message loop
                Self::event_loop(event_handlers.clone())
            }).unwrap();
            println!("Exiting message handler");
        })
    }

    fn event_loop(event_handlers: Arc<Mutex<Vec<WasmClientEvents>>>) -> Box<dyn Fn(Message) -> ws::Result<()>> {
        let f = Box::new(move |msg: Message| {
            // println!("Message: {}", msg);
            if let Ok(text) = msg.as_text() {
                if let Ok(json) = serde_json::from_str::<EventMessage>(text) {
                    for item in json.payload {
                        let parts: Vec<String> = item.split(": ").map(|s| s.into()).collect();
                        let event = ContractEvent {
                            chain_id: chain_id_from_string(&json.chain_id),
                            contract_id: hname_from_string(&parts[0]),
                            data: parts[1].clone(),
                        };
                        let event_handlers = event_handlers.lock().unwrap();
                        for event_processor in event_handlers.iter() {
                            event_processor.process_event(&event);
                        }
                    }
                }
            }
            return Ok(());
        });
        f
    }

    fn process_event(&self, event: &ContractEvent) {
        if event.contract_id != self.contract_id || event.chain_id != self.chain_id {
            return;
        }
        println!("{} {} {}", event.chain_id.to_string(), event.contract_id.to_string(), event.data);

        let mut params: Vec<String> = event.data.split("|").map(|s| s.into()).collect();
        for i in 0..params.len() {
            params[i] = Self::unescape(&params[i]);
        }
        let topic = params.remove(0);
        self.handler.call_handler(&topic, &params);
    }

    fn subscribe(sender: &Sender, topic: &str) {
        let cmd = SubscriptionCommand {
            command: String::from("subscribe"),
            topic: String::from(topic),
        };
        let json = serde_json::to_string(&cmd).unwrap();
        let _ = sender.send(json);
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
}
