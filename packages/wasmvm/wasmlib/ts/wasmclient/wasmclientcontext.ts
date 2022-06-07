// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib";
import * as wasmtypes from "wasmlib/wasmtypes";
import * as wasmrequests from "wasmlib/wasmrequests";
import * as wasmclient from "./index";
import {Base58, Hash, IKeyPair} from "./crypto";
import {IOnLedger} from "./goshimmer/models/on_ledger";
import {Colors} from "./colors";
import {Buffer} from "./buffer";
import {panic} from "./index";

declare type u8 = number;
declare type i32 = number;

export interface IEventHandler {
    callHandler(topic: string, params: string[]): void;
}

export class WasmClientContext implements wasmlib.ScHost {
    private serviceClient: wasmclient.WasmClientService;
    private webSocket: WebSocket | null = null;
    public keyPair: IKeyPair | null = null;
    private eventHandlers: Array<IEventHandler> = [];
    public scHname: wasmclient.Hname;
    private waspWebSocketUrl = "";

    constructor(client: wasmclient.WasmClientService, scHname: wasmclient.Hname) {
        this.serviceClient = client;
        this.scHname = scHname;
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
        if (this.eventHandlers.length === 0) this.configureWebSocketsEventHandlers();

        for (let i = 0; i < this.eventHandlers.length; i++) {
            if (this.eventHandlers[i] === handler) {
                return;
            }
        }
        this.eventHandlers.push(handler);
    }

    public unregister(handler: IEventHandler): void {
        // remove handler
        this.eventHandlers = this.eventHandlers.filter((h) => h !== handler);
        if (this.eventHandlers.length === 0) this.webSocket?.close();
    }

    // overrides default contract name
    public serviceContractName(contractName: string): void {
        this.scHname = Hash.from(Buffer.from(contractName)).readUInt32LE(0);
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
        this.webSocket = new WebSocket(this.waspWebSocketUrl);
        this.webSocket.addEventListener("open", () => this.handleOpenWebSocket());
        this.webSocket.addEventListener("close", () => this.handleCloseWebSocket());
        this.webSocket.addEventListener("error", (x) => this.handleErrorWebSocket(x));
        this.webSocket.addEventListener("message", (x) => this.handleIncomingMessage(x));
    }

    private handleOpenWebSocket(): void {
        console.log(`Connected to Websocket => ${this.waspWebSocketUrl}`);
    }

    private handleCloseWebSocket(): void {
        console.log(`Disconnected from Websocket => ${this.waspWebSocketUrl}`);
    }

    private handleErrorWebSocket(event: Event): void {
        console.error(`Web socket error  => ${this.waspWebSocketUrl} => ${event}`);
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

/////////////////////////////////////////////////////////////////

    exportName(index: i32, name: string): void{
        panic("WasmClientContext.ExportName");
    }
    
    sandbox(funcNr: i32, args: u8[] | null): u8[]{
        switch (funcNr) {
        case wasmlib.FnCall:
            return this.fnCall(args);
        case wasmlib.FnPost:
            return this.fnPost(args);
        case wasmlib.FnUtilsBase58Encode:
            return Base58.encode(args);
        case wasmlib.FnUtilsBase58Decode:
            return Base58.decode(wasmtypes.stringFromBytes(args));
         }
        panic("implement me")
        return [];
     }
    
    stateDelete(key: u8[]): void{
        panic("WasmClientContext.StateDelete");
    }
    
    stateExists(key: u8[]): bool{
        panic("WasmClientContext.StateExists");
        return false;
    }
    
    stateGet(key: u8[]): u8[] | null{
        panic("WasmClientContext.StateGet");
        return null;
    }
    
    stateSet(key: u8[], value: u8[]): void{
        panic("WasmClientContext.StateSet");
    }

/////////////////////////////////////////////////////////////////

    private fnCall(args: u8[]): u8[] {
        const req = wasmrequests.CallRequest.fromBytes(args);
        let hContract = this.cvt.IscpHname(req.contract);
        if (hContract != this.scHname) {
            this.Err = errors.Errorf("unknown contract: %this", req.contract.toString());
            return nil;
        }
        let params, err = dict.FromBytes(req.params);
        if (err != nil) {
            this.Err = err;
            return nil;
        }
        let hFunction = this.cvt.iscpHname(req.function);
        let res, err = this.svcClient.CallViewByHname(this.chainID, hContract, hFunction, params);
        if (err != nil) {
            this.Err = err;
            return nil;
        }
        return res.toBytes();
    }

    private fnPost(args: u8[]): u8[] {
        const req = wasmrequests.PostRequest.fromBytes(args);
        let chainID = this.cvt.iscpChainID(&req.chainID);
        if (!chainID.equals(this.chainID)) {
            this.Err = errors.Errorf("unknown chain id: %this", req.chainID.toString());
            return nil;
        }
        let hContract = this.cvt.IscpHname(req.contract);
        if (hContract != this.scHname) {
            this.Err = errors.Errorf("unknown contract: %this", req.contract.toString());
            return nil;
        }
        let params, err = dict.FromBytes(req.params);
        if (err != nil) {
            this.Err = err;
            return nil;
        }
        let scAssets = new wasmlib.ScAssets(req.transfer);
        let allowance = this.cvt.iscpAllowance(scAssets);
        let hFunction = this.cvt.iscpHname(req.function);
        this.postRequestOffLedger(hFunction, params, allowance, this.keyPair);
        return nil;
    }
}
