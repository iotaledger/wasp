// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as client from "./index"
import {Base58} from "./crypto";

export class Event {
    private index: client.Int32;
    private message: string[];
    public readonly timestamp: client.Int32;

    constructor(message: string[]) {
        this.message = message;
        this.index = 0;
        this.timestamp = Number(this.next());
    }

    private next(): string {
        return this.message[this.index++];
    }

    nextAddress(): client.Address {
        return this.next();
    }

    nextAgentID(): client.AgentID {
        return this.next();
    }

    nextBool(): client.Bool {
        return this.next() != "0";
    }

    nextBytes(): client.Bytes {
        return Base58.decode(this.next());
    }

    nextChainID(): client.ChainID {
        return this.next();
    }

    nextColor(): client.Color {
        return this.next();
    }

    nextHash(): client.Hash {
        return this.next();
    }

    nextHname(): client.Hname {
        return Number(this.next());
    }

    nextInt8(): client.Int8 {
        return Number(this.next());
    }

    nextInt16(): client.Int16 {
        return Number(this.next());
    }

    nextInt32(): client.Int32 {
        return Number(this.next());
    }

    nextInt64(): client.Int64 {
        return BigInt(this.next());
    }

    nextRequestID(): client.RequestID {
        return this.next();
    }

    nextString(): string {
        return this.next();
    }

    nextUint8(): client.Uint8 {
        return Number(this.next());
    }

    nextUint16(): client.Uint16 {
        return Number(this.next());
    }

    nextUint32(): client.Uint32 {
        return Number(this.next());
    }

    nextUint64(): client.Uint64 {
        return BigInt(this.next());
    }
}
