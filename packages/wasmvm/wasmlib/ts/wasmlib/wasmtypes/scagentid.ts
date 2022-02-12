// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from "../sandbox";
import * as wasmtypes from "./index"

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const nilAgentID: u8 = 0xff;
const nilAgent = agentIDFromBytes([]);

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

// note: only alias address can have a non-zero hname
// so there is no need to encode it when it is always zero

export function agentIDDecode(dec: wasmtypes.WasmDecoder): ScAgentID {
    if (dec.peek() == wasmtypes.ScAddressAlias) {
        return new ScAgentID(wasmtypes.addressDecode(dec), wasmtypes.hnameDecode(dec))
    }
    return new ScAgentID(wasmtypes.addressDecode(dec), new wasmtypes.ScHname(0));
}

export function agentIDEncode(enc: wasmtypes.WasmEncoder, value: ScAgentID): void {
    wasmtypes.addressEncode(enc, value._address);
    if (value._address.id[0] == wasmtypes.ScAddressAlias) {
        wasmtypes.hnameEncode(enc, value._hname);
    }
}

export function agentIDFromBytes(buf: u8[]): ScAgentID {
    if (buf.length == 0) {
        return new ScAgentID(wasmtypes.addressFromBytes([]), new wasmtypes.ScHname(0));
    }
    switch (buf[0]) {
        case wasmtypes.ScAddressAlias: {
            if (buf.length != wasmtypes.ScLengthAlias + wasmtypes.ScHnameLength) {
                panic("invalid AgentID length: Alias address");
            }
            const addr = wasmtypes.addressFromBytes(buf.slice(0, wasmtypes.ScLengthAlias));
            const hname = wasmtypes.hnameFromBytes(buf.slice(wasmtypes.ScLengthAlias));
            return new ScAgentID(addr, hname);
        }
        case wasmtypes.ScAddressEd25519: {
            if (buf.length != wasmtypes.ScLengthEd25519) {
                panic("invalid AgentID length: Ed25519 address");
            }
            return new ScAgentID(wasmtypes.addressFromBytes(buf), new wasmtypes.ScHname(0));
        }
        case wasmtypes.ScAddressNFT: {
            if (buf.length != wasmtypes.ScLengthNFT) {
                panic("invalid AgentID length: NFT address");
            }
            return new ScAgentID(wasmtypes.addressFromBytes(buf), new wasmtypes.ScHname(0));
        }
        case nilAgentID: {
            if (buf.length != 1) {
                panic("invalid AgentID length: nil AgentID")
            }
            break
        }
        default: {
            panic("invalid AgentID address type");
            break;
        }
    }
    return agentIDFromBytes([]);
}

export function agentIDToBytes(value: ScAgentID): u8[] {
    if (value.equals(nilAgent)) {
        return [nilAgentID]
    }
    let buf = wasmtypes.addressToBytes(value._address);
    if (buf[0] == wasmtypes.ScAddressAlias) {
        buf = buf.concat(wasmtypes.hnameToBytes(value._hname))
    }
    return buf;
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
