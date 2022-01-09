// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index";
import { Buffer } from "./buffer";
import { IResponse } from "./api_common/response_models";
import * as requestSender from "./api_common/request_sender";
import { Base58, ED25519, Hash, IKeyPair } from "./crypto";

interface ICallViewResponse extends IResponse {
    Items: [{ Key: string; Value: string }];
}

interface IOffLedgerRequest {
    Request: string;
}

export class WaspClient {
    private waspAPI: string;

    constructor(waspAPI: string) {
        if (waspAPI.startsWith("https://") || waspAPI.startsWith("http://")) this.waspAPI = waspAPI;
        else this.waspAPI = "http://" + waspAPI;
    }

    public async callView(chainID: string, contractHName: string, entryPoint: string, args: Buffer): Promise<wasmclient.Results> {
        const request = { Request: args.toString("base64") };
        const result = await requestSender.sendRequestExt<unknown, ICallViewResponse>(
            this.waspAPI,
            "post",
            `/chain/${chainID}/contract/${contractHName}/callview/${entryPoint}`,
            request
        );
        const res = new wasmclient.Results();

        if (result?.body !== null && result.body.Items) {
            for (const item of result.body.Items) {
                const key = Buffer.from(item.Key, "base64");
                const val = Buffer.from(item.Value, "base64");
                const stringKey = key.toString()
                res.res.set(stringKey, val);
                res.keys.set(stringKey, key);
            }
        }
        return res;
    }

    public async postRequest(chainID: string, offLedgerRequest: Buffer): Promise<void> {
        const request = { Request: offLedgerRequest.toString("base64") };
        await requestSender.sendRequestExt<IOffLedgerRequest, null>(this.waspAPI, "post", `/request/${chainID}`, request);
    }

    public async waitRequest(chainID: string, reqID: wasmclient.RequestID): Promise<void> {
        await requestSender.sendRequestExt<unknown, null>(this.waspAPI, "get", `/chain/${chainID}/request/${reqID}/wait`);
    }

    public async postOffLedgerRequest(
        chainId: string,
        scHName: wasmclient.Hname,
        hFuncName: wasmclient.Int32,
        args: wasmclient.Arguments,
        transfer: wasmclient.Transfer,
        keyPair: IKeyPair
    ): Promise<string> {
        // get request essence ready for signing
        let essence = Base58.decode(chainId);
        const hNames = Buffer.alloc(8);
        hNames.writeUInt32LE(scHName, 0);
        hNames.writeUInt32LE(hFuncName, 4);
        const nonce = Buffer.alloc(8);
        nonce.writeBigUInt64LE(BigInt(Math.trunc(performance.now())), 0);
        essence = Buffer.concat([essence, hNames, args.encode(), keyPair.publicKey, nonce, transfer.encode()]);

        let buf = Buffer.alloc(1);
        const requestTypeOffledger = 1;
        buf.writeUInt8(requestTypeOffledger, 0);
        buf = Buffer.concat([buf, essence, ED25519.privateSign(keyPair, essence)]);
        const hash = Hash.from(buf);
        const requestID = Buffer.concat([hash, Buffer.alloc(2)]);
        await this.postRequest(chainId, buf);

        return Base58.encode(requestID);
    }
}
