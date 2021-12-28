// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import * as wasp from "../../../../../contracts/wasm/fairroulette/frontend/src/lib/wasp_client";
import config from './config.dev';
import {Base58, ED25519, IKeyPair} from "./crypto";
import {Buffer} from "./buffer";

export type EventHandlers = { [key: string]: (message: string[]) => void };

export class FuncObject {
    svc: Service;

    constructor(svc: Service) {
        this.svc = svc;
    }
}

export class ViewResults {
    res: wasmclient.Results;

    constructor(res: wasmclient.Results) {
        this.res = res;
    }
}

export class Service {
    private client: wasmclient.ServiceClient;
    private walletService: wasp.WalletService;
    private webSocket: WebSocket;
    private eventHandlers: EventHandlers;
    private keyPair: IKeyPair;
    public chainId: string;
    public scHname: wasmclient.Hname;

    constructor(client: ServiceClient, chainId: string, scHname: wasmclient.Hname, eventHandlers: EventHandlers) {
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
    public async callView(viewName: string, args: wasmclient.Arguments): Promise<wasmclient.Results> {
        const response = await this.client.callView(
            this.chainId,
            this.scHname.toString(16),
            viewName,
            args,
        );

        const resultMap: ParameterResult = {};
        if (response.Items) {
            for (let item of response.Items) {
                const key = Buffer.from(item.Key, "base64").toString();
                const value = Buffer.from(item.Value, "base64");
                resultMap[key] = value;
            }
        }
        return resultMap;
    }

    // posts off-tangle request
    public async postRequest(hFuncName: wasmclient.Int32, args: wasmclient.Arguments, transfer?: wasmclient.Transfer): Promise<void> {
        if (!transfer) {
            transfer = new wasmclient.Transfer();
        }

        // get request essence ready for signing
        let essence = Base58.decode(this.chainId);
        essence.writeUInt32LE(this.scHname, essence.length);
        essence.writeUInt32LE(hFuncName, essence.length);
        essence = Buffer.concat([essence, args.encode(), this.keyPair.publicKey]);
        essence.writeBigUInt64LE(BigInt(performance.now()), essence.length);
        essence = Buffer.concat([essence, transfer.encode()]);

        let buf = Buffer.alloc(0);
        const requestTypeOffledger = 1;
        buf.writeUInt8(requestTypeOffledger, 0);
        buf = Buffer.concat([buf, essence, ED25519.privateSign(this.keyPair, essence)]);

        await this.client.sendOffLedgerRequest(this.chainId, request);
        await this.client.sendExecutionRequest(this.chainId, wasp.OffLedger.GetRequestId(request));
    }
}