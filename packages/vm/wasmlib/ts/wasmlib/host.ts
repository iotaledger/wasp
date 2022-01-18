// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// interface WasmLib to the VM host

// all type id values should exactly match their counterpart values on the host!

import * as keys from "./keys";
import {Key32} from "./keys";
import {Convert} from "./convert";

export const TYPE_ARRAY: i32 = 0x20;
export const TYPE_ARRAY16: i32 = 0x60;
export const TYPE_CALL: i32 = 0x80;
export const TYPE_MASK: i32 = 0x1f;

export const TYPE_ADDRESS: i32 = 1;
export const TYPE_AGENT_ID: i32 = 2;
export const TYPE_BOOL: i32 = 3;
export const TYPE_BYTES: i32 = 4;
export const TYPE_CHAIN_ID: i32 = 5;
export const TYPE_COLOR: i32 = 6;
export const TYPE_HASH: i32 = 7;
export const TYPE_HNAME: i32 = 8;
export const TYPE_INT8: i32 = 9;
export const TYPE_INT16: i32 = 10;
export const TYPE_INT32: i32 = 11;
export const TYPE_INT64: i32 = 12;
export const TYPE_MAP: i32 = 13;
export const TYPE_REQUEST_ID: i32 = 14;
export const TYPE_STRING: i32 = 15;

export const OBJ_ID_NULL: i32 = 0;
export const OBJ_ID_ROOT: i32 = 1;
export const OBJ_ID_STATE: i32 = 2;
export const OBJ_ID_PARAMS: i32 = 3;
export const OBJ_ID_RESULTS: i32 = 4;

// size in bytes of predefined types, indexed by the TYPE_* consts
export const TYPE_SIZES: u8[] = [0, 33, 37, 1, 0, 33, 32, 32, 4, 1, 2, 4, 8, 0, 34, 0];


// These 4 external functions are funneling the entire WasmLib functionality
// to their counterparts on the host.

// Copy the value data bytes of type <typeID> stored in the host container object <objID>,
// under key <keyID>, into the pre-allocated <buffer> which can hold len bytes.
// Returns the actual length of the value data bytes on the host.
@external("WasmLib", "hostGetBytes")
export declare function hostGetBytes(objID: i32, keyID: i32, typeID: i32, buffer: usize, size: i32): i32;

// Retrieve the key id associated with the <key> data bytes of length <len>.
// A negative length indicates a bytes key, positive indicates a string key
// We discern between the two for better readable logging purposes
@external("WasmLib", "hostGetKeyID")
export declare function hostGetKeyID(key: usize, size: i32): i32;

// Retrieve the id of the container sub-object of type <typeID> stored in
// the host container object <objID>, under key <keyID>.
@external("WasmLib", "hostGetObjectID")
export declare function hostGetObjectID(objID: i32, keyID: i32, typeID: i32): i32;

// copy the <len> value data bytes of type <typeID> from the <buffer>
// into the host container object <objID>, under key <keyID>.
@external("WasmLib", "hostSetBytes")
export declare function hostSetBytes(objID: i32, keyID: i32, typeID: i32, buffer: usize, size: i32): void;

@external("WasmLib", "hostStateGet")
export declare function hostStateGet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): i32;

@external("WasmLib", "hostStateSet")
export declare function hostStateSet(keyRef: usize, keyLen: i32, valRef: usize, valLen: i32): void;


export function callFunc(objID: i32, keyID: Key32, params: u8[]): u8[] {
    // variable-sized type, first query expected length of bytes array
    // (pass zero-length buffer)
    let size = hostGetBytes(objID, keyID.keyID, TYPE_CALL, params.dataStart, params.length);

    // -1 means non-existent, so return default value for type
    if (size <= 0) {
        return [];
    }

    // allocate a sufficient length byte array in Wasm memory
    // and let the host copy the actual data bytes into this Wasm byte array
    let result: u8[] = new Array(size);
    hostGetBytes(objID, keyID.keyID, TYPE_CALL + 1, result.dataStart, size);
    return result;
}

// Clear the entire contents of the specified container object.
// Removes all its sub-objects as well.
export function clear(objID: i32): void {
    // special key "length" is used with integer value zero
    setBytes(objID, keys.KEY_LENGTH, TYPE_INT32, Convert.fromI32(0));
}

