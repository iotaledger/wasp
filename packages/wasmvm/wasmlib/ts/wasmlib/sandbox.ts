// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmrequests from "./wasmrequests"
import * as wasmtypes from "./wasmtypes"
import {ScAssets, ScBalances, ScTransfers} from "./assets";
import {ScDict, ScImmutableDict} from "./dict";
import {sandbox} from "./host";
import {ScSandboxUtils} from "./sandboxutils";
import {ScImmutableState, ScState} from "./state";

// @formatter:off
export const FnAccountID           : i32 = -1;
export const FnBalance             : i32 = -2;
export const FnBalances            : i32 = -3;
export const FnBlockContext        : i32 = -4;
export const FnCall                : i32 = -5;
export const FnCaller              : i32 = -6;
export const FnChainID             : i32 = -7;
export const FnChainOwnerID        : i32 = -8;
export const FnContract            : i32 = -9;
export const FnContractCreator     : i32 = -10;
export const FnDeployContract      : i32 = -11;
export const FnEntropy             : i32 = -12;
export const FnEvent               : i32 = -13;
export const FnIncomingTransfer    : i32 = -14;
export const FnLog                 : i32 = -15;
export const FnMinted              : i32 = -16;
export const FnPanic               : i32 = -17;
export const FnParams              : i32 = -18;
export const FnPost                : i32 = -19;
export const FnRequest             : i32 = -20;
export const FnRequestID           : i32 = -21;
export const FnResults             : i32 = -22;
export const FnSend                : i32 = -23;
export const FnStateAnchor         : i32 = -24;
export const FnTimestamp           : i32 = -25;
export const FnTrace               : i32 = -26;
export const FnUtilsBase58Decode   : i32 = -27;
export const FnUtilsBase58Encode   : i32 = -28;
export const FnUtilsBlsAddress     : i32 = -29;
export const FnUtilsBlsAggregate   : i32 = -30;
export const FnUtilsBlsValid       : i32 = -31;
export const FnUtilsEd25519Address : i32 = -32;
export const FnUtilsEd25519Valid   : i32 = -33;
export const FnUtilsHashBlake2b    : i32 = -34;
export const FnUtilsHashName       : i32 = -35;
export const FnUtilsHashSha3       : i32 = -36;
// @formatter:on

// Direct logging of text to host log
export function log(text: string): void {
    sandbox(FnLog, wasmtypes.stringToBytes(text));
}

// Direct logging of error to host log, followed by panicking out of the Wasm code
export function panic(text: string): void {
    sandbox(FnPanic, wasmtypes.stringToBytes(text));
}

// Direct conditional logging of debug-level informational text to host log
export function trace(text: string): void {
    sandbox(FnTrace, wasmtypes.stringToBytes(text));
}

export function paramsProxy(): wasmtypes.Proxy {
    return new ScDict(sandbox(FnParams, null)).asProxy();
}

export class ScSandbox {
    // retrieve the agent id of this contract account
    public accountID(): wasmtypes.ScAgentID {
        return wasmtypes.agentIDFromBytes(sandbox(FnAccountID, null));
    }

    public balance(color: wasmtypes.ScColor): u64 {
        return wasmtypes.uint64FromBytes(sandbox(FnBalance, color.toBytes()));
    }

    // access the current balances for all assets
    public balances(): ScBalances {
        return new ScAssets(sandbox(FnBalances, null)).balances();
    }

    // calls a smart contract function
    protected callWithTransfer(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict | null, transfer: ScTransfers | null): ScImmutableDict {
        if (params === null) {
            params = new ScDict([]);
        }
        if (transfer === null) {
            transfer = new ScTransfers();
        }
        const req = new wasmrequests.CallRequest();
        req.contract = hContract;
        req.function = hFunction;
        req.params = params.toBytes();
        req.transfer = transfer.toBytes();
        const res = sandbox(FnCall, req.bytes());
        return new ScDict(res).immutable();
    }

    // retrieve the chain id of the chain this contract lives on
    public chainID(): wasmtypes.ScChainID {
        return wasmtypes.chainIDFromBytes(sandbox(FnChainID, null));
    }

    // retrieve the agent id of the owner of the chain this contract lives on
    public chainOwnerID(): wasmtypes.ScAgentID {
        return wasmtypes.agentIDFromBytes(sandbox(FnChainOwnerID, null));
    }

    // retrieve the hname of this contract
    public contract(): wasmtypes.ScHname {
        return wasmtypes.hnameFromBytes(sandbox(FnContract, null));
    }

    // retrieve the agent id of the creator of this contract
    public contractCreator(): wasmtypes.ScAgentID {
        return wasmtypes.agentIDFromBytes(sandbox(FnContractCreator, null));
    }

    // logs informational text message
    public log(text: string): void {
        sandbox(FnLog, wasmtypes.stringToBytes(text));
    }

    // logs error text message and then panics
    public panic(text: string): void {
        sandbox(FnPanic, wasmtypes.stringToBytes(text));
    }

    // retrieve parameters passed to the smart contract function that was called
    public params(): ScImmutableDict {
        return new ScDict(sandbox(FnParams, null)).immutable();
    }

