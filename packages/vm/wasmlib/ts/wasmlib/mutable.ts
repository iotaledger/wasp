// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// mutable proxies to host objects

import {base58Encode, ROOT} from "./context";
import {Convert} from "./convert";
import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import * as host from "./host";
import {
    ScImmutableAddressArray,
    ScImmutableAgentIDArray,
    ScImmutableBoolArray,
    ScImmutableBytesArray,
    ScImmutableChainIDArray,
    ScImmutableColorArray,
    ScImmutableHashArray,
    ScImmutableHnameArray,
    ScImmutableInt8Array,
    ScImmutableInt16Array,
    ScImmutableInt32Array,
    ScImmutableInt64Array,
    ScImmutableMap,
    ScImmutableMapArray,
    ScImmutableRequestIDArray,
    ScImmutableStringArray,
    ScImmutableUint8Array,
    ScImmutableUint16Array,
    ScImmutableUint32Array,
    ScImmutableUint64Array,
} from "./immutable";
import {Key32, KEY_MAPS, MapKey} from "./keys";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAddress in host container
export class ScMutableAddress {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_ADDRESS);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_ADDRESS);
    }

    // set value in host container
    setValue(val: ScAddress): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_ADDRESS, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScAddress {
        return ScAddress.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_ADDRESS));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScAddress
