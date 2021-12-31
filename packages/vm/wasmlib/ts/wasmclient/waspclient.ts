// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Buffer} from "./buffer";

export class WaspClient {
    private waspAPI: string;

    constructor(waspAPI: string) {
        this.waspAPI = "http://" + waspAPI;
    }

    public async sendOffLedgerRequest(chainId: string, offLedgerRequest: Buffer): Promise<void> {
        const request = { Request: offLedgerRequest.toString('base64') };
        await this.sendRequestExt<IOffLedgerRequest, null>(
            this.waspAPI,
            'post',
            `request/${chainId}`,
            request,
        );
    }

    public async sendExecutionRequest(chainId: string, offLedgerRequestId: string): Promise<void> {
        await this.sendRequestExt<IOffLedgerRequest, null>(
            this.waspAPI,
            'get',
            `chain/${chainId}/request/${offLedgerRequestId}/wait`,
        );
    }

    public async callView(chainId: string, contractHName: string, entryPoint: string, args: wasmclient.Arguments): Promise<CallViewResponse> {
        const url = `chain/${chainId}/contract/${contractHName}/callview/${entryPoint}`;

        const result = await this.sendRequestExt<unknown, CallViewResponse>(this.waspAPI, 'get', url);

        return result.body;
    }

    private async sendRequestExt<T, U extends IResponse>(
        url: string,
        verb: 'put' | 'post' | 'get' | 'delete',
        path: string,
        request?: T | undefined,
    ): Promise<IExtendedResponse<U>> {
        let response: U;
        let fetchResponse: Response;

        try {
            const headers: { [id: string]: string } = {
                'Content-Type': 'application/json',
            };

            if (verb == 'get' || verb == 'delete') {
                fetchResponse = await fetch(`${url}/${path}`, {
                    method: verb,
                    headers,
                });
            } else if (verb == 'post' || verb == 'put') {
                fetchResponse = await fetch(`${url}/${path}`, {
                    method: verb,
                    headers,
                    body: JSON.stringify(request),
                });
            }

            if (!fetchResponse) {
                throw new Error('No data was returned from the API');
            }

            try {
                response = await fetchResponse.json();
            } catch (err) {
                if (!fetchResponse.ok) {
                    const text = await fetchResponse.text();
                    throw new Error(err.message + '   ---   ' + text);
                }
            }
        } catch (err) {
            throw new Error(
                `The application is not able to complete the request, due to the following error:\n\n${err.message}`,
            );
        }

        return { body: response, response: fetchResponse };
    }
}
