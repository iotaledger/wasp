// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmrequests from "./wasmrequests"
import * as wasmtypes from "./wasmtypes"
import {ScAssets, ScTransfers} from "./assets";
import {ScDict} from "./dict";
import {sandbox} from "./host";
import {FnCall, FnPost, panic, ScSandbox} from "./sandbox";

// base contract objects

export interface ScFuncCallContext {
    canCallFunc(): void;
}

export interface ScViewCallContext {
    canCallView(): void;
}

export function newCallParamsProxy(v: ScView): wasmtypes.Proxy {
    v.params = new ScDict([]);
    return v.params.asProxy();
}

export function newCallResultsProxy(v: ScView): wasmtypes.Proxy {
    const proxy = new ScDict([]).asProxy();
    v.resultsProxy = proxy;
    return proxy
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScView {
    private static nilParams: ScDict = new ScDict([]);
    public static nilProxy: wasmtypes.Proxy = new wasmtypes.Proxy(ScView.nilParams);

    hContract: wasmtypes.ScHname;
    hFunction: wasmtypes.ScHname;
    params: ScDict;
    resultsProxy: wasmtypes.Proxy | null;

    constructor(hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname) {
        this.hContract = hContract;
        this.hFunction = hFunction;
        this.params = ScView.nilParams;
        this.resultsProxy = null;
    }

    call(): void {
        this.callWithTransfer(null);
    }

    protected callWithTransfer(transfer: ScAssets | null): void {
        //TODO new ScSandboxFunc().call(...)
        if (transfer === null) {
            transfer = new ScAssets([]);
        }
        const req = new wasmrequests.CallRequest();
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        req.transfer = transfer.toBytes();
        const res = sandbox(FnCall, req.bytes());
        const proxy = this.resultsProxy;
        if (proxy != null) {
            proxy.kvStore = new ScDict(res);
        }
    }

    ofContract(hContract: wasmtypes.ScHname): ScView {
        this.hContract = hContract;
        return this;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScInitFunc extends ScView {
    call(): void {
        return panic("cannot call init");
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScFunc extends ScView {
    delaySeconds: u32 = 0;
    transfers: ScAssets | null = null;

    call(): void {
        if (this.delaySeconds != 0) {
            return panic("cannot delay a call");
        }
        this.callWithTransfer(this.transfers);
    }

    delay(seconds: u32): ScFunc {
        this.delaySeconds = seconds;
        return this;
    }

    post(): void {
        return this.postToChain(new ScSandbox().chainID());
    }

    postToChain(chainID: wasmtypes.ScChainID): void {
        let transfer = this.transfers;
        if (transfer === null) {
            transfer = new ScAssets([]);
        }
        const req = new wasmrequests.PostRequest();
        req.chainID = chainID;
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        req.transfer = transfer.toBytes();
        req.delay = this.delaySeconds;
        const res = sandbox(FnPost, req.bytes());
        if (this.resultsProxy) {
            //TODO set kvStore directly?
            this.resultsProxy = new wasmtypes.Proxy(new ScDict(res));
        }
    }

    transfer(transfers: ScTransfers): ScFunc {
        this.transfers = transfers;
        return this;
    }

    transferIotas(amount: i64): ScFunc {
        return this.transfer(ScTransfers.iotas(amount));
    }
}
