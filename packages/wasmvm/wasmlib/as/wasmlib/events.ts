// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {ScFuncContext} from './context';
import {uint64FromString, uint64ToString} from './wasmtypes/scuint64';

export interface IEventHandlers {
    callHandler(topic: string, params: string[]): void;
    id(): u32;
}

let nextID: u32 = 0;

export function eventHandlersGenerateID(): u32 {
    nextID++;
    return nextID;
}

export class EventEncoder {
    event: string;

    constructor(eventName: string) {
        this.event = eventName;
        const timestamp = new ScFuncContext().timestamp();
        // convert nanoseconds to seconds
        this.encode(uint64ToString(timestamp / 1_000_000_000));
    }

    emit(): void {
        new ScFuncContext().event(this.event);
    }

    encode(value: string): void {
        value = value.replaceAll('~', '~~');
        value = value.replaceAll('|', '~/');
        value = value.replaceAll(' ', '~_');
        this.event += '|' + value;
    }
}

export class EventDecoder {
    msg: string[];

    constructor(msg: string[]) {
        this.msg = msg;
    }

    decode(): string {
        return this.msg.shift()!;
    }

    timestamp(): u64 {
        return uint64FromString(this.decode());
    }
}
