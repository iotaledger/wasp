// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import * as coreaccounts from 'wasmlib/coreaccounts';
import {WasmClientSandbox} from './wasmclientsandbox';
import {ContractEvent, WasmClientService} from './wasmclientservice';

export class WasmClientContext extends WasmClientSandbox implements wasmlib.ScFuncCallContext {
    private eventHandlers: wasmlib.IEventHandlers[] = [];

    public constructor(svcClient: WasmClientService, scName: string) {
        super(svcClient, scName);
    }

    public currentKeyPair(): isc.KeyPair | null {
        return this.keyPair;
    }

    public currentSvcClient(): WasmClientService {
        return this.svcClient;
    }

    // eslint-disable-next-line @typescript-eslint/no-empty-function
    public initFuncCallContext(): void {
    }

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    public initViewCallContext(_hContract: wasmlib.ScHname): wasmlib.ScHname {
        return this.scHname;
    }

    public register(handler: wasmlib.IEventHandlers): isc.Error {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] === handler) {
                return null;
            }
        }
        this.eventHandlers.push(handler);
        if (this.eventHandlers.length > 1) {
            return null;
        }
        return this.svcClient.subscribeEvents(this, (event) => this.processEvent(event));
    }

    public serviceContractName(contractName: string) {
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(contractName));
    }

    public signRequests(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;

        // TODO not here
        // get last used nonce from accounts core contract
        const agent = wasmlib.ScAgentID.fromAddress(keyPair.address());
        const ctx = new WasmClientContext(this.svcClient, coreaccounts.ScName);
        const n = coreaccounts.ScFuncs.getAccountNonce(ctx);
        n.params.agentID().setValue(agent);
        n.func.call();
        this.Err = ctx.Err;
        if (this.Err == null) {
            this.nonce = n.results.accountNonce().value();
        }
    }

    public unregister(handler: wasmlib.IEventHandlers): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] === handler) {
                const handlers = this.eventHandlers;
                this.eventHandlers = handlers.slice(0, i).concat(handlers.slice(i + 1));
                if (this.eventHandlers.length == 0) {
                    this.svcClient.unsubscribeEvents(this);
                }
                return;
            }
        }
    }

    public waitRequest(): void {
        this.waitRequestID(this.ReqID);
    }

    public waitRequestID(reqID: wasmlib.ScRequestID): void {
        this.Err = this.svcClient.waitUntilRequestProcessed(reqID, 60);
    }

    private processEvent(event: ContractEvent): void {
        const params = event.data.split('|');
        for (let i = 0; i < params.length; i++) {
            params[i] = this.unescape(params[i]);
        }
        const topic = params[0];
        params.shift();
        for (let i = 0; i < this.eventHandlers.length; i++) {
            this.eventHandlers[i].callHandler(topic, params);
        }
    }

    private unescape(param: string): string {
        const i = param.indexOf('~');
        if (i < 0) {
            return param;
        }

        switch (param.charAt(i + 1)) {
            case '~': // escaped escape character
                return param.slice(0, i) + '~' + this.unescape(param.slice(i + 2));
            case '/': // escaped vertical bar
                return param.slice(0, i) + '|' + this.unescape(param.slice(i + 2));
            case '_': // escaped space
                return param.slice(0, i) + ' ' + this.unescape(param.slice(i + 2));
            default:
                panic('invalid event encoding');
        }
        return '';
    }
}
