// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Buffer} from "./buffer";

const headers: { [id: string]: string } = {
    "Content-Type": "application/json",
};

interface IResponse {
    error?: string;
}

interface IExtendedResponse<U> {
    body: U;
    response: Response;
}

interface ICallViewResponse extends IResponse {
    Items: [{ Key: string; Value: string }];
}

interface IOffLedgerRequest {
    Request: string;
}

export class WaspClient {
    private waspAPI: string;

    constructor(waspAPI: string) {
        if(waspAPI.startsWith("https://") || waspAPI.startsWith("http://"))
            this.waspAPI = waspAPI;
        else
            this.waspAPI = "http://" + waspAPI;
    }

    public async callView(chainID: string, contractHName: string, entryPoint: string, args: Buffer): Promise<wasmclient.Results> {
        const request = {Request: args.toString("base64")};
        const result = await this.sendRequest<unknown, ICallViewResponse>(
            "get",
            "/chain/" + chainID + "/contract/ " + contractHName + "/callview/" + entryPoint,
            request
        );
        const res = new wasmclient.Results();
        if (result.body.Items) {
            for (const item of result.body.Items) {
                const key = Buffer.from(item.Key, "base64").toString();
                const value = Buffer.from(item.Value, "base64");
                res.res.set(key, value);
            }
        }
        return res;
    }

    public async postRequest(chainID: string, offLedgerRequest: Buffer): Promise<void> {
        const request = {Request: offLedgerRequest.toString("base64")};
        await this.sendRequest<IOffLedgerRequest, null>(
            "post",
            "/request/" + chainID,
            request,
        );
    }

    public async waitRequest(chainID: string, reqID: wasmclient.RequestID): Promise<void> {
        await this.sendRequest<unknown, null>(
            "get",
            "/chain/" + chainID + "/request/" + reqID + "/wait",
        );
    }

    private async sendRequest<T, U extends IResponse>(
        verb: "put" | "post" | "get" | "delete",
        path: string,
        request?: T | undefined,
    ): Promise<IExtendedResponse<U>> {
        let response: U;
        let fetchResponse: Response;

        try {
            const url = this.waspAPI + path;
            fetchResponse = await fetch(url, {
                method: verb,
                headers,
                body: JSON.stringify(request),
            });

            if (!fetchResponse) {
                throw new Error("No data was returned from the API");
            }

            try {
                response = await fetchResponse.json();
            } catch (err) {
                if (!fetchResponse.ok) {
                    const text = await fetchResponse.text();
                    throw new Error(err.message + "   ---   " + text);
                }
            }
        } catch (err) {
            throw new Error("sendRequest: " + err.message);
        }

        return {body: response, response: fetchResponse};
    }
}
