// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

export interface ScHost {
    exportName(index: i32, name: string): void;

    sandbox(funcNr: i32, params: u8[] | null): u8[];

    stateDelete(key: u8[]): void;

    stateExists(key: u8[]): bool;

    stateGet(key: u8[]): u8[] | null;

    stateSet(key: u8[], value: u8[]): void;
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

export function sandbox(funcNr: i32, params: u8[] | null): u8[] {
    return host.sandbox(funcNr, params);
}

export function stateDelete(key: u8[]): void {
    host.stateDelete(key);
}

export function stateExists(key: u8[]): bool {
    return host.stateExists(key);
}

export function stateGet(key: u8[]): u8[] | null {
    return host.stateGet(key);
}

export function stateSet(key: u8[], value: u8[]): void {
    host.stateSet(key, value);
}
