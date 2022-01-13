// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index";
import {Hash, IKeyPair} from "./crypto";
import {IOnLedger} from "./goshimmer/models/on_ledger";
import {Colors} from "./colors";
import {Buffer} from './buffer';

export interface IEventHandler {
    callHandler(topic: string, params: string[]): void;
}

export class Service {
    private serviceClient: wasmclient.ServiceClient;
    private webSocket: WebSocket | null = null;
    public keyPair: IKeyPair | null = null;
    private eventHandlers: Array<IEventHandler> = new Array();
    public scHname: wasmclient.Hname;
    private waspWebSocketUrl: string = "";

    constructor(client: wasmclient.ServiceClient, scHname: wasmclient.Hname) {
        this.serviceClient = client;
        this.scHname = scHname;
        this.configureWebSocketsEventHandlers();
    }

    public async callView(viewName: string, args: wasmclient.Arguments, res: wasmclient.Results): Promise<void> {
        await this.serviceClient.waspClient.callView(
            this.serviceClient.configuration.chainId,
            this.scHname.toString(16),
            viewName,
            args.encodeCall(),
            res
        );
    }

    public async postRequest(
        hFuncName: wasmclient.Int32,
        args: wasmclient.Arguments,
        transfer: wasmclient.Transfer,
        keyPair: IKeyPair,
        onLedger: boolean
    ): Promise<string> {
        const chainId = this.serviceClient.configuration.chainId;
        if (! onLedger) {
            // requested off-ledger request
            const requestID = await this.serviceClient.waspClient.postOffLedgerRequest(chainId, this.scHname, hFuncName, args, transfer, keyPair);
            return requestID;
        }

        // requested on-ledger request
        const payload: IOnLedger = {
            contract: this.scHname,
            entrypoint: hFuncName,
            //TODO: map args
            //arguments : args
        };
        const transferAmount = transfer.get(Colors.IOTA_COLOR);
        const transactionID = await this.serviceClient.goShimmerClient.postOnLedgerRequest(chainId, payload, transferAmount, keyPair);
        if (!transactionID) throw new Error("No transaction id");
        return transactionID;
    }

    public register(handler: IEventHandler): void {
        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] === handler) {
                return;
            }
        }
        this.eventHandlers.push(handler);
    }

    public unregister(handler: IEventHandler): void {
        // remove handler
        this.eventHandlers = this.eventHandlers.filter(h => h !== handler);
    }

    // overrides default contract name
    public serviceContractName(contractName: string): void {
        this.scHname = Hash.from(Buffer.from(contractName)).readUInt32LE(0)
    }

    public async waitRequest(reqID: wasmclient.RequestID): Promise<void> {
        await this.serviceClient.waspClient.waitRequest(this.serviceClient.configuration.chainId, reqID);
    }

    private configureWebSocketsEventHandlers() {
        if (
            this.serviceClient.configuration.waspWebSocketUrl.startsWith("wss://") ||
            this.serviceClient.configuration.waspWebSocketUrl.startsWith("ws://")
        )
            this.waspWebSocketUrl = this.serviceClient.configuration.waspWebSocketUrl;
        else this.waspWebSocketUrl = "ws://" + this.serviceClient.configuration.waspWebSocketUrl;

        this.waspWebSocketUrl = this.waspWebSocketUrl.replace("%chainId", this.serviceClient.configuration.chainId);

        this.connectWebSocket();
    }

    private connectWebSocket(): void {
        // eslint-disable-next-line no-console
        console.log(`Connecting to Websocket => ${this.waspWebSocketUrl}`);
        this.webSocket = new WebSocket(this.waspWebSocketUrl);
        this.webSocket.addEventListener("message", (x) => this.handleIncomingMessage(x));
        this.webSocket.addEventListener("close", () => setTimeout(this.connectWebSocket.bind(this), 1000));
    }

    private handleIncomingMessage(message: MessageEvent<string>): void {
        // expect vmmsg <chain ID> <contract hname> contract.event|parameters
        const msg = message.data.toString().split(" ");
        if (msg.length != 4 || msg[0] != "vmmsg") {
            return;
        }
        const topics = msg[3].split("|");
        const topic = topics[0];
        const params = topics.slice(1);
        for (let i = 0; i < this.eventHandlers.length; i++) {
            this.eventHandlers[i].callHandler(topic, params);
        }
    }
}
