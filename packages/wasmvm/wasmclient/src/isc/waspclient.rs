// pub use crate::gas::*;
pub use crate::offledgerrequest::*;
pub use crate::receipt::*;
use hyper::{
    client::HttpConnector,
    {Body, Client},
};
pub use std::time::*;
pub use wasmlib::*;

const DEFAULT_OPTIMISTIC_READ_TIMEOUT: Duration = Duration::from_millis(1100);

pub struct WaspClient {
    http_client: Client<HttpConnector, Body>,
    base_url: String,
    token: String,
}

impl WaspClient {
    pub fn new(base_url: &str, http_client: Option<Client<HttpConnector, Body>>) -> WaspClient {
        match http_client {
            Some(client) => {
                return WaspClient {
                    http_client: client,
                    base_url: base_url.to_string(),
                    token: String::from(""),
                }
            }
            None => {
                let client = hyper::Client::new();
                return WaspClient {
                    http_client: client,
                    base_url: base_url.to_string(),
                    token: String::from(""),
                };
            }
        }
    }
    pub fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: ScDict,
        optimistic_read_timeout: Option<Duration>,
    ) -> Result<ScDict, String> {
        let now = SystemTime::now();
        let deadline = match optimistic_read_timeout {
            Some(duration) => now.checked_add(duration).unwrap(),
            None => now.checked_add(DEFAULT_OPTIMISTIC_READ_TIMEOUT).unwrap(),
        };

        todo!()
    }
    pub fn post_offledger_request(
        &self,
        chain_id: &ScChainID,
        req: &OffLedgerRequestData,
    ) -> Result<(), String> {
        todo!()
    }
    pub fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> Result<Receipt, String> {
        todo!()
    }
}

fn send_request(method: &str, route: &str) -> Result<Vec<u8>, String> {
    todo!()
}
