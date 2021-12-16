// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasp from "../../../../../../contracts/wasm/fairroulette/frontend/src/lib/wasp_client";
import {Colors, IOffLedger, OffLedger} from "../../../../../../contracts/wasm/fairroulette/frontend/src/lib/wasp_client";
import config from './config.dev';
import * as client from "./index"

export type ServiceClient = wasp.BasicClient;

export type EventHandlers = { [key: string]: (message: string[]) => void };

export class FuncObject {
    svc: Service;

    constructor(svc: Service) {
        this.svc = svc;
    }
}

export class ViewResults {
    res: client.Results;

    constructor(res: client.Results) {
        this.res = res;
    }
}

export class Service {
    private client: ServiceClient;
    private walletService: wasp.WalletService;
    private webSocket: WebSocket;
    private eventHandlers: EventHandlers;
    public chainId: string;
    public scHname: client.Hname;

    constructor(client: ServiceClient, chainId: string, scHname: client.Hname, eventHandlers: EventHandlers) {
        this.client = client;
        this.chainId = chainId;
        this.scHname = scHname;
        this.eventHandlers = eventHandlers;
        this.walletService = new wasp.WalletService(client);
        this.connectWebSocket();
    }

    private connectWebSocket(): void {
        const webSocketUrl = config.waspWebSocketUrl.replace('%chainId', this.chainId);
        // eslint-disable-next-line no-console
        console.log(`Connecting to Websocket => ${webSocketUrl}`);
        this.webSocket = new WebSocket(webSocketUrl);
        this.webSocket.addEventListener('message', (x) => this.handleIncomingMessage(x));
        this.webSocket.addEventListener('close', () => setTimeout(this.connectWebSocket.bind(this), 1000));
    }

    private handleIncomingMessage(message: MessageEvent<string>): void {
        // expect vmmsg <chain ID> <contract hname> contract.event|parameters
        const msg = message.data.toString().split(' ');
        if (msg.length != 4 || msg[0] != 'vmmsg') {
            return;
        }
        const topics = msg[3].split('|');
        const topic = topics[0];
        if (this.eventHandlers[topic] != undefined) {
            this.eventHandlers[topic](msg.slice(1));
        }
    }

    // calls a view
    public async callView(viewName: string, args: client.Arguments): Promise<client.Results> {
        const response = await this.client.callView(
            this.chainId,
            this.scHname.toString(16),
            viewName,
            args,
        );

        const resultMap: ParameterResult = {};
        if (response.Items) {
            for (let item of response.Items) {
                const key = wasp.Buffer.from(item.Key, "base64").toString();
                const value = wasp.Buffer.from(item.Value, "base64");
                resultMap[key] = value;
            }
        }
        return resultMap;
    }

    // posts off-tangle request
    public async postRequest(funcName: string, args: client.Arguments): Promise<void> {
        let request: IOffLedger = {
            requestType: 1,
            noonce: BigInt(performance.now() + performance.timeOrigin * 10000000),
            contract: this.scHname,
            entrypoint: hFunc,
            arguments: [{key: '-number', value: betNumber}],
            balances: [{balance: take, color: Colors.IOTA_COLOR_BYTES}],
        };

        request = OffLedger.Sign(request, keyPair);

        await this.client.sendOffLedgerRequest(this.chainId, request);
        await this.client.sendExecutionRequest(this.chainId, OffLedger.GetRequestId(request));
    }
}