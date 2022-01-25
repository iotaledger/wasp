// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// immutable proxies to host objects

import { base58Encode } from "./context";
import {Convert} from "./convert";
import {ScAddress,ScAgentID,ScChainID,ScColor,ScHash,ScHname,ScRequestID} from "./hashtypes";
import * as host from "./host";
import {Key32,MapKey} from "./keys";

// value proxy for immutable ScAddress in host container
export class ScImmutableAddress {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_ADDRESS);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScAddress {
        return ScAddress.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_ADDRESS));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScAgentID in host container
export class ScImmutableAgentID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_AGENT_ID);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScAgentID {
        return ScAgentID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_AGENT_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Bool in host container
export class ScImmutableBool {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_BOOL);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): boolean {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_BOOL);
        return bytes[0] != 0;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable bytes array in host container
export class ScImmutableBytes {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_BYTES);
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.value());
    }

    // get value from host container
    value(): u8[] {
        return host.getBytes(this.objID, this.keyID, host.TYPE_BYTES);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScChainID in host container
export class ScImmutableChainID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_CHAIN_ID);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScChainID {
        return ScChainID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_CHAIN_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScColor in host container
export class ScImmutableColor {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_COLOR);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScColor {
        return ScColor.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_COLOR));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of ScColor
export class ScImmutableColorArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // get value proxy for item at index, index can be 0..length()-1
    getColor(index: u32): ScImmutableColor {
        return new ScImmutableColor(this.objID, new Key32(index as i32));
    }

    // number of items in array
    length(): u32 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScHash in host container
export class ScImmutableHash {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_HASH);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScHash {
        return ScHash.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_HASH));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable ScHname in host container
export class ScImmutableHname {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_HNAME);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScHname {
        return ScHname.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_HNAME));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int8 in host container
export class ScImmutableInt8 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT8);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): i8 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT8);
        return bytes[0] as i8;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int16 in host container
export class ScImmutableInt16 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT16);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): i16 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT16);
        return Convert.toI16(bytes);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int32 in host container
export class ScImmutableInt32 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT32);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): i32 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT32);
        return Convert.toI32(bytes);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Int64 in host container
export class ScImmutableInt64 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT64);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): i64 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT64);
        return Convert.toI64(bytes);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// map proxy for immutable map
export class ScImmutableMap {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    callFunc(keyID: Key32, params: u8[]): u8[] {
        return host.callFunc(this.objID, keyID, params);
    }

    // get value proxy for immutable ScAddress field specified by key
    getAddress(key: MapKey): ScImmutableAddress {
        return new ScImmutableAddress(this.objID, key.getKeyID());
    }

    // get value proxy for immutable ScAgentID field specified by key
    getAgentID(key: MapKey): ScImmutableAgentID {
        return new ScImmutableAgentID(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Bool field specified by key
    getBool(key: MapKey): ScImmutableBool {
        return new ScImmutableBool(this.objID, key.getKeyID());
    }

    // get value proxy for immutable bytes array field specified by key
    getBytes(key: MapKey): ScImmutableBytes {
        return new ScImmutableBytes(this.objID, key.getKeyID());
    }

    // get value proxy for immutable ScChainID field specified by key
    getChainID(key: MapKey): ScImmutableChainID {
        return new ScImmutableChainID(this.objID, key.getKeyID());
    }

    // get value proxy for immutable ScColor field specified by key
    getColor(key: MapKey): ScImmutableColor {
        return new ScImmutableColor(this.objID, key.getKeyID());
    }

    // get array proxy for ScImmutableColorArray specified by key
    getColorArray(key: MapKey): ScImmutableColorArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_COLOR | host.TYPE_ARRAY);
        return new ScImmutableColorArray(arrID);
    }

    // get value proxy for immutable ScHash field specified by key
    getHash(key: MapKey): ScImmutableHash {
        return new ScImmutableHash(this.objID, key.getKeyID());
    }

