// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use std::rc::Rc;

use crate::*;
use crate::host::*;
use crate::wasmrequests::*;

// @formatter:off
pub const FN_ACCOUNT_ID               : i32 = -1;
pub const FN_ALLOWANCE                : i32 = -2;
pub const FN_BALANCE                  : i32 = -3;
pub const FN_BALANCES                 : i32 = -4;
pub const FN_BLOCK_CONTEXT            : i32 = -5;
pub const FN_CALL                     : i32 = -6;
pub const FN_CALLER                   : i32 = -7;
pub const FN_CHAIN_ID                 : i32 = -8;
pub const FN_CHAIN_OWNER_ID           : i32 = -9;
pub const FN_CONTRACT                 : i32 = -10;
pub const FN_DEPLOY_CONTRACT          : i32 = -11;
pub const FN_ENTROPY                  : i32 = -12;
pub const FN_ESTIMATE_STORAGE_DEPOSIT : i32 = -13;
pub const FN_EVENT                    : i32 = -14;
pub const FN_LOG                      : i32 = -15;
pub const FN_MINTED                   : i32 = -16;
pub const FN_PANIC                    : i32 = -17;
pub const FN_PARAMS                   : i32 = -18;
pub const FN_POST                     : i32 = -19;
pub const FN_REQUEST                  : i32 = -20;
pub const FN_REQUEST_ID               : i32 = -21;
pub const FN_REQUEST_SENDER           : i32 = -22;
pub const FN_RESULTS                  : i32 = -23;
pub const FN_SEND                     : i32 = -24;
pub const FN_STATE_ANCHOR             : i32 = -25;
pub const FN_TIMESTAMP                : i32 = -26;
pub const FN_TRACE                    : i32 = -27;
pub const FN_TRANSFER_ALLOWED         : i32 = -28;
pub const FN_UTILS_BECH32_DECODE      : i32 = -29;
pub const FN_UTILS_BECH32_ENCODE      : i32 = -30;
pub const FN_UTILS_BLS_ADDRESS        : i32 = -31;
pub const FN_UTILS_BLS_AGGREGATE      : i32 = -32;
pub const FN_UTILS_BLS_VALID          : i32 = -33;
pub const FN_UTILS_ED25519_ADDRESS    : i32 = -34;
pub const FN_UTILS_ED25519_VALID      : i32 = -35;
pub const FN_UTILS_HASH_BLAKE2B       : i32 = -36;
pub const FN_UTILS_HASH_NAME          : i32 = -37;
pub const FN_UTILS_HASH_SHA3          : i32 = -38;
// @formatter:on

// Direct logging of informational text to host log
pub fn log(text: &str) {
    sandbox(FN_LOG, text.as_bytes());
}

// Direct logging of error to host log, followed by panicking out of the Wasm code
pub fn panic(text: &str) {
    sandbox(FN_PANIC, text.as_bytes());
}

// Direct logging of debug trace text to host log
pub fn trace(text: &str) {
    sandbox(FN_TRACE, text.as_bytes());
}

pub fn params_proxy() -> Proxy {
    let buf = sandbox(FN_PARAMS, &[]);
    Proxy::new(&Rc::new(ScDict::new(&buf)))
}

pub fn results_proxy() -> Proxy {
    Proxy::new(&Rc::new(ScDict::new(&[])))
}

pub fn state_proxy() -> Proxy {
    Proxy::new(&Rc::new(ScDict::state()))
}

pub trait ScSandbox {
    // retrieve the agent id of this contract account
    fn account_id(&self) -> ScAgentID {
        agent_id_from_bytes(&sandbox(FN_ACCOUNT_ID, &[]))
    }

    fn balance(&self, token_id: ScTokenID) -> u64 {
        uint64_from_bytes(&sandbox(FN_BALANCE, &token_id.to_bytes()))
    }

    // access the current balances for all assets
    fn balances(&self) -> ScBalances {
        ScAssets::new(&sandbox(FN_BALANCES, &[])).balances()
    }

