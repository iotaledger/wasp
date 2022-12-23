// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {stateDelete, stateExists, stateGet, stateSet} from "./host";
import {IKvStore, Proxy} from "./wasmtypes/proxy";

export class ScImmutableState {
    exists(key: Uint8Array): bool {
        return stateExists(key);
    }

    get(key: Uint8Array): Uint8Array {
        const val = stateGet(key);
        return val === null ? new Uint8Array(0) : val;
    }
}

export class ScState implements IKvStore {
    public static proxy(): Proxy {
        return new Proxy(new ScState());
    }

    delete(key: Uint8Array): void {
        stateDelete(key);
    }

    exists(key: Uint8Array): bool {
        return stateExists(key);
    }

    get(key: Uint8Array): Uint8Array {
        const val = stateGet(key);
        return val === null ? new Uint8Array(0) : val;
    }

    public immutable(): ScImmutableState {
        return this;
    }

    set(key: Uint8Array, value: Uint8Array): void {
        stateSet(key, value);
    }
}
