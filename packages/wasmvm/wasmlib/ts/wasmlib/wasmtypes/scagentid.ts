// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index"

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAgentIDLength = 37;

export class ScAgentID {
    _address: wasmtypes.ScAddress;
    _hname: wasmtypes.ScHname;

    constructor(address: wasmtypes.ScAddress, hname: wasmtypes.ScHname) {
        this._address = address;
        this._hname = hname;
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
        return this._hname.equals(new wasmtypes.ScHname(0));
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
    return new ScAgentID(wasmtypes.addressDecode(dec), wasmtypes.hnameDecode(dec))
}

export function agentIDEncode(enc: wasmtypes.WasmEncoder, value: ScAgentID): void {
    wasmtypes.addressEncode(enc, value._address);
    wasmtypes.hnameEncode(enc, value._hname);
}

export function agentIDFromBytes(buf: u8[]): ScAgentID {
    if (buf.length == 0) {
        return new ScAgentID(wasmtypes.addressFromBytes(buf), wasmtypes.hnameFromBytes(buf));
    }
    if (buf.length != ScAgentIDLength) {
        panic("invalid AgentID length");
    }
    if (buf[0] > wasmtypes.ScAddressAlias) {
        panic("invalid AgentID address type");
    }
    return new wasmtypes.ScAgentID(
        wasmtypes.addressFromBytes(buf.slice(0, wasmtypes.ScAddressLength)),
        wasmtypes.hnameFromBytes(buf.slice(wasmtypes.ScAddressLength)));
}

export function agentIDToBytes(value: ScAgentID): u8[] {
    const enc = new wasmtypes.WasmEncoder();
    agentIDEncode(enc, value);
    return enc.buf();
}

export function agentIDToString(value: ScAgentID): string {
    // TODO standardize human readable string
    return value._address.toString() + "::" + value._hname.toString();
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
