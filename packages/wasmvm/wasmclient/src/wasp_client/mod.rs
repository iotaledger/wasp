// WaspClient allows to make requests to the Wasp web API.
use hyper::{
    client::HttpConnector,
    {Body, Client},
};
pub mod gas;
pub mod offledger;
pub mod receipt;

pub use gas::*;
pub use offledger::*;
pub use receipt::*;
pub use std::time::*;
pub use wasmlib::*;
const default_optimistic_read_timeout: Duration = Duration::from_millis(1100);

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
    pub fn CallViewByHname(
        &self,
        chain_id: &ScChainID,
        contract_hname: ScHname,
        function_hname: ScHname,
        args: ScDict,
        optimistic_read_timeout: Option<Duration>,
    ) -> (Result<ScDict, String>) {
        let now = SystemTime::now();
        let mut deadline = match optimistic_read_timeout {
            Some(duration) => now.checked_add(duration).unwrap(),
            None => now.checked_add(default_optimistic_read_timeout).unwrap(),
        };

        // let dict = ScDict::new(&vec![0]);
        // ScDict::read_bytes()

        return Ok(ScDict::new(&vec![1, 2]));
    }
    pub fn PostOffLedgerRequest(
        &self,
        chain_id: &ScChainID,
        req: OffLedgerRequestData,
    ) -> Result<(), String> {
        // TODO err return with request ID
        Err("not impl".to_string())
    }
    pub fn WaitUntilRequestProcessed(
        &self,
        chain_id: &ScChainID,
        req_id: ScRequestID,
        timeout: Duration,
    ) -> Result<Receipt, String> {
        Err("not impl".to_string())
    }
}

fn send_request(method: &str, route: &str) -> Result<Vec<u8>, String> {
    Err("not impl".to_string())
}
