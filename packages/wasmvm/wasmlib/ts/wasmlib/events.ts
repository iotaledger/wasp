// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScSandboxFunc} from "./sandbox";

export class EventEncoder {
    event: string;

    // constructs an encoder
    constructor(eventName: string) {
        this.event = eventName;
        let timestamp = new ScSandboxFunc().timestamp();
        // convert nanoseconds to seconds
        this.encode(wasmtypes.uint64ToString(timestamp / 1_000_000_000));
    }

    emit(): void {
        new ScSandboxFunc().event(this.event);
    }

    encode(value: string): void {
        //TODO encode potential vertical bars that are present in the value string
        this.event += "|" + value;
    }
}
