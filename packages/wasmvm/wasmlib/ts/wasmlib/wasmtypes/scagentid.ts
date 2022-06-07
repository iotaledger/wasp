// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index"

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAgentIDNil: u8 = 0;
export const ScAgentIDAddress: u8 = 1;
export const ScAgentIDContract: u8 = 2;
export const ScAgentIDEthereum: u8 = 3;

export class ScAgentID {
    kind: u8;
    _address: wasmtypes.ScAddress;
    _hname: wasmtypes.ScHname;

    constructor(address: wasmtypes.ScAddress, hname: wasmtypes.ScHname) {
        this.kind = ScAgentIDContract;
        this._address = address;
        this._hname = hname;
    }

    public static fromAddress(address: wasmtypes.ScAddress): ScAgentID {
        const agentID = new ScAgentID(address, new wasmtypes.ScHname(0));
        if (address.id[0] != wasmtypes.ScAddressAlias) {
            agentID.kind = ScAgentIDAddress;
        }
        return agentID;
    }

    public equals(other: ScAgentID): bool {
        return this._address.equals(other._address) &&
            this._hname.equals(other._hname);
    }

    public address(): wasmtypes.ScAddress {
        return this._address;
    }

    public hname(): wasmtypes.ScHname {
        return this._hname;
    }

    public isAddress(): bool {
        return this.kind == ScAgentIDAddress;
    }

    public isContract(): bool {
        return this.kind == ScAgentIDContract;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return agentIDToBytes(this)
    }

    // human-readable string representation
    public toString(): string {
        return agentIDToString(this)
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function agentIDDecode(dec: wasmtypes.WasmDecoder): ScAgentID {
    return agentIDFromBytes(dec.bytes());
}

export function agentIDEncode(enc: wasmtypes.WasmEncoder, value: ScAgentID): void {
    enc.bytes(agentIDToBytes(value));
}

export function agentIDFromBytes(buf: u8[]): ScAgentID {
    if (buf.length == 0) {
        const agentID = ScAgentID.fromAddress(wasmtypes.addressFromBytes([]));
        agentID.kind = ScAgentIDNil;
        return agentID;
    }
    switch (buf[0]) {
        case ScAgentIDAddress: {
            buf = buf.slice(1)
            if (buf.length != wasmtypes.ScLengthAlias && buf.length != wasmtypes.ScLengthEd25519) {
                panic("invalid AgentID length: address agentID");
            }
            return ScAgentID.fromAddress(wasmtypes.addressFromBytes(buf));
        }
        case ScAgentIDContract: {
            buf = buf.slice(1)
            if (buf.length != wasmtypes.ScChainIDLength + wasmtypes.ScHnameLength) {
                panic("invalid AgentID length: contract agentID");
            }
            const chainID = wasmtypes.chainIDFromBytes(buf.slice(0, wasmtypes.ScChainIDLength));
            const hname = wasmtypes.hnameFromBytes(buf.slice(wasmtypes.ScChainIDLength));
            return new ScAgentID(chainID.address(), hname);
        }
        case ScAgentIDEthereum:
            panic("AgentIDFromBytes: unsupported ScAgentIDEthereum");
            break;
        case ScAgentIDNil:
            break;
        default: {
            panic("AgentIDFromBytes: invalid AgentID type");
            break;
        }
    }
    return agentIDFromBytes([]);
}

export function agentIDToBytes(value: ScAgentID): u8[] {
    let buf: u8[] = [value.kind];
    switch (value.kind) {
        case wasmtypes.ScAgentIDAddress:
            return buf.concat(wasmtypes.addressToBytes(value._address));
        case wasmtypes.ScAgentIDContract: {
            buf = wasmtypes.addressToBytes(value._address);
            buf[0] = value.kind;
            return buf.concat(wasmtypes.hnameToBytes(value._hname));
        }
        case ScAgentIDEthereum:
            panic("AgentIDToBytes: unsupported ScAgentIDEthereum");
            break;
        case ScAgentIDNil:
            panic("AgentIDToBytes: unsupported ScAgentIDNil");
            break;
        default: {
            panic("AgentIDToBytes: invalid AgentID type");
            break;
        }
    }
    return buf;
}

export function agentIDFromString(value: string): ScAgentID {
    //TODO ScAgentIDEthereum / ScAgentIDNil
    const parts = value.split("@");
    switch (parts.length) {
        case 1:
            return ScAgentID.fromAddress(wasmtypes.addressFromString(parts[0]));
        case 2:
            return new ScAgentID(wasmtypes.addressFromString(parts[1]), wasmtypes.hnameFromString(parts[0]));
        default:
            panic("invalid AgentID string");
            return agentIDFromBytes([]);
    }
}

export function agentIDToString(value: ScAgentID): string {
    //TODO ScAgentIDEthereum / ScAgentIDNil
    if (value.kind == ScAgentIDContract) {
        return wasmtypes.hnameToString(value.hname()) + "@" + wasmtypes.addressToString(value.address())
    }
    return wasmtypes.addressToString(value.address())
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableAgentID {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return agentIDToString(this.value());
    }

    value(): ScAgentID {
        return agentIDFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableAgentID extends ScImmutableAgentID {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScAgentID): void {
        this.proxy.set(agentIDToBytes(value));
    }
}
