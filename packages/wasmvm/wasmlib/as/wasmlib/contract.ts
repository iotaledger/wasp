// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScTransfer} from './assets';
import {ScDict} from './dict';
import {panic} from './sandbox';
import {CallRequest, PostRequest} from './wasmrequests';
import {ScChainID} from './wasmtypes/scchainid';
import {Proxy} from './wasmtypes/proxy';
import {ScHname} from './wasmtypes/schname';

// base contract objects

export interface ScViewClientContext {
    clientContract(hContract: ScHname): ScHname;

    fnCall(req: CallRequest): Uint8Array;

    fnChainID(): ScChainID;
}

export interface ScFuncClientContext extends ScViewClientContext {
    fnPost(req: PostRequest): Uint8Array;
}

export function newCallParamsProxy(v: ScView): Proxy {
    v.params = new ScDict(null);
    return v.params.asProxy();
}

export function newCallResultsProxy(v: ScView): Proxy {
    const proxy = new ScDict(null).asProxy();
    v.resultsProxy = proxy;
    return proxy;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScView {
    private static nilParams: ScDict = new ScDict(null);
    public static nilProxy: Proxy = new Proxy(ScView.nilParams);

    ctx: ScViewClientContext;
    hContract: ScHname;
    hFunction: ScHname;
    params: ScDict;
    resultsProxy: Proxy | null;

    constructor(ctx: ScViewClientContext, hContract: ScHname, hFunction: ScHname) {
        this.ctx = ctx;
        this.hContract = ctx.clientContract(hContract);
        this.hFunction = hFunction;
        this.params = ScView.nilParams;
        this.resultsProxy = null;
    }

    call(): void {
        this.callWithAllowance(null);
    }

    ofContract(hContract: ScHname): ScView {
        this.hContract = hContract;
        return this;
    }

    protected callWithAllowance(allowance: ScTransfer | null): void {
        const req = new CallRequest();
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        req.allowance = (allowance === null) ? new Uint8Array(0): allowance.toBytes();
        const res = this.ctx.fnCall(req);
        const proxy = this.resultsProxy;
        if (proxy != null) {
            proxy.kvStore = new ScDict(res);
        }
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScInitFunc extends ScView {
    constructor(ctx: ScFuncClientContext, hContract: ScHname, hFunction: ScHname) {
        super(ctx, hContract, hFunction);
    }

    call(): void {
        return panic('cannot call init');
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScFunc extends ScView {
    allowanceAssets: ScTransfer | null = null;
    delaySeconds: u32 = 0;
    fctx: ScFuncClientContext;
    transferAssets: ScTransfer | null = null;

    constructor(ctx: ScFuncClientContext, hContract: ScHname, hFunction: ScHname) {
        super(ctx, hContract, hFunction);
        this.fctx = ctx;
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
            panic('cannot transfer assets in a call');
            return;
        }
        if (this.delaySeconds != 0) {
            panic('cannot delay a call');
            return;
        }
        this.callWithAllowance(this.allowanceAssets);
    }

    delay(seconds: u32): ScFunc {
        this.delaySeconds = seconds;
        return this;
    }

    post(): void {
        return this.postToChain(this.ctx.fnChainID());
    }

    postToChain(chainID: ScChainID): void {
        const req = new PostRequest();
        req.chainID = chainID;
        req.contract = this.hContract;
        req.function = this.hFunction;
        req.params = this.params.toBytes();
        const allowance = this.allowanceAssets;
        if (allowance !== null) {
            req.allowance = allowance.toBytes();
        }
        const transfer = this.transferAssets;
        if (transfer !== null) {
            req.transfer = transfer.toBytes();
        }
        req.delay = this.delaySeconds;
        const res = this.fctx.fnPost(req);
        if (this.resultsProxy) {
            this.resultsProxy = new Proxy(new ScDict(res));
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
