// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {
    hexDecode,
    IEventHandlers,
    ScUint64Length,
    stringDecode,
    uint64Decode,
    uint64FromBytes,
    WasmDecoder
} from 'wasmlib';
import {RawData, WebSocket} from 'ws';

export class ContractEvent {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    topic: string;
    timestamp: u64;
    payload: Uint8Array;

    public constructor(chainID: string, contractID: string, dec: WasmDecoder) {
        this.chainID = wasmlib.chainIDFromString(chainID);
        this.contractID = wasmlib.hnameFromString(contractID);
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
            const parts = item.split(': ');
            const buf = hexDecode(parts[1]);
            const dec = new WasmDecoder(buf);
            const event = new ContractEvent(msg.chainID, parts[0], dec);
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
        const sep = event.data.indexOf('|');
        if (sep < 0) {
            return;
        }
        const topic = event.data.slice(0, sep);
        console.log(event.chainID.toString() + ' ' + event.contractID.toString() + ' ' + topic);
        const buf = hexDecode(event.data.slice(sep + 1));
        const dec = new WasmDecoder(buf);
        this.handler.callHandler(topic, dec);
    }
}