    // calls a smart contract function
    fn call_with_allowance(&self, h_contract: ScHname, h_function: ScHname, params: Option<ScDict>, allowance: Option<ScTransfer>) -> ScImmutableDict {
        let mut req = wasmrequests::CallRequest {
            contract: h_contract,
            function: h_function,
            params: vec![0; SC_UINT32_LENGTH],
            allowance: vec![0; SC_UINT32_LENGTH],
        };
        if let Some(params) = params {
            req.params = params.to_bytes();
        }
        if let Some(allowance) = allowance {
            req.allowance = allowance.to_bytes();
        }
        let buf = self.fn_call(&req);
        ScImmutableDict::new(ScDict::new(&buf))
    }

    fn fn_call(&self, req: &CallRequest) -> Vec<u8> {
        sandbox(FN_CALL, &req.to_bytes())
    }

    // retrieve the agent id of the owner of the chain this contract lives on
    fn chain_owner_id(&self) -> ScAgentID {
        agent_id_from_bytes(&sandbox(FN_CHAIN_OWNER_ID, &[]))
    }

    // retrieve the hname of this contract
    fn contract(&self) -> ScHname {
        hname_from_bytes(&sandbox(FN_CONTRACT, &[]))
    }

    // retrieve the chain id of the chain this contract lives on
    fn current_chain_id(&self) -> ScChainID {
        self.fn_chain_id()
    }

    fn fn_chain_id(&self) -> ScChainID {
        chain_id_from_bytes(&sandbox(FN_CHAIN_ID, &[]))
    }

    // logs informational text message
    fn log(&self, text: &str) {
        log(text);
    }

    // logs error text message and then panics
    fn panic(&self, text: &str) {
        panic(text);
    }

    // retrieve parameters passed to the smart contract function that was called
    fn params(&self) -> ScImmutableDict {
        let buf = sandbox(FN_PARAMS, &[]);
        ScImmutableDict::new(ScDict::new(&buf))
    }

    // panics if condition is not satisfied
    fn require(&self, cond: bool, msg: &str) {
        if !cond {
            panic(msg)
        }
    }

    fn results(&self, results: &Proxy) {
        sandbox(FN_RESULTS, &results.kv_store.to_bytes());
    }

    // deterministic time stamp fixed at the moment of calling the smart contract
    fn timestamp(&self) -> u64 {
        uint64_from_bytes(&sandbox(FN_TIMESTAMP, &[]))
    }

    // logs debugging trace text message
    fn trace(&self, text: &str) {
        trace(text);
    }

    // access diverse utility functions
    fn utility(&self) -> ScSandboxUtils {
        ScSandboxUtils {}
    }
}

pub trait ScSandboxView: ScSandbox {
    // calls a smart contract view
    fn call(&self, h_contract: ScHname, h_function: ScHname, params: Option<ScDict>) -> ScImmutableDict {
        return self.call_with_allowance(h_contract, h_function, params, None);
    }

    fn raw_state(&self) -> ScImmutableDict {
        ScImmutableDict::new(ScDict::state())
    }
}

pub trait ScSandboxFunc: ScSandbox {
    // access the allowance assets
    fn allowance(&self) -> ScBalances {
        let buf = sandbox(FN_ALLOWANCE, &[]);
        return ScAssets::new(&buf).balances();
    }

    //fn blockContext(&self, construct func(sandbox: ScSandbox) interface{}, onClose func(interface{})) -> interface{} {
    //	panic("implement me")
    //}

    // calls a smart contract func or view
    fn call(&self, h_contract: ScHname, h_function: ScHname, params: Option<ScDict>, allowance: Option<ScTransfer>) -> ScImmutableDict {
        return self.call_with_allowance(h_contract, h_function, params, allowance);
    }

    // retrieve the agent id of the caller of the smart contract
    fn caller(&self) -> ScAgentID {
        return agent_id_from_bytes(&sandbox(FN_CALLER, &[]));
    }

    // deploys a smart contract
    fn deploy_contract(&self, program_hash: &ScHash, name: &str, description: &str, init_params: Option<ScDict>) {
        let mut req = wasmrequests::DeployRequest {
            prog_hash: program_hash.clone(),
            name: name.to_string(),
            description: description.to_string(),
            params: vec![0; SC_UINT32_LENGTH],
        };
        if let Some(init_params) = init_params {
            req.params = init_params.to_bytes();
        }
        sandbox(FN_DEPLOY_CONTRACT, &req.to_bytes());
    }