// Delete the value with the specified key and type from the specified container object.
export function delKey(objID: i32, keyID: Key32, typeID: i32): void {
    // size -1 means delete
    // this removes the need for a separate hostDelete function
    hostSetBytes(objID, keyID.keyID, typeID, 0, -1);
}

// Check if the specified container object contains a value with the specified key and type.
export function exists(objID: i32, keyID: Key32, typeID: i32): boolean {
    // negative length (-1) means only test for existence
    // returned size -1 indicates keyID not found (or error)
    // this removes the need for a separate hostExists function
    return hostGetBytes(objID, keyID.keyID, typeID, 0, -1) >= 0;
}

// Retrieve the bytes stored in the specified container object under the specified key
// and with specified type. Note that if the key does not exist this function will
// return the default value for the specified type.
export function getBytes(objID: i32, keyID: Key32, typeID: i32): u8[] {
    let size = TYPE_SIZES[typeID] as i32;
    if (size == 0) {
        // variable-sized type, first query expected length of bytes array
        // (pass zero-length buffer)
        size = hostGetBytes(objID, keyID.keyID, typeID, 0, 0);

        // -1 means non-existent, so return default value for type
        if (size < 0) {
            return [];
        }
    }

    // allocate a sufficient length byte array in Wasm memory
    // and let the host copy the actual data bytes into this Wasm byte array
    let result: u8[] = new Array(size);
    hostGetBytes(objID, keyID.keyID, typeID, result.dataStart, size);
    return result;
}

// Retrieve the key id that the host has associated with the specified bytes key
export function getKeyID(bytes: u8[], size: i32): Key32 {
    // negative size indicates this is a bytes key
    return new Key32(hostGetKeyID(bytes.dataStart, size));
}

// Retrieve the key id that the host has associated with the specified bytes key
export function getKeyIDFromBytes(bytes: u8[]): Key32 {
    // negative size indicates this is a bytes key
    return getKeyID(bytes, -bytes.length - 1);
}

// Retrieve the key id that the host has associated with the specified string key
export function getKeyIDFromString(key: string): Key32 {
    let bytes = Convert.fromString(key);

    // non-negative size indicates this is a string key
    return getKeyID(bytes, bytes.length);
}

// Retrieve the key id that the host has associated with the specified integer key
export function getKeyIDFromUint64(key: u64, nrOfBytes: usize): Key32 {
    // negative size indicates this is a bytes key
    return getKeyID(Convert.fromI64(key as i64), -nrOfBytes - 1);
}

// Retrieve the length of an array container object on the host
export function getLength(objID: i32): i32 {
    // special integer key "length" is used
    let bytes = getBytes(objID, keys.KEY_LENGTH, TYPE_INT32);
    return Convert.toI32(bytes);
}

// Retrieve the id of the specified container sub-object
export function getObjectID(objID: i32, keyID: Key32, typeID: i32): i32 {
    return hostGetObjectID(objID, keyID.keyID, typeID);
}

// Direct logging of informational text to host log
export function log(text: string): void {
    setBytes(1, keys.KEY_LOG, TYPE_STRING, Convert.fromString(text));
}

// Direct logging of error to host log, followed by panicking out of the Wasm code
export function panic(text: string): void {
    setBytes(1, keys.KEY_PANIC, TYPE_STRING, Convert.fromString(text));
}

// Store the provided value bytes of specified type in the specified container object
// under the specified key. Note that if the key does not exist this function will
// create it first.
export function setBytes(objID: i32, keyID: Key32, typeID: i32, value: u8[]): void {
    return hostSetBytes(objID, keyID.keyID, typeID, value.dataStart, value.length);
}

// Direct logging of debug trace text to host log
export function trace(text: string): void {
    setBytes(1, keys.KEY_TRACE, TYPE_STRING, Convert.fromString(text));
}

export function sandbox(funcNr: i32, params: u8[]): u8[] {
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

export function stateDelete(key: u8[]): void {
    hostStateSet(key.dataStart, key.length as i32, 0, -1);
}

export function stateExistst(key: u8[]): bool {
    return hostStateGet(key.dataStart, key.length as i32, 0, -1) >= 0;
}

export function stateGet(key: u8[]): u8[] | null {
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

export function stateSet(key: u8[], value: u8[]): void {
    hostStateSet(key.dataStart, key.length as i32, value.dataStart, value.length as i32);
}
