// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmtypes from "wasmlib/wasmtypes"
import {connectHost, ScHost} from "wasmlib/host"

// interface WasmLib to the VM host

// These 2 external functions are funneling the entire
// WasmLib functionality to their counterparts on the host.

@external("WasmLib", "hostStateGet")
export declare function hostStateGet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): i32;

@external("WasmLib", "hostStateSet")
export declare function hostStateSet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): void;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class WasmVMHost implements ScHost {
    static connect(): void {
        connectHost(new WasmVMHost());
    }

    exportName(index: i32, name: string): void {
        const buf = wasmtypes.stringToBytes(name);
        hostStateSet(0, index, buf.dataStart, buf.length as i32);
    }

    sandbox(funcNr: i32, params: Uint8Array | null): Uint8Array {
        if (params === null) {
            params = new Uint8Array(0);
        }
        // call sandbox function, result value will be cached by host
        // always negative funcNr as keyLen indicates sandbox call
        // this removes the need for a separate hostSandbox function
        let size = hostStateGet(0, funcNr, params.dataStart, params.length as i32);

        // zero length, no need to retrieve cached value
        if (size == 0) {
            return new Uint8Array(0);
        }

        // retrieve cached value from host
        let result = new Uint8Array(size);
        hostStateGet(0, 0, result.dataStart, size);
        return result;
    }

    stateDelete(key: Uint8Array): void {
        hostStateSet(key.dataStart, key.length as i32, 0, -1);
    }

    stateExists(key: Uint8Array): bool {
        return hostStateGet(key.dataStart, key.length as i32, 0, -1) >= 0;
    }

    stateGet(key: Uint8Array): Uint8Array | null {
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
            return new Uint8Array(0);
        }

        // retrieve cached value from host
        let result = new Uint8Array(size);
        hostStateGet(0, 0, result.dataStart, size);
        return result;
    }

    stateSet(key: Uint8Array, value: Uint8Array): void {
        hostStateSet(key.dataStart, key.length as i32, value.dataStart, value.length as i32);
    }
}
