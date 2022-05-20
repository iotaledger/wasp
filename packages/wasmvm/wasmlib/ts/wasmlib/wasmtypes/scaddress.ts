// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { panic } from "../sandbox";
import * as wasmtypes from "./index"
import {ScSandboxUtils} from "../sandboxutils";

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAddressAlias: u8 = 8;
export const ScAddressEd25519: u8 = 0;
export const ScAddressNFT: u8 = 16;

export const ScLengthAlias = 33;
export const ScLengthEd25519 = 33;
export const ScLengthNFT = 33;

export const ScAddressLength = ScLengthEd25519;

export class ScAddress {
    id: u8[] = wasmtypes.zeroes(ScAddressLength);

    asAgentID(): wasmtypes.ScAgentID {
        // agentID for address has Hname zero
        return wasmtypes.ScAgentID.fromAddress(this);
    }

    public equals(other: ScAddress): bool {
        return wasmtypes.bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): u8[] {
        return addressToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return addressToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// TODO address type-dependent encoding/decoding?
export function addressDecode(dec: wasmtypes.WasmDecoder): ScAddress {
    let addr = new ScAddress();
    addr.id = dec.fixedBytes(ScAddressLength);
    return addr;
}

export function addressEncode(enc: wasmtypes.WasmEncoder, value: ScAddress): void {
    enc.fixedBytes(value.id, ScAddressLength)
}

export function addressFromBytes(buf: u8[]): ScAddress {
    const addr = new ScAddress();
    if (buf.length == 0) {
        return addr;
    }
    switch (buf[0]) {
        case ScAddressAlias:
            if (buf.length != ScLengthAlias) {
                panic("invalid Address length: Alias");
            }
            break;
        case ScAddressEd25519:
            if (buf.length != ScLengthEd25519) {
                panic("invalid Address length: Ed25519");
            }
            break;
        case ScAddressNFT:
            if (buf.length != ScLengthNFT) {
                panic("invalid Address length: NFT");
            }
            break;
        default:
            panic("invalid Address type")
    }
    for (let i = 0; i < buf.length; i++) {
        addr.id[i] = buf[i];
    }
    return addr
}

export function addressToBytes(value: ScAddress): u8[] {
    switch (value.id[0]) {
        case ScAddressAlias:
            return value.id.slice(0, ScLengthAlias);
        case ScAddressEd25519:
            return value.id.slice(0, ScLengthEd25519);
        case ScAddressNFT:
            return value.id.slice(0, ScLengthNFT);
        default:
            panic("unexpected Address type")
    }
    return [];
}

export function addressFromString(value: string): ScAddress {
    const utils = new ScSandboxUtils();
    return utils.bech32Decode(value)
}

export function addressToString(value: ScAddress): string {
    const utils = new ScSandboxUtils();
    return utils.bech32Encode(value)
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableAddress {
    proxy: wasmtypes.Proxy;

    constructor(proxy: wasmtypes.Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return addressToString(this.value());
    }

    value(): ScAddress {
        return addressFromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableAddress extends ScImmutableAddress {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: ScAddress): void {
        this.proxy.set(addressToBytes(value));
    }
}
