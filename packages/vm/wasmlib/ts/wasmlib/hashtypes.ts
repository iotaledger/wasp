// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// standard value types used by the ISCP

import {base58Encode, ScFuncContext} from "./context";
import {Convert} from "./convert";
import {getKeyIDFromBytes, panic} from "./host";
import {Key32, MapKey} from "./keys";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

function zeroes(count: i32): u8[] {
    let buf : u8[] = new Array(count);
    buf.fill(0);
    return buf;
}

// value object for 33-byte Tangle address ids
export class ScAddress implements MapKey {
    id: u8[] = zeroes(33);

    // construct from byte array
    static fromBytes(bytes: u8[]): ScAddress {
        let o = new ScAddress();
        if (bytes.length != o.id.length) {
            panic("invalid address id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    // returns agent id representation of this Tangle address
    asAgentID(): ScAgentID {
        return ScAgentID.fromParts(this, ScHname.zero);
    }

    equals(other: ScAddress): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 37-byte agent ids
export class ScAgentID implements MapKey {
    id: u8[] = zeroes(37);

    // construct from byte array
    static fromBytes(bytes: u8[]): ScAgentID {
        let o = new ScAgentID();
        if (bytes.length != o.id.length) {
            panic("invalid agent id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    // construct from address and contract name hash
    static fromParts(address: ScAddress, hname: ScHname): ScAgentID {
        let agentID = new ScAgentID();
        agentID.id = address.id.concat(hname.id);
        return agentID;
    }

    // gets Tangle address from agent id
    address(): ScAddress {
        let address = new ScAddress();
        address.id = this.id.slice(0, address.id.length);
        return address;
    }

    equals(other: ScAgentID): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // get contract name hash for this agent
    hname(): ScHname {
        let hname = new ScHname(0);
        hname.id = this.id.slice(this.id.length - hname.id.length);
        return hname;
    }

    // checks to see if agent id represents a Tangle address
    isAddress(): boolean {
        return this.hname().equals(ScHname.zero);
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 33-byte chain ids
export class ScChainID implements MapKey {
    id: u8[] = zeroes(33);

    // construct from byte array
    static fromBytes(bytes: u8[]): ScChainID {
        let o = new ScChainID();
        if (bytes.length != o.id.length) {
            panic("invalid chain id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    // gets Tangle address from chain id
    address(): ScAddress {
        let address = new ScAddress();
        address.id = this.id.slice(0, address.id.length);
        return address;
    }

    equals(other: ScChainID): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 32-byte token color
export class ScColor implements MapKey {
    id: u8[] = new Array(32);

    // predefined colors
    static IOTA: ScColor = new ScColor(0x00);
    static MINT: ScColor = new ScColor(0xff);

    constructor(fill: u8) {
        this.id.fill(fill)
    }

    // construct from byte array
    static fromBytes(bytes: u8[]): ScColor {
        let o = new ScColor(0);
        if (bytes.length != o.id.length) {
            panic("invalid color id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    // construct from request id, this will return newly minted color
    static fromRequestID(requestID: ScRequestID): ScColor {
        let color = new ScColor(0);
        color.id = requestID.id.slice(0, color.id.length)
        return color;
    }

    equals(other: ScColor): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 32-byte hash value
export class ScHash implements MapKey {
    id: u8[] = zeroes(32);

    // construct from byte array
    static fromBytes(bytes: u8[]): ScHash {
        let o = new ScHash();
        if (bytes.length != o.id.length) {
            panic("invalid hash id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    equals(other: ScHash): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 4-byte name hash
export class ScHname implements MapKey {
    id: u8[] = zeroes(4)

    static zero: ScHname = new ScHname(0);

    // construct from name string
    constructor(id: u32) {
        this.id = Convert.fromI32(id);
    }

    static fromName(name: string): ScHname {
        return new ScFuncContext().utility().hname(name);
    }

    // construct from byte array
    static fromBytes(bytes: u8[]): ScHname {
        let o = new ScHname(0);
        if (bytes.length != o.id.length) {
            panic("invalid hname length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    equals(other: ScHname): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.id);
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        let id = Convert.toI32(this.id);
        return id.toString();
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// value object for 34-byte transaction request ids
export class ScRequestID implements MapKey {
    id: u8[] = zeroes(34);

    // construct from byte array
    static fromBytes(bytes: u8[]): ScRequestID {
        let o = new ScRequestID();
        if (bytes.length != o.id.length) {
            panic("invalid request id length");
        }
        o.id = bytes.slice(0);
        return o;
    }

    equals(other: ScRequestID): boolean {
        return Convert.equals(this.id, other.id);
    }

    // can be used as key in maps
    getKeyID(): Key32 {
        return getKeyIDFromBytes(this.toBytes());
    }

    // convert to byte array representation
    toBytes(): u8[] {
        return this.id;
    }

    // human-readable string representation
    toString(): string {
        return base58Encode(this.id);
    }
}
