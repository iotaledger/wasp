// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScFuncContext} from "./context";

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
        //TODO encode potential vertical bars that are present in the value string
        this.event += "|" + value;
    }
}
