// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// mutable proxies to host objects

import {base58Encode, ROOT} from "./context";
import {Convert} from "./convert";
import {ScAddress, ScAgentID, ScChainID, ScColor, ScHash, ScHname, ScRequestID} from "./hashtypes";
import * as host from "./host";
import {
    ScImmutableMap,
    ScImmutableMapArray,
    ScImmutableStringArray,
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

    // get value proxy for mutable ScAgentID field specified by key
    getAgentID(key: MapKey): ScMutableAgentID {
        return new ScMutableAgentID(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Bool field specified by key
    getBool(key: MapKey): ScMutableBool {
        return new ScMutableBool(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Bytes array field specified by key
    getBytes(key: MapKey): ScMutableBytes {
        return new ScMutableBytes(this.objID, key.getKeyID());
    }

    // get value proxy for mutable ScChainID field specified by key
    getChainID(key: MapKey): ScMutableChainID {
        return new ScMutableChainID(this.objID, key.getKeyID());
    }

    // get value proxy for mutable ScColor field specified by key
    getColor(key: MapKey): ScMutableColor {
        return new ScMutableColor(this.objID, key.getKeyID());
    }

    // get value proxy for mutable ScHash field specified by key
    getHash(key: MapKey): ScMutableHash {
        return new ScMutableHash(this.objID, key.getKeyID());
    }

    // get value proxy for mutable ScHname field specified by key
    getHname(key: MapKey): ScMutableHname {
        return new ScMutableHname(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Int8 field specified by key
    getInt8(key: MapKey): ScMutableInt8 {
        return new ScMutableInt8(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Int16 field specified by key
    getInt16(key: MapKey): ScMutableInt16 {
        return new ScMutableInt16(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Int32 field specified by key
    getInt32(key: MapKey): ScMutableInt32 {
        return new ScMutableInt32(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Int64 field specified by key
    getInt64(key: MapKey): ScMutableInt64 {
        return new ScMutableInt64(this.objID, key.getKeyID());
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

    // get value proxy for mutable Uint16 field specified by key
    getUint16(key: MapKey): ScMutableUint16 {
        return new ScMutableUint16(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Uint32 field specified by key
    getUint32(key: MapKey): ScMutableUint32 {
        return new ScMutableUint32(this.objID, key.getKeyID());
    }

    // get value proxy for mutable Uint64 field specified by key
    getUint64(key: MapKey): ScMutableUint64 {
        return new ScMutableUint64(this.objID, key.getKeyID());
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
