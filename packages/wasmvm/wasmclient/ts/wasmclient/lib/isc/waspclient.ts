// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Base64} from '@iota/util.js';
import * as wasmlib from 'wasmlib';
import {SyncRequestClient} from './ts-sync-request';
import {OffLedgerRequest} from "./offledgerrequest";
import {Codec, JsonReq, JsonResp} from "./codec";

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
        const req = new SyncRequestClient();
        req.addHeader("Content-Type", "application/json")
        const body = Codec.jsonEncode(args);
        try {
            const resp = req.post<JsonReq, JsonResp>(url, body);
            const result = Codec.jsonDecode(resp);
            return [result, null];
        } catch (error) {
            let message
            if (error instanceof Error) message = error.message
            else message = String(error)
            return [new Uint8Array(0), message];
        }
    }

    public postOffLedgerRequest(chainID: wasmlib.ScChainID, signed: OffLedgerRequest): Error {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/request';
        const req = new SyncRequestClient();
        req.addHeader("Content-Type", "application/json")
        const body = { Request: Base64.encode(signed.bytes()) };
        try {
            req.post(url, body);
            return null;
        } catch (error) {
            let message
            if (error instanceof Error) message = error.message
            else message = String(error)
            return message;
        }
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): Error {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/request/' + reqID.toString() + '/wait';
        const response = new SyncRequestClient().get(url);
        return null;
    }
}
