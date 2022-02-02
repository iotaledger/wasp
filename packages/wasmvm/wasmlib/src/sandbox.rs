// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use crate::host::*;
use crate::wasmtypes;

// @formatter:off
pub const FN_ACCOUNT_ID            : i32 = -1;
pub const FN_BALANCE               : i32 = -2;
pub const FN_BALANCES              : i32 = -3;
pub const FN_BLOCK_CONTEXT         : i32 = -4;
pub const FN_CALL                  : i32 = -5;
pub const FN_CALLER                : i32 = -6;
pub const FN_CHAIN_ID              : i32 = -7;
pub const FN_CHAIN_OWNER_ID        : i32 = -8;
pub const FN_CONTRACT              : i32 = -9;
pub const FN_CONTRACT_CREATOR      : i32 = -10;
pub const FN_DEPLOY_CONTRACT       : i32 = -11;
pub const FN_ENTROPY               : i32 = -12;
pub const FN_EVENT                 : i32 = -13;
pub const FN_INCOMING_TRANSFER     : i32 = -14;
pub const FN_LOG                   : i32 = -15;
pub const FN_MINTED                : i32 = -16;
pub const FN_PANIC                 : i32 = -17;
pub const FN_PARAMS                : i32 = -18;
pub const FN_POST                  : i32 = -19;
pub const FN_REQUEST               : i32 = -20;
pub const FN_REQUEST_ID            : i32 = -21;
pub const FN_RESULTS               : i32 = -22;
pub const FN_SEND                  : i32 = -23;
pub const FN_STATE_ANCHOR          : i32 = -24;
pub const FN_TIMESTAMP             : i32 = -25;
pub const FN_TRACE                 : i32 = -26;
pub const FN_UTILS_BASE58_DECODE   : i32 = -27;
pub const FN_UTILS_BASE58_ENCODE   : i32 = -28;
pub const FN_UTILS_BLS_ADDRESS     : i32 = -29;
pub const FN_UTILS_BLS_AGGREGATE   : i32 = -30;
pub const FN_UTILS_BLS_VALID       : i32 = -31;
pub const FN_UTILS_ED25519_ADDRESS : i32 = -32;
pub const FN_UTILS_ED25519_VALID   : i32 = -33;
pub const FN_UTILS_HASH_BLAKE2B    : i32 = -34;
pub const FN_UTILS_HASH_NAME       : i32 = -35;
pub const FN_UTILS_HASH_SHA3       : i32 = -36;
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

pub struct ScSandbox {}

impl ScSandbox {
    // retrieve the agent id of this contract account
    pub fn accountID() -> wasmtypes::ScAgentID {
     wasmtypes::agent_id_from_bytes(sandbox(FN_ACCOUNT_ID, null))
    }

    pub fn balance(color: wasmtypes::ScColor) -> u64 {
         wasmtypes::uint64_from_bytes(sandbox(FN_BALANCE, &color.toBytes()))
    }

    // access the current balances for all assets
    pub fn balances() -> ScBalances {
        return ScAssets::new(sandbox(FN_BALANCES, null)).balances()
    }

    // calls a smart contract function
    fn callWithTransfer(hContract: wasmtypes::ScHname, hFunction: wasmtypes::ScHname, params: Option<ScDict>, transfer: Option<ScTransfers>) -> ScImmutableDict {
    if (params == None) {
    params = new ScDict([]);
    }
    if (transfer == None) {
    transfer = new ScTransfers();
    }
    const req = new wasmrequests.CallRequest();
    req.contract = hContract;
    req.function = hFunction;
    req.params = params.toBytes();
    req.transfer = transfer.toBytes();
    const res = sandbox(FN_CALL, req.bytes());
    return new ScDict(res).immutable()
    }

    // retrieve the chain id of the chain this contract lives on
    pub fn chainID() -> wasmtypes::ScChainID {
    return wasmtypes::chainIDFromBytes(sandbox(FN_CHAIN_ID, null))
    }

    // retrieve the agent id of the owner of the chain this contract lives on
    pub fn chainOwnerID() -> wasmtypes::ScAgentID {
    return wasmtypes::agentIDFromBytes(sandbox(FN_CHAIN_OWNER_ID, null))
    }

    // retrieve the hname of this contract
    pub fn contract() -> wasmtypes::ScHname {
    return wasmtypes::hnameFromBytes(sandbox(FN_CONTRACT, null))
    }

    // retrieve the agent id of the creator of this contract
    pub fn contractCreator() -> wasmtypes::ScAgentID {
    return wasmtypes::agentIDFromBytes(sandbox(FN_CONTRACT_CREATOR, null))
    }

    // logs informational text message
    pub fn log(text: string) -> void {
        sandbox(FN_LOG, wasmtypes::stringToBytes(text))
    }

    // logs error text message and then panics
    pub fn panic(text: string) -> void {
        sandbox(FN_PANIC, wasmtypes::stringToBytes(text))
    }

    // retrieve parameters passed to the smart contract function that was called
    pub fn params() -> ScImmutableDict {
        return new
        ScDict(sandbox(FN_PARAMS, null)).immutable()
    }

    // panics if condition is not satisfied
    pub fn require(cond: bool, msg: string) -> void {
        if (!cond) {
            this.panic(msg)
        }
    }

    pub fn results(results: ScDict) -> void {
        sandbox(FN_RESULTS, results.toBytes())
    }

    // deterministic time stamp fixed at the moment of calling the smart contract
    pub fn timestamp() -> u64 {
        return wasmtypes::uint64FromBytes(sandbox(FN_TIMESTAMP, null))
    }

    // logs debugging trace text message
    pub fn trace(text: string) -> void {
        sandbox(FN_TRACE, wasmtypes::stringToBytes(text));
    }

    // access diverse utility functions
    pub fn utility() -> ScSandboxUtils {
        return ScSandboxUtils {};
    }
}

pub struct ScSandboxView {}

impl ScSandboxView {
    pub fn rawState() -> ScImmutableState {
        return ScImmutableState {};
    }
}

pub struct ScSandboxFunc {}

impl ScSandboxFunc {
    pub fn rawState() -> ScState {
        return ScState {};
    }
}
