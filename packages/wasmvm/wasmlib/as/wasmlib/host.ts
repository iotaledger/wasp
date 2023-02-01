// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from './sandbox';

export interface ScHost {
    exportName(index: i32, name: string): void;

    sandbox(funcNr: i32, params: Uint8Array | null): Uint8Array;

    stateDelete(key: Uint8Array): void;

    stateExists(key: Uint8Array): bool;

    stateGet(key: Uint8Array): Uint8Array | null;

    stateSet(key: Uint8Array, value: Uint8Array): void;
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

class NullVmHost implements ScHost {
    exportName(index: i32, name: string): void {
        panic('NullVmHost::exportName');
    }

    sandbox(funcNr: i32, params: Uint8Array | null): Uint8Array {
        panic('NullVmHost::sandbox');
        return new Uint8Array(0);
    }

    stateDelete(key: Uint8Array): void {
        panic('NullVmHost::stateDelete');
    }

    stateExists(key: Uint8Array): bool {
        panic('NullVmHost::stateExists');
        return false;
    }

    stateGet(key: Uint8Array): Uint8Array | null {
        panic('NullVmHost::stateGet');
        return null;
    }

    stateSet(key: Uint8Array, value: Uint8Array): void {
        panic('NullVmHost::stateSet');
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

let host: ScHost = new NullVmHost();

export function connectHost(h: ScHost): ScHost {
    const oldHost = host;
    host = h;
    return oldHost;
}

export function exportName(index: i32, name: string): void {
    host.exportName(index, name);
}

export function sandbox(funcNr: i32, params: Uint8Array | null): Uint8Array {
    return host.sandbox(funcNr, params);
}

export function stateDelete(key: Uint8Array): void {
    host.stateDelete(key);
}

export function stateExists(key: Uint8Array): bool {
    return host.stateExists(key);
}

export function stateGet(key: Uint8Array): Uint8Array | null {
    return host.stateGet(key);
}

export function stateSet(key: Uint8Array, value: Uint8Array): void {
    host.stateSet(key, value);
}
