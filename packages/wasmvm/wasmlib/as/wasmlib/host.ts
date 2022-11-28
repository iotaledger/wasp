// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

export interface ScHost {
    exportName(index: i32, name: string): void;

    sandbox(funcNr: i32, params: Uint8Array | null): Uint8Array;

    stateDelete(key: Uint8Array): void;

    stateExists(key: Uint8Array): bool;

    stateGet(key: Uint8Array): Uint8Array | null;

    stateSet(key: Uint8Array, value: Uint8Array): void;
}

let host: ScHost;

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
