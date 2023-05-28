// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {
    hexDecode,
    hnameDecode,
    IEventHandlers,
    ScUint64Length,
    stringDecode,
    uint64FromBytes,
    WasmDecoder
} from 'wasmlib';
import {RawData, WebSocket} from 'ws';

export class ContractEvent {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    payload: Uint8Array;
    timestamp: u64;
    topic: string;

    public constructor(chainID: string, eventData: Uint8Array) {
        this.chainID = wasmlib.chainIDFromString(chainID);
        const dec = new WasmDecoder(eventData);
        this.contractID = hnameDecode(dec);
        this.topic = stringDecode(dec);
        this.payload = dec.fixedBytes(dec.length());
        this.timestamp = uint64FromBytes(this.payload.slice(0, ScUint64Length));
    }
}

export class WasmClientEvents {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    handler: IEventHandlers;

    constructor(chainID: wasmlib.ScChainID, contractID: wasmlib.ScHname, handler: IEventHandlers) {
        this.chainID = chainID;
        this.contractID = contractID;
        this.handler = handler;
    }

    public static startEventLoop(ws: WebSocket, eventHandlers: WasmClientEvents[]): isc.Error {
        ws.on('open', () => {
            this.subscribe(ws, 'chains');
            this.subscribe(ws, 'block_events');
        });
        ws.on('error', (err) => {
            // callback(['error', err.toString()]);
        });
        ws.on('message', (data) => this.eventLoop(data, eventHandlers));
        return null;
    }

    private static eventLoop(data: RawData, eventHandlers: WasmClientEvents[]) {
        let msg: any;
        try {
            const json = data.toString();
            // console.log(json);
            msg = JSON.parse(json);
            if (!msg.kind) {
                // filter out subscribe responses
                return;
            }
            console.log(msg);
        } catch (ex) {
            console.log(`Failed to parse expected JSON message: ${data} ${ex}`);
            return;
        }

        const items: string[] = msg.payload;
        for (const item of items) {
            const eventData = hexDecode(item);
            const event = new ContractEvent(msg.chainID, eventData);
            for (const h of eventHandlers) {
                h.processEvent(event);
            }
        }
    }

    private static subscribe(ws: WebSocket, topic: string) {
        const msg = {
            command: 'subscribe',
            topic: topic,
        };
        const rawMsg = JSON.stringify(msg);
        ws.send(rawMsg);
    }

    private processEvent(event: ContractEvent) {
        if (!event.contractID.equals(this.contractID) || !event.chainID.equals(this.chainID)) {
            return;
        }
        console.log(event.chainID.toString() + ' ' + event.contractID.toString() + ' ' + event.topic);
        const dec = new WasmDecoder(event.payload);
        this.handler.callHandler(event.topic, dec);
    }
}
