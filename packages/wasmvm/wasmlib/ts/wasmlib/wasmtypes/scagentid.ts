// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import {WasmDecoder, WasmEncoder} from "./codec";
import {Proxy} from "./proxy";
import {addressDecode, addressEncode, addressFromBytes, ScAddress, ScAddressAlias, ScAddressLength} from "./scaddress";
import {hnameDecode, hnameEncode, hnameFromBytes, ScHname} from "./schname";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAgentIDLength = 37;

export class ScAgentID {
    _address: ScAddress;
    _hname: ScHname;

    constructor(address: ScAddress, hname: ScHname) {
        this._address = address;
        this._hname = hname;
    }

    public equals(other: ScAgentID): bool {
        return this._address.equals(other._address) &&
            this._hname.equals(other._hname);
    }

    public address(): ScAddress {
        return this._address;
    }

    public hname(): ScHname {
        return this._hname;
    }

    public isAddress(): bool {
        return this._hname.equals(new ScHname(0));
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

export function agentIDDecode(dec: WasmDecoder): ScAgentID {
    return new ScAgentID(addressDecode(dec), hnameDecode(dec))
}

export function agentIDEncode(enc: WasmEncoder, value: ScAgentID): void {
    addressEncode(enc, value._address);
    hnameEncode(enc, value._hname);
}

export function agentIDFromBytes(buf: u8[]): ScAgentID {
    if (buf.length == 0) {
        return new ScAgentID(addressFromBytes(buf), hnameFromBytes(buf));
    }
    if (buf.length != ScAgentIDLength) {
        panic("invalid AgentID length");
    }
    if (buf[0] > ScAddressAlias) {
        panic("invalid AgentID address type");
    }
    return new ScAgentID(
        addressFromBytes(buf.slice(0, ScAddressLength)),
        hnameFromBytes(buf.slice(ScAddressLength)));
}

export function agentIDToBytes(value: ScAgentID): u8[] {
    const enc = new WasmEncoder();
    agentIDEncode(enc, value);
    return enc.buf();
}

export function agentIDToString(value: ScAgentID): string {
    // TODO standardize human readable string
    return value._address.toString() + "::" + value._hname.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableAgentID {
    proxy: Proxy;

    constructor(proxy: Proxy) {
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
