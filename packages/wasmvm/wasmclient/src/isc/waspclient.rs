// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

pub use crate::*;
pub use reqwest::*;
pub use std::time::*;
pub use wasmlib::*;

const DEFAULT_OPTIMISTIC_READ_TIMEOUT: Duration = Duration::from_millis(1100);

#[derive(Clone, PartialEq)]
pub struct WaspClient {
    base_url: String,
    token: String,
}

impl WaspClient {
    pub fn new(base_url: &str) -> WaspClient {
        return WaspClient {
            base_url: base_url.to_string(),
            token: String::from(""),
        };
    }
    pub fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &ScDict,
        optimistic_read_timeout: Option<Duration>,
    ) -> errors::Result<()> {
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

        let client = reqwest::blocking::Client::builder()
            .timeout(deadline)
            .build()
            .unwrap();
        let _ = client.post(url).body(args.to_bytes()).send();

        Ok(())
    }
    pub fn post_offledger_request(
        &self,
        chain_id: &ScChainID,
        req: &offledgerrequest::OffLedgerRequestData,
    ) -> errors::Result<()> {
        let url = format!("{}/chain/{}/request", self.base_url, chain_id.to_string());
        let client = reqwest::blocking::Client::new();
        let _ = client.post(url).body(req.to_bytes()).send();
        Ok(())
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
        let client = reqwest::blocking::Client::builder()
            .timeout(timeout)
            .build()
            .unwrap();
        let _ = client.get(url).send();
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::waspclient;
    use httpmock::prelude::*;

    #[test]
    fn waspclient_new() {
        let client = waspclient::WaspClient::new("http://localhost");
        assert!(client.base_url == "http://localhost");
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

        let client = waspclient::WaspClient::new(&mock_server.base_url());
        let sc_chain_id = wasmlib::chain_id_from_bytes(&chain_id_bytes);
        let sc_contract_hname = wasmlib::hname_from_bytes(&wasmlib::uint32_to_bytes(0x89703a45));
        let sc_function_hname = wasmlib::hname_from_bytes(&wasmlib::uint32_to_bytes(0x78cc397a));
        let args = wasmlib::ScDict::new(&vec![]);
        let _ = client
            .call_view_by_hname(
                &sc_chain_id,
                &sc_contract_hname,
                &sc_function_hname,
                &args,
                None,
            )
            .unwrap();
        call_view_by_hname_mock.assert();
    }
}
