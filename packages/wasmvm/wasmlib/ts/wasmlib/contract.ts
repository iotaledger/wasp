// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmrequests from "./wasmrequests"
import * as wasmtypes from "./wasmtypes"
import {ScTransfer} from "./assets";
import {ScDict} from "./dict";
import {sandbox} from "./host";
import {FnCall, FnPost, panic, ScSandbox} from "./sandbox";

// base contract objects

export interface ScViewCallContext {
    chainID(): wasmtypes.ScChainID;
    initViewCallContext(hContract: wasmtypes.ScHname): wasmtypes.ScHname;
}

export interface ScFuncCallContext extends ScViewCallContext {
    initFuncCallContext(): void;
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

    constructor(ctx: ScViewCallContext, hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname) {
        this.hContract = hContract;
        this.hFunction = hFunction;
        this.params = ScView.nilParams;
        this.resultsProxy = null;
    }

    call(): void {
        this.callWithAllowance(null);
    }

    protected callWithAllowance(allowance: ScTransfer | null): void {
        const req = new wasmrequests.CallRequest();
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        if (allowance === null) {
            allowance = new ScTransfer();
        }
        req.allowance = allowance.toBytes();
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
    constructor(ctx: ScFuncCallContext, hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname) {
        super(ctx, hContract, hFunction);
    }

    call(): void {
        return panic("cannot call init");
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScFunc extends ScView {
    delaySeconds: u32 = 0;
    allowanceAssets: ScTransfer | null = null;
    transferAssets: ScTransfer | null = null;

    constructor(ctx: ScFuncCallContext, hContract: wasmtypes.ScHname, hFunction: wasmtypes.ScHname) {
        super(ctx, hContract, hFunction);
    }

    allowance(allowance: ScTransfer): ScFunc {
        this.allowanceAssets = allowance;
        return this;
    }

    allowanceBaseTokens(amount: i64): ScFunc {
        return this.allowance(ScTransfer.baseTokens(amount));
    }

    call(): void {
        if (this.transferAssets != null) {
            panic("cannot transfer assets in a call");
            return
        }
        if (this.delaySeconds != 0) {
            panic("cannot delay a call");
            return
        }
        this.callWithAllowance(this.allowanceAssets);
    }

    delay(seconds: u32): ScFunc {
        this.delaySeconds = seconds;
        return this;
    }

    post(): void {
        return this.postToChain(new ScSandbox().currentChainID());
    }

    postToChain(chainID: wasmtypes.ScChainID): void {
        const req = new wasmrequests.PostRequest();
        req.chainID = chainID;
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        let allowance = this.allowanceAssets;
        if (allowance === null) {
            allowance = new ScTransfer();
        }
        req.allowance = allowance.toBytes();
        let transfer = this.transferAssets;
        if (transfer === null) {
            transfer = new ScTransfer();
        }
        req.transfer = transfer.toBytes();
        req.delay = this.delaySeconds;
        const res = sandbox(FnPost, req.bytes());
        if (this.resultsProxy) {
            this.resultsProxy = new wasmtypes.Proxy(new ScDict(res));
        }
    }

    transfer(transfer: ScTransfer): ScFunc {
        this.transferAssets = transfer;
        return this;
    }

    transferBaseTokens(amount: i64): ScFunc {
        return this.transfer(ScTransfer.baseTokens(amount));
    }
}
