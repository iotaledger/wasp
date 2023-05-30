// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScFuncContext} from './context';
import {stringEncode, WasmDecoder, WasmEncoder} from "./wasmtypes";

export interface IEventHandlers {
    callHandler(topic: string, dec: WasmDecoder): void;

    id(): u32;
}

let nextID: u32 = 0;

export function eventHandlersGenerateID(): u32 {
    nextID++;
    return nextID;
}

export function eventEncoder(topic: string): WasmEncoder {
    const enc = new WasmEncoder();
    stringEncode(enc, topic);
    return enc;
}

export function eventEmit(enc: WasmEncoder): void {
    new ScFuncContext().event(enc.buf());
}
