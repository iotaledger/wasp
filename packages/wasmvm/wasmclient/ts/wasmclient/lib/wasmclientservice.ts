// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';

export interface IClientService {
    callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, isc.Error];

    postRequest(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array, allowance: wasmlib.ScAssets, keyPair: isc.KeyPair, nonce: u64): [wasmlib.ScRequestID, isc.Error];

    subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error;

    waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error;
}

export class WasmClientService implements IClientService {
    waspClient: isc.WaspClient;
    eventPort: string;

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

    public subscribeEvents(msg: /* chan */ string[], done: /* chan */ bool): isc.Error {
        //TODO
        // return subscribe.Subscribe(this.eventPort, msg, done, false, "");
        return null;
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        return this.waspClient.waitUntilRequestProcessed(chainID, reqID, timeout);
    }
}
