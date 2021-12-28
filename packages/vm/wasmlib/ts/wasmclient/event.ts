// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Base58} from "./crypto";

export class Event {
    private index: wasmclient.Int32;
    private message: string[];
    public readonly timestamp: wasmclient.Int32;

    constructor(message: string[]) {
        this.message = message;
        this.index = 0;
        this.timestamp = Number(this.next());
    }

    private next(): string {
        return this.message[this.index++];
    }

    nextAddress(): wasmclient.Address {
        return this.next();
    }

    nextAgentID(): wasmclient.AgentID {
        return this.next();
    }

    nextBool(): wasmclient.Bool {
        return this.next() != "0";
    }

    nextBytes(): wasmclient.Bytes {
        return Base58.decode(this.next());
    }

    nextChainID(): wasmclient.ChainID {
        return this.next();
    }

    nextColor(): wasmclient.Color {
        return this.next();
    }

    nextHash(): wasmclient.Hash {
        return this.next();
    }

    nextHname(): wasmclient.Hname {
        return Number(this.next());
    }

    nextInt8(): wasmclient.Int8 {
        return Number(this.next());
    }

    nextInt16(): wasmclient.Int16 {
        return Number(this.next());
    }

    nextInt32(): wasmclient.Int32 {
        return Number(this.next());
    }

    nextInt64(): wasmclient.Int64 {
        return BigInt(this.next());
    }

    nextRequestID(): wasmclient.RequestID {
        return this.next();
    }

    nextString(): string {
        return this.next();
    }

    nextUint8(): wasmclient.Uint8 {
        return Number(this.next());
    }

    nextUint16(): wasmclient.Uint16 {
        return Number(this.next());
    }

    nextUint32(): wasmclient.Uint32 {
        return Number(this.next());
    }

    nextUint64(): wasmclient.Uint64 {
        return BigInt(this.next());
    }
}
