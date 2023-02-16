// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::time::*;

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
        let url = format!("{}/requests/callview", self.base_url);

        let client = blocking::Client::builder()
            .timeout(deadline)
            .build()
            .unwrap();
        let body = APICallViewRequest {
            arguments: json_encode(args),
            chain_id: chain_id.to_string(),
            contract_hname: contract_hname.to_string(),
            function_hname: function_hname.to_string(),
        };
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
        let url = format!("{}/requests/offledger", self.base_url);
        let client = blocking::Client::new();
        let body = APIOffLedgerRequest {
            chain_id: chain_id.to_string(),
            request: hex_encode(&req.to_bytes()),
        };
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
            "{}/chains/{}/requests/{}/wait",
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
}
