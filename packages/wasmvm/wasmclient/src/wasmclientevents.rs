// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use std::sync::{Arc, mpsc, Mutex};
use std::thread::{JoinHandle, spawn};
use base64::Engine;
use base64::engine::general_purpose;

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

#[derive(Deserialize)]
pub struct ISCPayload {
    #[serde(rename = "contractID")]
    pub contract_id: u32,
    #[serde(rename = "topic")]
    pub topic: String,
    #[serde(rename = "timestamp")]
    pub timestamp: u64,
    #[serde(rename = "payload")]
    pub payload: String,
}

#[derive(Deserialize)]
pub struct ISCEvent {
    #[serde(rename = "kind")]
    pub kind: String,
    #[serde(rename = "issuer")]
    pub issuer: String,
    #[serde(rename = "requestID")]
    pub request_id: String,
    #[serde(rename = "chainID")]
    pub chain_id: String,
    #[serde(rename = "payload")]
    pub payload: Vec<ISCPayload>,
}

pub struct Event {
    pub chain_id: ScChainID,
    pub contract_id: ScHname,
    pub payload: Vec<u8>,
    pub timestamp: u64,
    pub topic: String,
}

impl Event {
    pub fn new(chain_id: &str, event: &ISCPayload) -> Event {
        let mut payload = uint64_to_bytes(event.timestamp);
        payload.append(&mut general_purpose::STANDARD.decode(&event.payload).unwrap());
        Event {
            chain_id: chain_id_from_string(chain_id),
            contract_id: ScHname(event.contract_id),
            timestamp: event.timestamp,
            topic: event.topic.clone(),
            payload: payload,
        }
    }
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
                if let Ok(json) = serde_json::from_str::<ISCEvent>(text) {
                    for item in json.payload {
                        let event = Event::new(&json.chain_id, &item);
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

    fn process_event(&self, event: &Event) {
        if event.contract_id != self.contract_id || event.chain_id != self.chain_id {
            return;
        }
        println!("{} {} {}", event.chain_id.to_string(), event.contract_id.to_string(), event.topic.to_string());
        let mut dec = WasmDecoder::new(&event.payload);
        self.handler.call_handler(&event.topic, &mut dec);
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
