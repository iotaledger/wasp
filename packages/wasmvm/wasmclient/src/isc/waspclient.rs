// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::{
    sync::{Arc, mpsc, RwLock},
    thread::spawn,
    time::*,
};

use base64::{Engine as _, engine::general_purpose};
use reqwest::*;
use wasmlib::*;

use codec::*;

use crate::*;
use crate::offledgerrequest::OffLedgerRequestData;

const DEFAULT_OPTIMISTIC_READ_TIMEOUT: Duration = Duration::from_millis(1100);

pub const ISC_EVENT_KIND_NEW_BLOCK: &str = "new_block";
pub const ISC_EVENT_KIND_RECEIPT: &str = "receipt";
pub const ISC_EVENT_KIND_SMART_CONTRACT: &str = "contract";
pub const ISC_EVENT_KIND_ERROR: &str = "error";

#[derive(Clone, PartialEq, Debug)]
pub struct WaspClient {
    base_url: String,
    event_port: String,
    token: String,
}

impl WaspClient {
    pub fn new(base_url: &str, event_port: &str) -> WaspClient {
        return WaspClient {
            base_url: base_url.to_string(),
            event_port: event_port.to_string(),
            token: String::from(""),
        };
    }

    pub fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
        optimistic_read_timeout: Option<Duration>,
    ) -> errors::Result<Vec<u8>> {
        let deadline = match optimistic_read_timeout {
            Some(duration) => duration,
            None => DEFAULT_OPTIMISTIC_READ_TIMEOUT,
        };
        let url = format!(
            "{}/chain/{}/contract/{}/callviewbyhname/{}",
            self.base_url,
            chain_id.to_string(),
            contract_hname.to_string(),
            function_hname.to_string()
        );

        let client = blocking::Client::builder()
            .timeout(deadline)
            .build()
            .unwrap();
        let body = json_encode(args);
        let res = client.post(url).json(&body).send();

        match res {
            Ok(v) => match v.status() {
                StatusCode::OK => {
                    match v.json::<JsonResponse>() {
                        Ok(json_obj) => {
                            return Ok(json_decode(json_obj));
                        }
                        Err(e) => {
                            return Err(format!("parse post response failed: {}", e.to_string()));
                        }
                    };
                }
                failed_status_code => {
                    let status_code = failed_status_code.as_u16();
                    match v.json::<JsonError>() {
                        Ok(err_msg) => {
                            return Err(format!("{status_code}: {}", err_msg.message));
                        }
                        Err(e) => return Err(e.to_string()),
                    }
                }
            },
            Err(e) => {
                return Err(format!("call() request failed: {}", e.to_string()));
            }
        }
    }

    pub fn post_offledger_request(
        &self,
        chain_id: &ScChainID,
        req: &OffLedgerRequestData,
    ) -> errors::Result<()> {
        let url = format!("{}/chain/{}/request", self.base_url, chain_id.to_string());
        let client = blocking::Client::new();
        let body = JsonPostRequest { request: general_purpose::STANDARD.encode(req.to_bytes()) };
        let res = client.post(url).json(&body).send();
        match res {
            Ok(v) => match v.status() {
                StatusCode::OK => {
                    return Ok(());
                }
                StatusCode::ACCEPTED => {
                    return Ok(());
                }
                failed_status_code => {
                    let status_code = failed_status_code.as_u16();
                    match v.json::<JsonError>() {
                        Ok(err_msg) => {
                            return Err(format!("{status_code}: {}", err_msg.message));
                        }
                        Err(e) => return Err(e.to_string()),
                    }
                }
            },
            Err(e) => {
                return Err(format!("post() request failed: {}", e.to_string()));
            }
        }
    }

    pub fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> errors::Result<()> {
        let url = format!(
            "{}/chain/{}/request/{}/wait",
            self.base_url,
            chain_id.to_string(),
            req_id.to_string()
        );
        let client = blocking::Client::builder()
            .timeout(timeout)
            .build()
            .unwrap();
        let res = client.get(url).header("Content-Type", "application/json").send();
        match res {
            Ok(v) => match v.status() {
                StatusCode::OK => {
                    return Ok(());
                }
                failed_status_code => {
                    let status_code = failed_status_code.as_u16();
                    match v.text() {
                        Ok(err_msg) => {
                            return Err(format!("{status_code}: {err_msg}"));
                        }
                        Err(e) => return Err(e.to_string()),
                    }
                }
            },
            Err(e) => {
                return Err(format!("request failed: {}", e.to_string()));
            }
        }
    }

    pub fn subscribe(&self, ch: mpsc::Sender<Vec<String>>, done: Arc<RwLock<bool>>) {
        // FIXME should not reconnect every time
        let (mut socket, _) = tungstenite::connect(&self.event_port).unwrap();
        let read_done = Arc::clone(&done);
        spawn(move || loop {
            match socket.read_message() {
                Ok(raw_msg) => {
                    let raw_msg_str = raw_msg.to_string();
                    if raw_msg_str != "" {
                        let msg: Vec<String> = raw_msg_str.split(" ").map(|s| s.into()).collect();
                        if msg[0] == ISC_EVENT_KIND_SMART_CONTRACT {
                            ch.send(msg).unwrap();
                        }
                    }
                }
                Err(tungstenite::Error::ConnectionClosed) => {
                    return Ok(());
                }
                Err(e) => {
                    return Err(format!("subscribe err: {}", e));
                }
            };

            if *read_done.read().unwrap() {
                socket.close(None).unwrap();
                let mut mut_done = read_done.write().unwrap();
                *mut_done = false;
                return Ok(());
            }
        });
    }
}

