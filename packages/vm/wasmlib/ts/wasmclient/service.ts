// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Base58, ED25519, IKeyPair, Hash} from "./crypto";
import {Buffer} from "./buffer";

export type EventHandlers = { [key: string]: (message: string[]) => void };

export class Service {
    private waspClient: wasmclient.WaspClient;
    private webSocket: WebSocket;
    private webSocketUrl: string;
    public keyPair: IKeyPair;
    private eventHandlers: EventHandlers;
    public chainId: string;
    public scHname: wasmclient.Hname;

    constructor(client: wasmclient.ServiceClient, chainId: string, scHname: wasmclient.Hname, eventHandlers: EventHandlers) {
        this.waspClient = client.waspClient;
        this.chainId = chainId;
        this.scHname = scHname;
        this.startEventHandlers(client.eventPort, eventHandlers);
    }

    public async callView(viewName: string, args: wasmclient.Arguments): Promise<wasmclient.Results> {
        return await this.waspClient.callView(
            this.chainId,
            this.scHname.toString(16),
            viewName,
            args,
        );
    }

    public async postRequest(hFuncName: wasmclient.Int32, args: wasmclient.Arguments, transfer: wasmclient.Transfer, keyPair: IKeyPair): Promise<wasmclient.RequestID> {
        // get request essence ready for signing
        let essence = Base58.decode(this.chainId);
        essence.writeUInt32LE(this.scHname, essence.length);
        essence.writeUInt32LE(hFuncName, essence.length);
        essence = Buffer.concat([essence, args.encode(), keyPair.publicKey]);
        essence.writeBigUInt64LE(BigInt(performance.now()), essence.length);
        essence = Buffer.concat([essence, transfer.encode()]);

        let buf = Buffer.alloc(0);
        const requestTypeOffledger = 1;
        buf.writeUInt8(requestTypeOffledger, 0);
        buf = Buffer.concat([buf, essence, ED25519.privateSign(keyPair, essence)]);
        const hash = Hash.from(buf);
        const requestID = Buffer.concat([hash, Buffer.alloc(2)]);

        await this.waspClient.postOffLedgerRequest(this.chainId, buf);
        return Base58.encode(requestID);
    }

    public async waitRequest(reqID: wasmclient.RequestID): Promise<void> {
        await this.waspClient.waitRequest(this.chainId, reqID);
    }

    private startEventHandlers(eventPort: string, eventHandlers: EventHandlers) {
        this.webSocketUrl = "ws://" + eventPort + "/chain/" + this.chainId + "/ws";
        this.eventHandlers = eventHandlers
        this.connectWebSocket();
    }

    private connectWebSocket(): void {
        // eslint-disable-next-line no-console
        console.log(`Connecting to Websocket => ${this.webSocketUrl}`);
        this.webSocket = new WebSocket(this.webSocketUrl);
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
}