// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as isc from './isc';
import * as wasmlib from 'wasmlib';
import {IEventHandlers, panic} from 'wasmlib';
import {RawData, WebSocket} from 'ws';

export class ContractEvent {
    chainID: wasmlib.ScChainID;
    contractID: wasmlib.ScHname;
    data: string;

    public constructor(chainID: string, contractID: string, data: string) {
        this.chainID = wasmlib.chainIDFromString(chainID);
        this.contractID = wasmlib.hnameFromString(contractID);
        this.data = data;
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
            const event = new ContractEvent(msg.chainID, parts[0], parts[1]);
            for (const h of eventHandlers) {
                h.processEvent(event);
            }
        }
    }

    private processEvent(event: ContractEvent) {
        if (!event.contractID.equals(this.contractID) || !event.chainID.equals(this.chainID)) {
            return;
        }
        console.log(event.chainID.toString() + ' ' + event.contractID.toString() + ' ' + event.data);
        const params = event.data.split('|');
        for (let i = 0; i < params.length; i++) {
            params[i] = this.unescape(params[i]);
        }
        const topic = params[0];
        params.shift();
        this.handler.callHandler(topic, params);
    }

    private static subscribe(ws: WebSocket, topic: string) {
        const msg = {
            command: 'subscribe',
            topic: topic,
        };
        const rawMsg = JSON.stringify(msg);
        ws.send(rawMsg);
    }

    private unescape(param: string): string {
        const i = param.indexOf('~');
        if (i < 0) {
            return param;
        }

        switch (param.charAt(i + 1)) {
            case '~': // escaped escape character
                return param.slice(0, i) + '~' + this.unescape(param.slice(i + 2));
            case '/': // escaped vertical bar
                return param.slice(0, i) + '|' + this.unescape(param.slice(i + 2));
            case '_': // escaped space
                return param.slice(0, i) + ' ' + this.unescape(param.slice(i + 2));
            default:
                panic('invalid event encoding');
        }
        return '';
    }
}