export class ScMutableAddressArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getAddress(index: i32): ScMutableAddress {
        return new ScMutableAddress(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableAddressArray {
        return new ScImmutableAddressArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScAgentID in host container
export class ScMutableAgentID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_AGENT_ID);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_AGENT_ID);
    }

    // set value in host container
    setValue(val: ScAgentID): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_AGENT_ID, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScAgentID {
        return ScAgentID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_AGENT_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScAgentID
export class ScMutableAgentIDArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getAgentID(index: i32): ScMutableAgentID {
        return new ScMutableAgentID(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableAgentIDArray {
        return new ScImmutableAgentIDArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Bool in host container
export class ScMutableBool {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_BOOL);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_BOOL);
    }

    // set value in host container
    setValue(val: boolean): void {
        let bytes: u8[] = [(val ? 1 : 0) as u8]
        host.setBytes(this.objID, this.keyID, host.TYPE_BOOL, bytes);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): boolean {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_BOOL);
        return bytes[0] != 0;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Bool
export class ScMutableBoolArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getBool(index: i32): ScMutableBool {
        return new ScMutableBool(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableBoolArray {
        return new ScImmutableBoolArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable bytes array in host container
export class ScMutableBytes {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_BYTES);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_BYTES);
    }

    // set value in host container
    setValue(val: u8[]): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_BYTES, val);
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.value());
    }

    // retrieve value from host container
    value(): u8[] {
        return host.getBytes(this.objID, this.keyID, host.TYPE_BYTES);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of byte array
export class ScMutableBytesArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getBytes(index: i32): ScMutableBytes {
        return new ScMutableBytes(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableBytesArray {
        return new ScImmutableBytesArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScChainID in host container
export class ScMutableChainID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_CHAIN_ID);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_CHAIN_ID);
    }

    // set value in host container
    setValue(val: ScChainID): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_CHAIN_ID, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScChainID {
        return ScChainID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_CHAIN_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScChainID
export class ScMutableChainIDArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getChainID(index: i32): ScMutableChainID {
        return new ScMutableChainID(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableChainIDArray {
        return new ScImmutableChainIDArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScColor in host container
export class ScMutableColor {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_COLOR);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_COLOR);
    }

    // set value in host container
    setValue(val: ScColor): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_COLOR, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScColor {
        return ScColor.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_COLOR));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScColor
export class ScMutableColorArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getColor(index: i32): ScMutableColor {
        return new ScMutableColor(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableColorArray {
        return new ScImmutableColorArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHash in host container
export class ScMutableHash {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_HASH);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_HASH);
    }

    // set value in host container
    setValue(val: ScHash): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_HASH, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScHash {
        return ScHash.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_HASH));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScHash
export class ScMutableHashArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getHash(index: i32): ScMutableHash {
        return new ScMutableHash(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableHashArray {
        return new ScImmutableHashArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScHname in host container
export class ScMutableHname {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_HNAME);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_HNAME);
    }

    // set value in host container
    setValue(val: ScHname): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_HNAME, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScHname {
        return ScHname.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_HNAME));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScHname
export class ScMutableHnameArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getHname(index: i32): ScMutableHname {
        return new ScMutableHname(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableHnameArray {
        return new ScImmutableHnameArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int8 in host container
export class ScMutableInt8 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT8);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT8);
    }

    // set value in host container
    setValue(val: i8): void {
        let bytes: u8[] = [val as u8];
        host.setBytes(this.objID, this.keyID, host.TYPE_INT8, bytes);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): i8 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT8);
        return bytes[0] as i8;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int8
export class ScMutableInt8Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getInt8(index: i32): ScMutableInt8 {
        return new ScMutableInt8(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableInt8Array {
        return new ScImmutableInt8Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int16 in host container
export class ScMutableInt16 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT16);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT16);
    }

    // set value in host container
    setValue(val: i16): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT16, Convert.fromI16(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): i16 {
        return Convert.toI16(host.getBytes(this.objID, this.keyID, host.TYPE_INT16));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int16
export class ScMutableInt16Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getInt16(index: i32): ScMutableInt16 {
        return new ScMutableInt16(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableInt16Array {
        return new ScImmutableInt16Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int32 in host container
export class ScMutableInt32 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT32);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT32);
    }

    // set value in host container
    setValue(val: i32): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT32, Convert.fromI32(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): i32 {
        return Convert.toI32(host.getBytes(this.objID, this.keyID, host.TYPE_INT32));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int32
export class ScMutableInt32Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getInt32(index: i32): ScMutableInt32 {
        return new ScMutableInt32(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableInt32Array {
        return new ScImmutableInt32Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Int64 in host container
export class ScMutableInt64 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT64);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT64);
    }

    // set value in host container
    setValue(val: i64): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT64, Convert.fromI64(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): i64 {
        return Convert.toI64(host.getBytes(this.objID, this.keyID, host.TYPE_INT64));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Int64
export class ScMutableInt64Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getInt64(index: i32): ScMutableInt64 {
        return new ScMutableInt64(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableInt64Array {
        return new ScImmutableInt64Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// map proxy for mutable map
export class ScMutableMap {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // construct a new map on the host and return a map proxy for it
    static create(): ScMutableMap {
        let maps = ROOT.getMapArray(KEY_MAPS);
        return maps.getMap(maps.length());
    }

    callFunc(keyID: Key32, params: u8[]): u8[] {
        return host.callFunc(this.objID, keyID, params);
    }

    // empty the map
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for mutable ScAddress field specified by key
    getAddress(key: MapKey): ScMutableAddress {
        return new ScMutableAddress(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableAddressArray specified by key
    getAddressArray(key: MapKey): ScMutableAddressArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_ADDRESS | host.TYPE_ARRAY);
        return new ScMutableAddressArray(arrID);
    }

    // get value proxy for mutable ScAgentID field specified by key
    getAgentID(key: MapKey): ScMutableAgentID {
        return new ScMutableAgentID(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableAgentIDArray specified by key
    getAgentIDArray(key: MapKey): ScMutableAgentIDArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_AGENT_ID | host.TYPE_ARRAY);
        return new ScMutableAgentIDArray(arrID);
    }

    // get value proxy for mutable Bool field specified by key
    getBool(key: MapKey): ScMutableBool {
        return new ScMutableBool(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableBoolArray specified by key
    getBoolArray(key: MapKey): ScMutableBoolArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_BOOL | host.TYPE_ARRAY);
        return new ScMutableBoolArray(arrID);
    }

    // get value proxy for mutable Bytes array field specified by key
    getBytes(key: MapKey): ScMutableBytes {
        return new ScMutableBytes(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableBytesArray specified by key
    getBytesArray(key: MapKey): ScMutableBytesArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_BYTES | host.TYPE_ARRAY);
        return new ScMutableBytesArray(arrID);
    }

    // get value proxy for mutable ScChainID field specified by key
    getChainID(key: MapKey): ScMutableChainID {
        return new ScMutableChainID(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableChainIDArray specified by key
    getChainIDArray(key: MapKey): ScMutableChainIDArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_CHAIN_ID | host.TYPE_ARRAY);
        return new ScMutableChainIDArray(arrID);
    }

    // get value proxy for mutable ScColor field specified by key
    getColor(key: MapKey): ScMutableColor {
        return new ScMutableColor(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableColorArray specified by key
    getColorArray(key: MapKey): ScMutableColorArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_COLOR | host.TYPE_ARRAY);
        return new ScMutableColorArray(arrID);
    }

    // get value proxy for mutable ScHash field specified by key
    getHash(key: MapKey): ScMutableHash {
        return new ScMutableHash(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableHashArray specified by key
    getHashArray(key: MapKey): ScMutableHashArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_HASH | host.TYPE_ARRAY);
        return new ScMutableHashArray(arrID);
    }

    // get value proxy for mutable ScHname field specified by key
    getHname(key: MapKey): ScMutableHname {
        return new ScMutableHname(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableHnameArray specified by key
    getHnameArray(key: MapKey): ScMutableHnameArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_HNAME | host.TYPE_ARRAY);
        return new ScMutableHnameArray(arrID);
    }

    // get value proxy for mutable Int8 field specified by key
    getInt8(key: MapKey): ScMutableInt8 {
        return new ScMutableInt8(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableInt8Array specified by key
    getInt8Array(key: MapKey): ScMutableInt8Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT8 | host.TYPE_ARRAY);
        return new ScMutableInt8Array(arrID);
    }

    // get value proxy for mutable Int16 field specified by key
    getInt16(key: MapKey): ScMutableInt16 {
        return new ScMutableInt16(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableInt16Array specified by key
    getInt16Array(key: MapKey): ScMutableInt16Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT16 | host.TYPE_ARRAY);
        return new ScMutableInt16Array(arrID);
    }

    // get value proxy for mutable Int32 field specified by key
    getInt32(key: MapKey): ScMutableInt32 {
        return new ScMutableInt32(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableInt32Array specified by key
    getInt32Array(key: MapKey): ScMutableInt32Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT32 | host.TYPE_ARRAY);
        return new ScMutableInt32Array(arrID);
    }

    // get value proxy for mutable Int64 field specified by key
    getInt64(key: MapKey): ScMutableInt64 {
        return new ScMutableInt64(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableInt64Array specified by key
    getInt64Array(key: MapKey): ScMutableInt64Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT64 | host.TYPE_ARRAY);
        return new ScMutableInt64Array(arrID);
    }

    // get map proxy for ScMutableMap specified by key
    getMap(key: MapKey): ScMutableMap {
        let mapID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_MAP);
        return new ScMutableMap(mapID);
    }

    // get array proxy for ScMutableMapArray specified by key
    getMapArray(key: MapKey): ScMutableMapArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_MAP | host.TYPE_ARRAY);
        return new ScMutableMapArray(arrID);
    }

    // get value proxy for mutable ScRequestID field specified by key
    getRequestID(key: MapKey): ScMutableRequestID {
        return new ScMutableRequestID(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableRequestIDArray specified by key
    getRequestIDArray(key: MapKey): ScMutableRequestIDArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_REQUEST_ID | host.TYPE_ARRAY);
        return new ScMutableRequestIDArray(arrID);
    }

    // get value proxy for mutable UTF-8 text string field specified by key
    getString(key: MapKey): ScMutableString {
        return new ScMutableString(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableStringArray specified by key
    getStringArray(key: MapKey): ScMutableStringArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_STRING | host.TYPE_ARRAY);
        return new ScMutableStringArray(arrID);
    }

    // get value proxy for mutable Uint8 field specified by key
    getUint8(key: MapKey): ScMutableUint8 {
        return new ScMutableUint8(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableUint8Array specified by key
    getUint8Array(key: MapKey): ScMutableUint8Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT8 | host.TYPE_ARRAY);
        return new ScMutableUint8Array(arrID);
    }

    // get value proxy for mutable Uint16 field specified by key
    getUint16(key: MapKey): ScMutableUint16 {
        return new ScMutableUint16(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableUint16Array specified by key
    getUint16Array(key: MapKey): ScMutableUint16Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT16 | host.TYPE_ARRAY);
        return new ScMutableUint16Array(arrID);
    }

    // get value proxy for mutable Uint32 field specified by key
    getUint32(key: MapKey): ScMutableUint32 {
        return new ScMutableUint32(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableUint32Array specified by key
    getUint32Array(key: MapKey): ScMutableUint32Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT32 | host.TYPE_ARRAY);
        return new ScMutableUint32Array(arrID);
    }

    // get value proxy for mutable Uint64 field specified by key
    getUint64(key: MapKey): ScMutableUint64 {
        return new ScMutableUint64(this.objID, key.getKeyID());
    }

    // get array proxy for ScMutableUint64Array specified by key
    getUint64Array(key: MapKey): ScMutableUint64Array {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_INT64 | host.TYPE_ARRAY);
        return new ScMutableUint64Array(arrID);
    }

    // get immutable version of map proxy
    immutable(): ScImmutableMap {
        return new ScImmutableMap(this.objID);
    }

    mapID(): i32 {
        return this.objID;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of maps
export class ScMutableMapArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getMap(index: i32): ScMutableMap {
        let mapID = host.getObjectID(this.objID, new Key32(index), host.TYPE_MAP);
        return new ScMutableMap(mapID);
    }

    // get immutable version of array proxy
    immutable(): ScImmutableMapArray {
        return new ScImmutableMapArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable ScRequestID in host container
export class ScMutableRequestID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_REQUEST_ID);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_REQUEST_ID);
    }

    // set value in host container
    setValue(val: ScRequestID): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_REQUEST_ID, val.toBytes());
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): ScRequestID {
        return ScRequestID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_REQUEST_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of ScRequestID
export class ScMutableRequestIDArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getRequestID(index: i32): ScMutableRequestID {
        return new ScMutableRequestID(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableRequestIDArray {
        return new ScImmutableRequestIDArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable UTF-8 text string in host container
export class ScMutableString {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_STRING);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_STRING);
    }

    // set value in host container
    setValue(val: string): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_STRING, Convert.fromString(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value();
    }

    // retrieve value from host container
    value(): string {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_STRING);
        return Convert.toString(bytes);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of UTF-8 text string
export class ScMutableStringArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getString(index: i32): ScMutableString {
        return new ScMutableString(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableStringArray {
        return new ScImmutableStringArray(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint8 in host container
export class ScMutableUint8 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT8);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT8);
    }

    // set value in host container
    setValue(val: u8): void {
        let bytes: u8[] = [val as u8];
        host.setBytes(this.objID, this.keyID, host.TYPE_INT8, bytes);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): u8 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT8);
        return bytes[0] as u8;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint8
export class ScMutableUint8Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getUint8(index: i32): ScMutableUint8 {
        return new ScMutableUint8(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableUint8Array {
        return new ScImmutableUint8Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint16 in host container
export class ScMutableUint16 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT16);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT16);
    }

    // set value in host container
    setValue(val: u16): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT16, Convert.fromI16(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): u16 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT16);
        return Convert.toI16(bytes) as u16;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint16
export class ScMutableUint16Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getUint16(index: i32): ScMutableUint16 {
        return new ScMutableUint16(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableUint16Array {
        return new ScImmutableUint16Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint32 in host container
export class ScMutableUint32 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT32);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT32);
    }

    // set value in host container
    setValue(val: u32): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT32, Convert.fromI32(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): u32 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT32);
        return Convert.toI32(bytes) as u32;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint32
export class ScMutableUint32Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getUint32(index: i32): ScMutableUint32 {
        return new ScMutableUint32(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableUint32Array {
        return new ScImmutableUint32Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for mutable Uint64 in host container
export class ScMutableUint64 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // delete value from host container
    delete(): void {
        host.delKey(this.objID, this.keyID, host.TYPE_INT64);
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT64);
    }

    // set value in host container
    setValue(val: u64): void {
        host.setBytes(this.objID, this.keyID, host.TYPE_INT64, Convert.fromI64(val));
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // retrieve value from host container
    value(): u64 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT64);
        return Convert.toI64(bytes) as u64;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for mutable array of Uint64
export class ScMutableUint64Array {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // empty the array
    clear(): void {
        host.clear(this.objID);
    }

    // get value proxy for item at index, index can be 0..length()
    // when index equals length() a new item is appended
    getUint64(index: i32): ScMutableUint64 {
        return new ScMutableUint64(this.objID, new Key32(index));
    }

    // get immutable version of array proxy
    immutable(): ScImmutableUint64Array {
        return new ScImmutableUint64Array(this.objID);
    }

    // number of items in array
    length(): i32 {
        return host.getLength(this.objID);
    }
}
