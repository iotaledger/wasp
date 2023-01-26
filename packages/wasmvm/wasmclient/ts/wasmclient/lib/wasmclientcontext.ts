// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import * as coreaccounts from 'wasmlib/coreaccounts';
import {WasmClientSandbox} from './wasmclientsandbox';
import {IClientService} from './wasmclientservice';

export class WasmClientContext extends WasmClientSandbox implements wasmlib.ScFuncCallContext {
    private eventHandlers: wasmlib.IEventHandlers[] = [];

    public constructor(svcClient: IClientService, chain: string, scName: string) {
        super(svcClient, chain, scName);
    }

    public currentChainID(): wasmlib.ScChainID {
        return this.chainID;
    }

    public currentKeyPair(): isc.KeyPair | null {
        return this.keyPair;
    }

    public currentSvcClient(): IClientService {
        return this.svcClient;
    }

    public initFuncCallContext(): void {
        wasmlib.connectHost(this);
    }

    public initViewCallContext(_hContract: wasmlib.ScHname): wasmlib.ScHname {
        wasmlib.connectHost(this);
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
        return this.startEventHandlers();
    }

    public serviceContractName(contractName: string) {
        this.scHname = wasmlib.hnameFromBytes(isc.Codec.hNameBytes(contractName));
    }

    public signRequests(keyPair: isc.KeyPair) {
        this.keyPair = keyPair;

        // get last used nonce from accounts core contract
        const agent = wasmlib.ScAgentID.fromAddress(keyPair.address());
        console.log('Chain: ' + this.chainID.toString());
        console.log('Agent: ' + agent.toString());
        const ctx = new WasmClientContext(this.svcClient, this.chainID.toString(), coreaccounts.ScName);
        const n = coreaccounts.ScFuncs.getAccountNonce(ctx);
        n.params.agentID().setValue(agent);
        n.func.call();
        this.nonce = n.results.accountNonce().value();
    }

    public unregister(handler: wasmlib.IEventHandlers): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] === handler) {
                const handlers = this.eventHandlers;
                this.eventHandlers = handlers.slice(0, i).concat(handlers.slice(i + 1));
                if (this.eventHandlers.length == 0) {
                    this.stopEventHandlers();
                }
                return;
            }
        }
    }

    public async waitEvent(): Promise<void> {
        this.Err = null;
        await this.waitEventTimeout(10000);
    }

    private async waitEventTimeout(msec: number): Promise<void> {
        // eslint-disable-next-line @typescript-eslint/no-this-alias
        const self = this;
        return new Promise(function (resolve) {
            setTimeout(function () {
                if (self.eventReceived) {
                    self.eventReceived = false;
                    resolve();
                } else if (msec <= 0) {
                    self.Err = 'event wait timeout';
                    resolve();
                } else {
                    self.waitEventTimeout(msec - 100).then(resolve);
                }
            }, 5);
        });
    }

    public waitRequest(): void {
        this.waitRequestID(this.ReqID);
    }

    public waitRequestID(reqID: wasmlib.ScRequestID): void {
        this.Err = this.svcClient.waitUntilRequestProcessed(this.chainID, reqID, 60);
    }

    private processEvent(msg: string[]): void {
        if (msg[0] == 'error') {
            this.Err = msg[1];
            this.eventReceived = true;
            return;
        }

        if (msg[0] != 'contract' || msg[1] != this.chainID.toString()) {
            // not intended for us
            return;
        }

        const params = msg[6].split('|');
        for (let i = 0; i < params.length; i++) {
            params[i] = this.unescape(params[i]);
        }
        const topic = params[0];
        params.shift();
        for (let i = 0; i < this.eventHandlers.length; i++) {
            this.eventHandlers[i].callHandler(topic, params);
        }

        this.eventReceived = true;
    }

    public startEventHandlers(): isc.Error {
        if (this.eventHandlers.length != 1) {
            return null;
        }
        return this.svcClient.subscribeEvents(this, (msg: string[]) => {
            this.processEvent(msg);
        });
    }

    public stopEventHandlers(): void {
        if (this.eventHandlers.length == 0) {
            this.svcClient.unsubscribeEvents(this);
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
