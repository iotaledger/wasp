// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScFuncContext} from './context';
import {concat, hexEncode, stringToBytes, uint64Encode, WasmDecoder, WasmEncoder} from "./wasmtypes";

export interface IEventHandlers {
    callHandler(topic: string, dec: WasmDecoder): void;

    id(): u32;
}

let nextID: u32 = 0;

export function eventHandlersGenerateID(): u32 {
    nextID++;
    return nextID;
}

export function eventEncoder(): WasmEncoder {
    const enc = new WasmEncoder();
    uint64Encode(enc, new ScFuncContext().timestamp());
    return enc;
}

export function eventEmit(topic: string, enc: WasmEncoder): void {
    new ScFuncContext().event(topic + "|" + hexEncode(enc.buf()));
}
