// // Copyright 2020 IOTA Stiftung
// // SPDX-License-Identifier: Apache-2.0

use std::time::Duration;

use wasmlib::*;

use isc::offledgerrequest::*;

use crate::*;
use crate::keypair::KeyPair;
use crate::waspclient::WaspClient;

pub trait IClientService {
    fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
    ) -> errors::Result<Vec<u8>>;

    fn post_request(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
        allowance: &ScAssets,
        key_pair: &KeyPair,
        nonce: u64,
    ) -> errors::Result<ScRequestID>;

    fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> errors::Result<()>;
}

#[derive(Clone, PartialEq)]
pub struct WasmClientService {
    client: WaspClient,
    last_err: errors::Result<()>,
}

impl IClientService for WasmClientService {
    fn call_view_by_hname(
        &self,
        chain_id: &ScChainID,
        contract_hname: &ScHname,
        function_hname: &ScHname,
        args: &[u8],
    ) -> errors::Result<Vec<u8>> {
        return self.client.call_view_by_hname(
            chain_id,
            contract_hname,
            function_hname,
            args,
            None,
        );
    }

    fn post_request(
        &self,
        chain_id: &ScChainID,
        h_contract: &ScHname,
        h_function: &ScHname,
        args: &[u8],
        allowance: &ScAssets,
        key_pair: &KeyPair,
        nonce: u64,
    ) -> errors::Result<ScRequestID> {
        let mut req: OffLedgerRequestData =
            OffLedgerRequest::new(
                chain_id,
                h_contract,
                h_function,
                args,
                nonce,
            );
        req.with_allowance(&allowance);
        let signed = req.sign(key_pair);
        let res = self.client.post_offledger_request(&chain_id, &signed);
        if let Err(e) = res {
            return Err(e);
        }
        Ok(signed.id())
    }

    fn wait_until_request_processed(
        &self,
        chain_id: &ScChainID,
        req_id: &ScRequestID,
        timeout: Duration,
    ) -> errors::Result<()> {
        return self
            .client
            .wait_until_request_processed(&chain_id, req_id, timeout);
    }
}

impl WasmClientService {
    pub fn new(wasp_api: &str, event_port: &str) -> Self {
        return WasmClientService {
            client: WaspClient::new(wasp_api, event_port),
            last_err: Ok(()),
        };
    }
}

impl Default for WasmClientService {
    fn default() -> Self {
        return WasmClientService {
            client: WaspClient::new("127.0.0.1:19090", "127.0.0.1:15550"),
            last_err: Ok(()),
        };
    }
}

// impl std::fmt::Debug for WasmClientService {
//     fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> core::result::Result<(), std::fmt::Error> {
//         f.debug_tuple("WasmClientService")
//             .field(&self.client)
//             .field(&self.event_port)
//             .finish()
//     }
// }

#[cfg(test)]
mod tests {
    use crate::isc::waspclient;
    use crate::WasmClientService;
    use crate::waspclient::WaspClient;

    #[test]
    fn service_default() {
        let service = WasmClientService::default();
        let default_service = WasmClientService {
            client: WaspClient::new("127.0.0.1:19090", "127.0.0.1:15550"),
            last_err: Ok(()),
        };
        assert!(default_service.event_port == service.event_port);
        assert!(default_service.last_err == Ok(()));
    }
}
