// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScFuncContext} from "./context";

export interface IEventHandler {
    callHandler(topic: string, params: string[]): void;
}

export class EventEncoder {
    event: string;

    constructor(eventName: string) {
        this.event = eventName;
        let timestamp = new ScFuncContext().timestamp();
        // convert nanoseconds to seconds
        this.encode(wasmtypes.uint64ToString(timestamp / 1_000_000_000));
    }

    emit(): void {
        new ScFuncContext().event(this.event);
    }

    encode(value: string): void {
        value = value.replaceAll("~", "~~")
        value = value.replaceAll("|", "~/")
        value = value.replaceAll(" ", "~_")
        this.event += "|" + value;
    }
}

export class EventDecoder {
    msg: string[];

    constructor(msg: string[]) {
        this.msg = msg;
    }

    decode(): string {
        return this.msg.shift();
    }

    timestamp(): u32 {
        return wasmtypes.uint32FromString(this.decode());
    }
}
