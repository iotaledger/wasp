// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Base64 } from '@iota/util.js';
import * as wasmlib from 'wasmlib';
import { SyncRequestClient } from 'ts-sync-request';
import {OffLedgerRequest} from "./offledgerrequest";

export type Error = string | null;

// TODO
export class WaspClient {
    baseURL: string;

    public constructor(baseURL: string) {
        if (!baseURL.startsWith('http')) {
            baseURL = 'http://' + baseURL;
        }
        this.baseURL = baseURL;
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, Error] {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/contract/' + hContract.toString() + '/callviewbyhname/' + hFunction.toString();
        const request = Base64.encode(wasmlib.bytesToUint8Array(args));
        const response = new SyncRequestClient();

        const result = response.post(url, request) as Uint8Array;

        return [result, null];
    }

    public postOffLedgerRequest(chainID: wasmlib.ScChainID, signed: OffLedgerRequest): Error {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/request';
        const request = Base64.encode(signed.bytes());
        const response = new SyncRequestClient().post(url, request);
        return null;
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): Error {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/request/' + reqID.toString() + '/wait';
        const response = new SyncRequestClient().get(url);
        return null;
    }
}