    // returns random entropy data for current request.
    fn entropy(&self) -> ScHash {
        return hash_from_bytes(&sandbox(FN_ENTROPY, &[]));
    }

    fn estimate_storage_deposit(&self, f: &ScFunc) -> u64 {
        let req = f.post_request(ScFuncContext {}.current_chain_id());
        uint64_from_bytes(&sandbox(FN_ESTIMATE_STORAGE_DEPOSIT, &req.to_bytes()))
    }

    // signals an event on the node that external entities can subscribe to
    fn event(&self, msg: &str) {
        sandbox(FN_EVENT, &string_to_bytes(msg));
    }

    // retrieve the assets that were minted in this transaction
    fn minted(&self) -> ScBalances {
        let buf = sandbox(FN_MINTED, &[]);
        return ScAssets::new(&buf).balances();
    }

    // Post (delayed) posts a SC function request
    fn post(&self, chain_id: ScChainID, h_contract: ScHname, h_function: ScHname, params: ScDict, allowance: ScTransfer, transfer: ScTransfer, delay: u32) {
        let req = wasmrequests::PostRequest {
            chain_id,
            contract: h_contract,
            function: h_function,
            params: params.to_bytes(),
            allowance: allowance.to_bytes(),
            transfer: transfer.to_bytes(),
            delay: delay,
        };
        self.fn_post(&req);
    }

    fn fn_post(&self, req: &PostRequest) -> Vec<u8> {
        sandbox(FN_POST, &req.to_bytes())
    }

    // generates a random value from 0 to max (exclusive: max) using a deterministic RNG
    fn random(&self, max: u64) -> u64 {
        if max == 0 {
            panic("random: max parameter should be > 0");
        }
        unsafe {
            static mut ENTROPY: Vec<u8> = Vec::new();
            static mut OFFSET: usize = 0;
            // note that entropy gets reset for every request
            if ENTROPY.len() == 0 {
                // first time in self: request, initialize with current request entropy
                ENTROPY = self.entropy().to_bytes();
                OFFSET = 0;
            }
            if OFFSET == 32 {
                // ran out of entropy: data, hash entropy for next pseudo-random entropy
                ENTROPY = self.utility().hash_blake2b(&ENTROPY).to_bytes();
                OFFSET = 0;
            }
            let rnd = uint64_from_bytes(&ENTROPY[OFFSET..OFFSET + 8]) % max;
            OFFSET += 8;
            return rnd;
        }
    }

    fn raw_state(&self) -> ScDict {
        ScDict::state()
    }

    //fn request(&self) -> ScRequest {
    //	panic("implement me");
    //}

    // retrieve the request id of this transaction
    fn request_id(&self) -> ScRequestID {
        return request_id_from_bytes(&sandbox(FN_REQUEST_ID, &[]));
    }

    // retrieve the request sender of this transaction
    fn request_sender(&self) -> ScAgentID {
        return agent_id_from_bytes(&sandbox(FN_REQUEST_SENDER, &[]));
    }

    // Send transfers SC assets to the specified address
    fn send(&self, address: &ScAddress, transfer: &ScTransfer) {
        // we need some assets to send
        if transfer.is_empty() {
            return;
        }

        let req = wasmrequests::SendRequest {
            address: address.clone(),
            transfer: transfer.to_bytes(),
        };
        sandbox(FN_SEND, &req.to_bytes());
    }

    //fn stateAnchor(&self) -> interface{} {
    //	panic("implement me")
    //}

    // TransferAllowed transfers allowed assets from caller to the specified account
    fn transfer_allowed(&self, agent_id: &ScAgentID, transfer: &ScTransfer) {
        // we need some assets to send
        if transfer.is_empty() {
            return;
        }

        let req = wasmrequests::TransferRequest {
            agent_id: agent_id.clone(),
            transfer: transfer.to_bytes(),
        };
        sandbox(FN_TRANSFER_ALLOWED, &req.to_bytes());
    }
}
