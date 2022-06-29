// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmrequests from "./wasmrequests"
import * as wasmtypes from "./wasmtypes"
import {ScAssets, ScBalances, ScTransfer} from "./assets";
import {ScDict, ScImmutableDict} from "./dict";
import {sandbox} from "./host";
import {ScSandboxUtils} from "./sandboxutils";
import {ScImmutableState, ScState} from "./state";
import {ScFunc} from "./contract";

// @formatter:off
export const FnAccountID           : i32 = -1;
export const FnAllowance           : i32 = -2;
export const FnBalance             : i32 = -3;
export const FnBalances            : i32 = -4;
export const FnBlockContext        : i32 = -5;
export const FnCall                : i32 = -6;
export const FnCaller              : i32 = -7;
export const FnChainID             : i32 = -8;
export const FnChainOwnerID        : i32 = -9;
export const FnContract            : i32 = -10;
export const FnDeployContract      : i32 = -12;
export const FnEntropy             : i32 = -13;
export const FnEstimateDust        : i32 = -14;
export const FnEvent               : i32 = -15;
export const FnLog                 : i32 = -16;
export const FnMinted              : i32 = -17;
export const FnPanic               : i32 = -18;
export const FnParams              : i32 = -19;
export const FnPost                : i32 = -20;
export const FnRequest             : i32 = -21;
export const FnRequestID           : i32 = -22;
export const FnResults             : i32 = -23;
export const FnSend                : i32 = -24;
export const FnStateAnchor         : i32 = -25;
export const FnTimestamp           : i32 = -26;
export const FnTrace               : i32 = -27;
export const FnTransferAllowed     : i32 = -28;
export const FnUtilsBase58Decode   : i32 = -29;
export const FnUtilsBase58Encode   : i32 = -30;
export const FnUtilsBech32Decode   : i32 = -31;
export const FnUtilsBech32Encode   : i32 = -32;
export const FnUtilsBlsAddress     : i32 = -33;
export const FnUtilsBlsAggregate   : i32 = -34;
export const FnUtilsBlsValid       : i32 = -35;
export const FnUtilsEd25519Address : i32 = -36;
export const FnUtilsEd25519Valid   : i32 = -37;
export const FnUtilsHashBlake2b    : i32 = -38;
export const FnUtilsHashName       : i32 = -39;
export const FnUtilsHashSha3       : i32 = -40;
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

    public balance(tokenID: wasmtypes.ScTokenID): u64 {
        return wasmtypes.uint64FromBytes(sandbox(FnBalance, tokenID.toBytes()));
    }

    // access the current balances for all assets
    public balances(): ScBalances {
        return new ScAssets(sandbox(FnBalances, null)).balances();
    }

    // calls a smart contract function
    protected callWithAllowance(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict | null, allowance: ScTransfer | null): ScImmutableDict {
        const req = new wasmrequests.CallRequest();
        req.contract = hContract;
        req.function = hFunction;
        if (params === null) {
            params = new ScDict([]);
        }
        req.params = params.toBytes();
        if (allowance === null) {
            allowance = new ScTransfer();
        }
        req.allowance = allowance.toBytes();
        const res = sandbox(FnCall, req.bytes());
        return new ScDict(res).immutable();
    }

    // retrieve the agent id of the owner of the chain this contract lives on
    public chainOwnerID(): wasmtypes.ScAgentID {
        return wasmtypes.agentIDFromBytes(sandbox(FnChainOwnerID, null));
    }

    // retrieve the hname of this contract
    public contract(): wasmtypes.ScHname {
        return wasmtypes.hnameFromBytes(sandbox(FnContract, null));
    }

    // retrieve the chain id of the chain this contract lives on
    public currentChainID(): wasmtypes.ScChainID {
        return wasmtypes.chainIDFromBytes(sandbox(FnChainID, null));
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
        return this.callWithAllowance(hContract, hFunction, params, null);
    }

    public rawState(): ScImmutableState {
        return new ScImmutableState();
    }
}

export class ScSandboxFunc extends ScSandbox {
    private static entropy: u8[] = [];
    private static offset: u32 = 0;

    // access the allowance assets
    public allowance(): ScBalances {
        const buf = sandbox(FnAllowance, null);
        return new ScAssets(buf).balances();
    }

    //public blockContext(construct func(sandbox: ScSandbox) interface{}, onClose func(interface{})): interface{} {
    //	panic("implement me")
    //}

    // calls a smart contract function
    public call(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict | null, allowance: ScTransfer | null): ScImmutableDict {
        return this.callWithAllowance(hContract, hFunction, params, allowance);
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

    public estimateDust(fn: ScFunc): u64 {
        const req = new wasmrequests.PostRequest();
        req.contract = fn.hContract;
        req.function = fn.hFunction;
        req.params = fn.params.toBytes();
        let allowance = fn.allowanceAssets;
        if (allowance === null) {
            allowance = new ScTransfer();
        }
        req.allowance = allowance.toBytes();
        let transfer = fn.transferAssets;
        if (transfer === null) {
            transfer = new ScTransfer();
        }
        req.transfer = transfer.toBytes();
        req.delay = fn.delaySeconds;
        return wasmtypes.uint64FromBytes(sandbox(FnEstimateDust, req.bytes()));
    }

    // signals an event on the node that external entities can subscribe to
    public event(msg: string): void {
        sandbox(FnEvent, wasmtypes.stringToBytes(msg));
    }

    // retrieve the assets that were minted in this transaction
    public minted(): ScBalances {
        return new ScAssets(sandbox(FnMinted, null)).balances();
    }

    // Post (delayed) posts a SC function request
    public post(chainID: wasmtypes.ScChainID, hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname, params: ScDict, allowance: ScTransfer, transfer: ScTransfer, delay: u32): void {
        const req = new wasmrequests.PostRequest();
        req.chainID = chainID;
        req.contract = hContract;
        req.function = hFunction;
        req.params = params.toBytes();
        req.allowance = allowance.toBytes();
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

    // Send transfers SC assets to the specified address
    public send(address: wasmtypes.ScAddress, transfer: ScTransfer): void {
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

    // TransferAllowed transfers allowed assets from caller to the specified account
    public transferAllowed(agentID: wasmtypes.ScAgentID, transfer: ScTransfer, create: bool): void {
        // we need some assets to send
        if (transfer.isEmpty()) {
            return;
        }

        const req = new wasmrequests.TransferRequest();
        req.agentID = agentID;
        req.create = create;
        req.transfer = transfer.toBytes();
        sandbox(FnTransferAllowed, req.bytes());
    }
}
