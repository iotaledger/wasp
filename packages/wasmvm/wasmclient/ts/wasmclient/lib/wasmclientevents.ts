// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Converter} from '@iota/util.js';
import * as iscclient from './iscclient';
import * as wasmlib from 'wasmlib';
import {RawData, WebSocket} from 'ws';

export class Event {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    payload: Uint8Array;
    timestamp: string;
    topic: string;

    public constructor(chainID: string, event: any) {
        this.chainID = wasmlib.chainIDFromString(chainID);
        this.contractID = new wasmlib.ScHname(event.contractID);
        this.topic = event.topic;
        this.timestamp = event.timestamp.toString();
        const timestamp = wasmlib.uint64FromString(this.timestamp);
        this.payload = wasmlib.concat(wasmlib.uint64ToBytes(timestamp), Converter.base64ToBytes(event.payload));
    }
}

export class WasmClientEvents {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    handler: wasmlib.IEventHandlers;

    constructor(chainID: wasmlib.ScChainID, contractID: wasmlib.ScHname, handler: wasmlib.IEventHandlers) {
        this.chainID = chainID;
        this.contractID = contractID;
        this.handler = handler;
    }

    public static startEventLoop(ws: WebSocket, eventHandlers: WasmClientEvents[]): iscclient.Error {
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
            console.log(json);
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

        for (const item of msg.payload) {
            const event = new Event(msg.chainID, item);
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

    private processEvent(event: Event) {
        if (!event.contractID.equals(this.contractID) || !event.chainID.equals(this.chainID)) {
            return;
        }
        console.log(event.chainID.toString() + ' ' + event.contractID.toString() + ' ' + event.topic);
        const dec = new wasmlib.WasmDecoder(event.payload);
        this.handler.callHandler(event.topic, dec);
    }
}