    // get value proxy for immutable ScHname field specified by key
    getHname(key: MapKey): ScImmutableHname {
        return new ScImmutableHname(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Int8 field specified by key
    getInt8(key: MapKey): ScImmutableInt8 {
        return new ScImmutableInt8(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Int16 field specified by key
    getInt16(key: MapKey): ScImmutableInt16 {
        return new ScImmutableInt16(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Int32 field specified by key
    getInt32(key: MapKey): ScImmutableInt32 {
        return new ScImmutableInt32(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Int64 field specified by key
    getInt64(key: MapKey): ScImmutableInt64 {
        return new ScImmutableInt64(this.objID, key.getKeyID());
    }

    // get map proxy for ScImmutableMap specified by key
    getMap(key: MapKey): ScImmutableMap {
        let mapID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_MAP);
        return new ScImmutableMap(mapID);
    }

    // get array proxy for ScImmutableMapArray specified by key
    getMapArray(key: MapKey): ScImmutableMapArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_MAP | host.TYPE_ARRAY);
        return new ScImmutableMapArray(arrID);
    }

    // get value proxy for immutable ScRequestID field specified by key
    getRequestID(key: MapKey): ScImmutableRequestID {
        return new ScImmutableRequestID(this.objID, key.getKeyID());
    }

    // get value proxy for immutable UTF-8 text string field specified by key
    getString(key: MapKey): ScImmutableString {
        return new ScImmutableString(this.objID, key.getKeyID());
    }

    // get array proxy for ScImmutableStringArray specified by key
    getStringArray(key: MapKey): ScImmutableStringArray {
        let arrID = host.getObjectID(this.objID, key.getKeyID(), host.TYPE_STRING | host.TYPE_ARRAY);
        return new ScImmutableStringArray(arrID);
    }

    // get value proxy for immutable Uint8 field specified by key
    getUint8(key: MapKey): ScImmutableUint8 {
        return new ScImmutableUint8(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Uint16 field specified by key
    getUint16(key: MapKey): ScImmutableUint16 {
        return new ScImmutableUint16(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Uint32 field specified by key
    getUint32(key: MapKey): ScImmutableUint32 {
        return new ScImmutableUint32(this.objID, key.getKeyID());
    }

    // get value proxy for immutable Uint64 field specified by key
    getUint64(key: MapKey): ScImmutableUint64 {
        return new ScImmutableUint64(this.objID, key.getKeyID());
    }

    mapID(): i32 {
        return this.objID;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of maps
export class ScImmutableMapArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // get value proxy for item at index, index can be 0..length()-1
    getMap(index: u32): ScImmutableMap {
        let mapID = host.getObjectID(this.objID, new Key32(index as i32), host.TYPE_MAP);
        return new ScImmutableMap(mapID);
    }

    // number of items in array
    length(): u32 {
        return host.getLength(this.objID);
    }
}

// value proxy for immutable ScRequestID in host container
export class ScImmutableRequestID {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_REQUEST_ID);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): ScRequestID {
        return ScRequestID.fromBytes(host.getBytes(this.objID, this.keyID, host.TYPE_REQUEST_ID));
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable UTF-8 text string in host container
export class ScImmutableString {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_STRING);
    }

    // human-readable string representation
    toString(): string {
        return this.value();
    }

    // get value from host container
    value(): string {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_STRING);
        return Convert.toString(bytes);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// array proxy for immutable array of UTF-8 text string
export class ScImmutableStringArray {
    objID: i32;

    constructor(id: i32) {
        this.objID = id;
    }

    // get value proxy for item at index, index can be 0..length()-1
    getString(index: u32): ScImmutableString {
        return new ScImmutableString(this.objID, new Key32(index as i32));
    }

    // number of items in array
    length(): u2 {
        return host.getLength(this.objID);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint8 in host container
export class ScImmutableUint8 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT8);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): u8 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT8);
        return bytes[0] as u8;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint16 in host container
export class ScImmutableUint16 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT16);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): u16 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT16);
        return Convert.toI16(bytes) as u16;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint32 in host container
export class ScImmutableUint32 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT32);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): u32 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT32);
        return Convert.toI32(bytes) as u32;
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value proxy for immutable Uint64 in host container
export class ScImmutableUint64 {
    objID: i32;
    keyID: Key32;

    constructor(objID: i32, keyID: Key32) {
        this.objID = objID;
        this.keyID = keyID;
    }

    // check if value exists in host container
    exists(): boolean {
        return host.exists(this.objID, this.keyID, host.TYPE_INT64);
    }

    // human-readable string representation
    toString(): string {
        return this.value().toString();
    }

    // get value from host container
    value(): u64 {
        let bytes = host.getBytes(this.objID, this.keyID, host.TYPE_INT64);
        return Convert.toI64(bytes) as u64;
    }
}
