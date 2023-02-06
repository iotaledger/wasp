// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Base64} from '@iota/util.js';
import * as wasmlib from 'wasmlib';
import {SyncRequestClient} from './ts-sync-request';
import {OffLedgerRequest} from './offledgerrequest';
import {APICallViewRequest, APIOffLedgerRequest, Codec, JsonReq, JsonResp} from './codec';
import { encode, decode } from 'as-hex';

export type Error = string | null;

export class WaspClient {
    baseURL: string;

    public constructor(baseURL: string) {
        if (!baseURL.startsWith('http')) {
            baseURL = 'http://' + baseURL;
        }
        this.baseURL = baseURL;
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: Uint8Array): [Uint8Array, Error] {
        const url = this.baseURL + '/requests/callview';
        const req = new SyncRequestClient();
        req.addHeader('Content-Type', 'application/json');

        const callViewRequest: APICallViewRequest = {
            contractHName: hContract.toString(),
            functionHName: hFunction.toString(),
            chainId: chainID.toString(),
            arguments: Codec.jsonEncode(args),
        };

        try {
            const resp = req.post<APICallViewRequest, JsonResp>(url, callViewRequest);
            const result = Codec.jsonDecode(resp);
            return [result, null];
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return [new Uint8Array(0), message];
        }
    }



    public postOffLedgerRequest(chainID: wasmlib.ScChainID, signed: OffLedgerRequest): Error {
        const url = this.baseURL + '/chain/' + chainID.toString() + '/request';
        const req = new SyncRequestClient();
        req.addHeader('Content-Type', 'application/json');

        const offLedgerRequest: APIOffLedgerRequest = {
            chainId: chainID.toString(),
            // Validate if this is actually valid to do. This byte array needs to be sent as hex.
            request: encode(signed.bytes().toString()),
        };

        try {
            req.post(url, offLedgerRequest);
            return null;
        } catch (error) {
            let message;
            if (error instanceof Error) message = error.message;
            else message = String(error);
            return message;
        }
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): Error {
        // Timeout of the wait can be set with `/wait?timeoutSeconds=`. Max seconds are 60secs.
        const url = this.baseURL + '/chains/' + chainID.toString() + '/requests/' + reqID.toString() + '/wait';
        const response = new SyncRequestClient().get(url);
        return null;
    }
}
