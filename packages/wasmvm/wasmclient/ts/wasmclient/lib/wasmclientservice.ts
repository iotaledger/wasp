// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import nano, {Socket} from 'nanomsg';

export interface IClientService {
    callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, isc.Error];

    postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair, nonce: u64): [wasmlib.ScRequestID, isc.Error];

    subscribeEvents(who: any, callback: (msg: string[]) => void): isc.Error;

    unsubscribeEvents(who: any): void;

    waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error;
}

type ClientCallBack = (msg: string[]) => void;

export class WasmClientService implements IClientService {
    private callbacks: ClientCallBack[] = [];
    private eventPort: string;
    private eventListener: Socket = nano.socket('sub');
    private subscribers: any[] = [];
    private waspClient: isc.WaspClient;

    public constructor(waspAPI: string, eventPort: string) {
        this.waspClient = new isc.WaspClient(waspAPI);
        this.eventPort = eventPort;
    }

    public static DefaultWasmClientService(): WasmClientService {
        return new WasmClientService('127.0.0.1:19090', '127.0.0.1:15550');
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, isc.Error] {
        return this.waspClient.callViewByHname(chainID, hContract, hFunction, args);
    }

    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair, nonce: u64): [wasmlib.ScRequestID, isc.Error] {
        const req = new isc.OffLedgerRequest(chainID, hContract, hFunction, args, nonce);
        req.withAllowance(allowance);
        const signed = req.sign(keyPair);
        const reqID = signed.ID();
        const err = this.waspClient.postOffLedgerRequest(chainID, signed);
        return [reqID, err];
    }

    public subscribeEvents(who: any, callback: (msg: string[]) => void): isc.Error {
        // eslint-disable-next-line @typescript-eslint/no-this-alias
        const self = this;
        this.callbacks.push(callback);
        this.subscribers.push(who);
        if (this.subscribers.length == 1) {
            this.eventListener.on('error', function (err: any) {
                callback(['error', err.toString()]);
            });
            this.eventListener.on('data', function (buf: any) {
                const txt = buf.toString();
                const msg = txt.split(' ');
                if (msg[0] == 'contract') {
                    for (let i = 0; i < self.callbacks.length; i++) {
                        self.callbacks[i](msg);
                    }
                }
            });
            this.eventListener.connect('tcp://' + this.eventPort);
        }
        return null;
    }

    public unsubscribeEvents(who: any): void {
        for (let i = 0; i < this.subscribers.length; i++) {
            if (this.subscribers[i] === who) {
                this.subscribers.splice(i, 1);
                this.callbacks.splice(i, 1);
                if (this.subscribers.length == 0) {
                    this.eventListener.close();
                }
                return;
            }
        }
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        return this.waspClient.waitUntilRequestProcessed(chainID, reqID, timeout);
    }
}
