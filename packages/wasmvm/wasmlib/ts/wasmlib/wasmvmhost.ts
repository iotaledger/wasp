// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "./wasmtypes"
import {ScHost} from "./host"

// interface WasmLib to the VM host

// These 2 external functions are funneling the entire
// WasmLib functionality to their counterparts on the host.

@external("WasmLib", "hostStateGet")
export declare function hostStateGet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): i32;

@external("WasmLib", "hostStateSet")
export declare function hostStateSet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): void;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class WasmVMHost implements ScHost {
    exportName(index: i32, name: string): void {
        const buf = wasmtypes.stringToBytes(name);
        hostStateSet(0, index, buf.dataStart, buf.length as i32);
    }

    sandbox(funcNr: i32, params: u8[] | null): u8[] {
        if (params === null) {
            params = [];
        }
        // call sandbox function, result value will be cached by host
        // always negative funcNr as keyLen indicates sandbox call
        // this removes the need for a separate hostSandbox function
        let size = hostStateGet(0, funcNr, params.dataStart, params.length as i32);

        // zero length, no need to retrieve cached value
        if (size == 0) {
            return [];
        }

        // retrieve cached value from host
        let result: u8[] = new Array(size);
        hostStateGet(0, 0, result.dataStart, size);
        return result;
    }

    stateDelete(key: u8[]): void {
        hostStateSet(key.dataStart, key.length as i32, 0, -1);
    }

    stateExists(key: u8[]): bool {
        return hostStateGet(key.dataStart, key.length as i32, 0, -1) >= 0;
    }

    stateGet(key: u8[]): u8[] | null {
        // variable sized result expected,
        // query size first by passing zero length buffer
        // value will be cached by host
        let size = hostStateGet(key.dataStart, key.length as i32, 0, 0);

        // -1 means non-existent
        if (size < 0) {
            return null;
        }

        // zero length, no need to retrieve cached value
        if (size == 0) {
            return [];
        }

        // retrieve cached value from host
        let result: u8[] = new Array(size);
        hostStateGet(0, 0, result.dataStart, size);
        return result;
    }

    stateSet(key: u8[], value: u8[]): void {
        hostStateSet(key.dataStart, key.length as i32, value.dataStart, value.length as i32);
    }
}
