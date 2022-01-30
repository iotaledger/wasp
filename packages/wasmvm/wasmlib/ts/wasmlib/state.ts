// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {IKvStore} from "./wasmtypes"
import {stateDelete, stateExists, stateGet, stateSet} from "./host";

export class ScImmutableState {
    exists(key: u8[]): bool {
        return stateExists(key);
    }

    get(key: u8[]): u8[] | null {
        return stateGet(key);
    }
}

export class ScState implements IKvStore {
    public static proxy(): wasmtypes.Proxy {
        return new wasmtypes.Proxy(new ScState());
    }

    delete(key: u8[]): void {
        stateDelete(key);
    }

    exists(key: u8[]): bool {
        return stateExists(key);
    }

    get(key: u8[]): u8[] | null {
        return stateGet(key);
    }

    public immutable(): ScImmutableState {
        return this;
    }

    set(key: u8[], value: u8[]): void {
        stateSet(key, value);
    }
}
