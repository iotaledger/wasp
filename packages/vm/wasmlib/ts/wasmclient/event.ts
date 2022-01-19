// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Base58} from "./crypto";

export class Event {
    private index = 0;
    private readonly msg: string[];
    public readonly timestamp: wasmclient.Int32;

    protected constructor(msg: string[]) {
        this.msg = msg;
        this.timestamp = Number(this.next());
    }

    private next(): string {
        return this.msg[this.index++] ?? "";
    }

    protected nextAddress(): wasmclient.Address {
        return this.next();
    }

    protected nextAgentID(): wasmclient.AgentID {
        return this.next();
    }

    protected nextBool(): wasmclient.Bool {
        return this.next() != "0";
    }

    protected nextBytes(): wasmclient.Bytes {
        return Base58.decode(this.next());
    }

    protected nextChainID(): wasmclient.ChainID {
        return this.next();
    }

    protected nextColor(): wasmclient.Color {
        return this.next();
    }

    protected nextHash(): wasmclient.Hash {
        return this.next();
    }

    protected nextHname(): wasmclient.Hname {
        return Number(this.next());
    }

    protected nextInt8(): wasmclient.Int8 {
        return Number(this.next());
    }

    protected nextInt16(): wasmclient.Int16 {
        return Number(this.next());
    }

    protected nextInt32(): wasmclient.Int32 {
        return Number(this.next());
    }

    protected nextInt64(): wasmclient.Int64 {
        return BigInt(this.next());
    }

    protected nextRequestID(): wasmclient.RequestID {
        return this.next();
    }

    protected nextString(): string {
        return this.next();
    }

    protected nextUint8(): wasmclient.Uint8 {
        return Number(this.next());
    }

    protected nextUint16(): wasmclient.Uint16 {
        return Number(this.next());
    }

    protected nextUint32(): wasmclient.Uint32 {
        return Number(this.next());
    }

    protected nextUint64(): wasmclient.Uint64 {
        return BigInt(this.next());
    }
}
