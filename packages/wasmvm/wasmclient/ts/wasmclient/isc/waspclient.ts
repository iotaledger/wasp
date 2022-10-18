// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Base64} from "@iota/util.js"
import * as isc from "./index"
import * as wasmlib from "wasmlib"
import {SyncRequestClient} from "ts-sync-request"

export type Error = string | null;

// TODO
export class WaspClient {
    baseURL: string;
    Err: isc.Error = null;

    public constructor(baseURL: string) {
        if (!baseURL.startsWith("http")) {
            baseURL = "http://" + baseURL;
        }
        this.baseURL = baseURL;
    }

    public callViewByHname(chainID: wasmlib.ScChainID, hContract: wasmlib.ScHname, hFunction: wasmlib.ScHname, args: u8[]): u8[] {
        this.Err = null;
        const url = "/chain/" + chainID.toString() + "/contract/" + hContract.toString() + "/callviewbyhname/" + hFunction.toString();
        const request = Base64.encode(wasmlib.bytesToUint8Array(args));
        const response = new SyncRequestClient().post(url, request);
        return [];
    }

    public postOffLedgerRequest(chainID: wasmlib.ScChainID, signed: isc.OffLedgerRequest): isc.Error {
        this.Err = null;
        const url = "/chain/" + chainID.toString() + "/request";
        const request = Base64.encode(wasmlib.bytesToUint8Array(signed.bytes()));
        const response = new SyncRequestClient().post(url, request);
        return null;
    }

    public waitUntilRequestProcessed(chainID: wasmlib.ScChainID, reqID: wasmlib.ScRequestID, timeout: u32): isc.Error {
        const url = "/chain/" + chainID.toString() + "/request/" + reqID.toString() + "/wait";
        const response = new SyncRequestClient().get(url);
        return null;
    }
}
