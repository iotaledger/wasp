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
    pub topic: String,
    pub timestamp: u64,
    pub payload: Vec<u8>,
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
                        let buf = hex_decode(&parts[1]);
                        let mut dec = WasmDecoder::new(&buf);
                        let topic = string_decode(&mut dec);
                        let payload = dec.fixed_bytes(dec.length() as usize);
                        let event = ContractEvent {
                            chain_id: chain_id_from_string(&json.chain_id),
                            contract_id: hname_from_string(&parts[0]),
                            topic: topic,
                            payload: payload,
                            timestamp: uint64_from_bytes(&payload[..SC_UINT64_LENGTH]),
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
        let sep = event.data.find('|');
        if sep.is_none() {
            return;
        }
        let sep = sep.unwrap();
        let topic = &event.data[..sep];
        println!("{} {} {}", event.chain_id.to_string(), event.contract_id.to_string(), topic);
        let buf = hex_decode(&event.data[sep + 1..]);
        let mut dec = WasmDecoder::new(&buf);
        self.handler.call_handler(topic, &mut dec);
    }

    fn subscribe(sender: &Sender, topic: &str) {
        let cmd = SubscriptionCommand {
            command: String::from("subscribe"),
            topic: String::from(topic),
        };
        let json = serde_json::to_string(&cmd).unwrap();
        let _ = sender.send(json);
    }
}
