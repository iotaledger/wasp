// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';

export interface IClientService {
    callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): Uint8Array;

    Err(): isc.Error;

    postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair, nonce: u64): wasmlib.ScRequestID;

    subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error;

    waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error;
}

export class WasmClientService implements IClientService {
    waspClient: isc.WaspClient;
    lastError: isc.Error = null;
    eventPort: string;

    public constructor(waspAPI: string, eventPort: string) {
        this.waspClient = new isc.WaspClient(waspAPI);
        this.eventPort = eventPort;
    }

    public static DefaultWasmClientService(): WasmClientService {
        return new WasmClientService('127.0.0.1:9090', '127.0.0.1:5550');
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): Uint8Array {
        const res = this.waspClient.callViewByHname(chainID, hContract, hFunction, args);
        this.lastError = this.waspClient.Err;
        if (this.lastError != null) {
            return new Uint8Array(0);
        }
        return res;
    }

    public Err(): isc.Error {
        return this.lastError;
    }

    public postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair, nonce: u64): wasmlib.ScRequestID {
        const req = new isc.OffLedgerRequest(chainID, hContract, hFunction, args, nonce);
        req.withAllowance(allowance);
        const signed = req.sign(keyPair);
        this.lastError = this.waspClient.postOffLedgerRequest(chainID, signed);
        if (this.lastError != null) {
            return wasmlib.requestIDFromBytes(new Uint8Array(0));
        }
        return signed.ID();
    }

    public subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error {
        //TODO
        // return subscribe.Subscribe(this.eventPort, msg, done, false, "");
        return null;
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        return this.waspClient.waitUntilRequestProcessed(chainID, reqID, timeout);
    }
}