    // panics if condition is not satisfied
    public require(cond: bool, msg: string): void {
        if (!cond) {
            this.panic(msg);
        }
    }

    public results(results: ScDict): void {
        sandbox(FnResults, results.toBytes());
    }

    // deterministic time stamp fixed at the moment of calling the smart contract
    public timestamp(): u64 {
        return wasmtypes.uint64FromBytes(sandbox(FnTimestamp, null));
    }

    // logs debugging trace text message
    public trace(text: string): void {
        sandbox(FnTrace, wasmtypes.stringToBytes(text));
    }

    // access diverse utility functions
    public utility(): ScSandboxUtils {
        return new ScSandboxUtils();
    }
}

export class ScSandboxView extends ScSandbox {
    // calls a smart contract view
    public call(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict | null): ScImmutableDict {
        return this.callWithTransfer(hContract, hFunction, params, null);
    }

    public rawState(): ScImmutableState {
        return new ScImmutableState();
    }
}

export class ScSandboxFunc extends ScSandbox {
    private static entropy: u8[] = [];
    private static offset: u32 = 0;

    //public blockContext(construct func(sandbox: ScSandbox) interface{}, onClose func(interface{})): interface{} {
    //	panic("implement me")
    //}

    // calls a smart contract function
    public call(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict | null, transfer: ScTransfers | null): ScImmutableDict {
        return this.callWithTransfer(hContract, hFunction, params, transfer);
    }

    // retrieve the agent id of the caller of the smart contract
    public caller(): wasmtypes.ScAgentID {
        return wasmtypes.agentIDFromBytes(sandbox(FnCaller, null));
    }

    // deploys a smart contract
    public deployContract(programHash: wasmtypes.ScHash, name: string, description: string, initParams: ScDict | null): void {
        if (initParams === null) {
            initParams = new ScDict([]);
        }
        const req = new wasmrequests.DeployRequest();
        req.progHash = programHash;
        req.name = name;
        req.description = description;
        req.params = initParams.toBytes();
        sandbox(FnDeployContract, req.bytes());
    }

    // returns random entropy data for current request.
    public entropy(): wasmtypes.ScHash {
        return wasmtypes.hashFromBytes(sandbox(FnEntropy, null));
    }

    // signals an event on the node that external entities can subscribe to
    public event(msg: string): void {
        sandbox(FnEvent, wasmtypes.stringToBytes(msg));
    }

    // access the incoming balances for all assets
    public incomingTransfer(): ScBalances {
        const buf = sandbox(FnIncomingTransfer, null);
        return new ScAssets(buf).balances();
    }

    // retrieve the assets that were minted in this transaction
    public minted(): ScBalances {
        return new ScAssets(sandbox(FnMinted, null)).balances();
    }

    // (delayed) posts a smart contract function request
    public post(chainID: wasmtypes.ScChainID, hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict, transfer: ScTransfers, delay: u32): void {
        if (transfer.balances().colors().length == 0) {
            this.panic("missing transfer");
        }
        const req = new wasmrequests.PostRequest();
        req.chainID = chainID;
        req.contract = hContract;
        req.function = hFunction;
        req.params = params.toBytes();
        req.transfer = transfer.toBytes();
        req.delay = delay;
        sandbox(FnPost, req.bytes());
    }

    // generates a random value from 0 to max (exclusive: max) using a deterministic RNG
    public random(max: u64): u64 {
        if (max == 0) {
            this.panic("random: max parameter should be non-zero");
        }

        // note that entropy gets reset for every request
        if (ScSandboxFunc.entropy.length == 0) {
            // first time in this: request, initialize with current request entropy
            ScSandboxFunc.entropy = this.entropy().toBytes();
            ScSandboxFunc.offset = 0;
        }
        if (ScSandboxFunc.offset == 32) {
            // ran out of entropy: data, hash entropy for next pseudo-random entropy
            ScSandboxFunc.entropy = this.utility().hashBlake2b(ScSandboxFunc.entropy).toBytes();
            ScSandboxFunc.offset = 0;
        }
        let rnd = wasmtypes.uint64FromBytes(ScSandboxFunc.entropy.slice(ScSandboxFunc.offset, ScSandboxFunc.offset + 8)) % max;
        ScSandboxFunc.offset += 8;
        return rnd;
    }

    public rawState(): ScState {
        return new ScState();
    }

    //public request(): ScRequest {
    //	panic("implement me")
    //}

    // retrieve the request id of this transaction
    public requestID(): wasmtypes.ScRequestID {
        return wasmtypes.requestIDFromBytes(sandbox(FnRequestID, null));
    }

    // transfer assets to the specified Tangle ledger address
    public send(address: wasmtypes.ScAddress, transfer: ScTransfers): void {
        // we need some assets to send
        if (transfer.isEmpty()) {
            return;
        }

        const req = new wasmrequests.SendRequest();
        req.address = address;
        req.transfer = transfer.toBytes();
        sandbox(FnSend, req.bytes());
    }

    //public stateAnchor(): interface{} {
    //	panic("implement me")
    //}
}