#[cfg(test)]
mod tests {
    use std::{
        net::TcpListener,
        sync::{Arc, mpsc, RwLock},
        thread::spawn,
    };

    use httpmock::prelude::*;
    use tungstenite::accept;
    use wasmlib::*;

    use crate::waspclient::WaspClient;

    #[test]
    fn waspclient_new() {
        let client = WaspClient::new("http://localhost:19090", "ws://localhost:15550");
        assert!(client.base_url == "http://localhost:19090");
        assert!(client.event_port == "ws://localhost:15550");
    }

    #[test]
    fn test_call_view_by_hname() {
        let chain_id_bytes = vec![
            41, 180, 220, 182, 186, 38, 166, 60, 91, 105, 181, 183, 219, 243, 200, 162, 131, 181,
            57, 142, 41, 30, 236, 92, 178, 1, 116, 229, 174, 86, 156, 210,
        ];
        let chain_id_str = "tgl1pq5mfh9khgn2v0zmdx6m0klnez3g8dfe3c53amzukgqhfedw26wdy8tztdy";
        let contract_hname = "testwasmlib";
        let function_hname = "tokenBalance";

        let mock_server = MockServer::start();
        let call_view_by_hname_mock = mock_server.mock(|when, then| {
            when.method(POST).path(format!(
                "/chain/{}/contract/{}/callviewbyhname/{}",
                chain_id_str, contract_hname, function_hname
            ));
            then.status(200);
        });

        let client = WaspClient::new(&mock_server.base_url(), "");
        let sc_chain_id = chain_id_from_bytes(&chain_id_bytes);
        let sc_contract_hname = hname_from_bytes(&uint32_to_bytes(0x89703a45));
        let sc_function_hname = hname_from_bytes(&uint32_to_bytes(0x78cc397a));
        let _ = client
            .call_view_by_hname(
                &sc_chain_id,
                &sc_contract_hname,
                &sc_function_hname,
                &[],
                None,
            )
            .unwrap();
        call_view_by_hname_mock.assert();
    }

    #[derive(Clone)]
    struct MockServerMsg {
        msg: String,
        count: u32,
    }

    #[test]
    fn test_subscribe() {
        let url = "ws://localhost:3012";
        let test_msg = "contract tgl1pp0j5wr5e5dxhk4hzwlgfs9vu7r025zeq0et7ftkrzmf8lwa44wy645r2hj | vm (contract): 89703a45: testwasmlib.test|1012000000|tgl1pp0j5wr5e5dxhk4hzwlgfs9vu7r025zeq0et7ftkrzmf8lwa44wy645r2hj|Lala";
        mock_server(
            url,
            Some(MockServerMsg {
                msg: test_msg.to_string(),
                count: 3,
            }),
        );
        let client = WaspClient::new("", &url);
        let (tx, rx): (mpsc::Sender<Vec<String>>, mpsc::Receiver<Vec<String>>) = mpsc::channel();
        let lock = Arc::new(RwLock::new(false));
        client.subscribe(tx, lock);
        let mut cnt = 0;
        for msgs in rx.iter() {
            let target_str = format!("{} cnt:{}", test_msg, cnt);
            let target: Vec<String> = target_str.split(" ").map(|s| s.into()).collect();
            for i in 0..msgs.len() {
                assert!(msgs[i] == target[i]);
            }
            cnt += 1;
        }
    }

    #[test]
    fn test_subscribe_stop() {
        let url = "ws://localhost:3013";
        let test_msg = "contract tgl1pp0j5wr5e5dxhk4hzwlgfs9vu7r025zeq0et7ftkrzmf8lwa44wy645r2hj | vm (contract): 89703a45: testwasmlib.test|1012000000|tgl1pp0j5wr5e5dxhk4hzwlgfs9vu7r025zeq0et7ftkrzmf8lwa44wy645r2hj|Lala";
        mock_server(
            url,
            Some(MockServerMsg {
                msg: test_msg.to_string(),
                count: 3,
            }),
        );
        let client = WaspClient::new("", &url);
        let (tx, rx): (mpsc::Sender<Vec<String>>, mpsc::Receiver<Vec<String>>) = mpsc::channel();
        let lock = Arc::new(RwLock::new(true));
        let sub_lock = Arc::clone(&lock);
        client.subscribe(tx, sub_lock);
        let mut cnt = 0;
        for msgs in rx.iter() {
            let target_str = format!("{} cnt:{}", test_msg, cnt);
            let target: Vec<String> = target_str.split(" ").map(|s| s.into()).collect();
            for i in 0..msgs.len() {
                assert!(msgs[i] == target[i]);
            }
            cnt += 1;
        }
        assert!(cnt == 1);
        assert!(*lock.read().unwrap() == false);
    }

    fn mock_server(input_url: &str, response_msg: Option<MockServerMsg>) {
        let ws_prefix = "ws://";
        let url = match input_url.to_string().strip_prefix(ws_prefix) {
            Some(v) => v.to_string(),
            None => input_url.to_string(),
        };
        spawn(move || {
            let server = TcpListener::bind(url).unwrap();
            for stream in server.incoming() {
                let mut socket = accept(stream.unwrap()).unwrap();
                let msg_opt = response_msg.to_owned();
                if !msg_opt.is_none() {
                    let msg = msg_opt.unwrap();
                    for cnt in 0..msg.count {
                        let msg = format!("{} cnt:{}", msg.msg, cnt);
                        socket
                            .write_message(tungstenite::Message::Text(msg).into())
                            .unwrap();
                    }
                    socket.close(None).unwrap();
                }
            }
        });
    }
}
