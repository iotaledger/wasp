// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from '../sandbox';
import {bech32Decode, bech32Encode, hexDecode, hexEncode, WasmDecoder, WasmEncoder, zeroes} from './codec';
import {Proxy} from './proxy';
import {bytesCompare} from './scbytes';
import {ScAgentID} from './scagentid';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export const ScAddressAlias: u8 = 8;
export const ScAddressEd25519: u8 = 0;
export const ScAddressNFT: u8 = 16;
export const ScAddressEth: u8 = 32;

export const ScLengthAlias = 33;
export const ScLengthEd25519 = 33;
export const ScLengthNFT = 33;

export const ScAddressLength = ScLengthEd25519;
export const ScAddressEthLength = 21;

export class ScAddress {
    id: Uint8Array = zeroes(ScAddressLength);

    asAgentID(): ScAgentID {
        // agentID for address has Hname zero
        return ScAgentID.fromAddress(this);
    }

    public equals(other: ScAddress): bool {
        return bytesCompare(this.id, other.id) == 0;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
        return addressToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return addressToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

// TODO address type-dependent encoding/decoding?
export function addressDecode(dec: WasmDecoder): ScAddress {
    const addr = new ScAddress();
    addr.id = dec.fixedBytes(ScAddressLength);
    return addr;
}

export function addressEncode(enc: WasmEncoder, value: ScAddress): void {
    enc.fixedBytes(value.id, ScAddressLength);
}

export function addressFromBytes(buf: Uint8Array | null): ScAddress {
    const addr = new ScAddress();
    if (buf === null || buf.length == 0) {
        return addr;
    }
    switch (buf[0]) {
        case ScAddressAlias:
            if (buf.length != ScLengthAlias) {
                panic('invalid Address length: Alias');
            }
            break;
        case ScAddressEd25519:
            if (buf.length != ScLengthEd25519) {
                panic('invalid Address length: Ed25519');
            }
            break;
        case ScAddressNFT:
            if (buf.length != ScLengthNFT) {
                panic('invalid Address length: NFT');
            }
            break;
        case ScAddressEth:
            if (buf.length != ScAddressEthLength) {
                panic('invalid Address length: Eth');
            }
            break;
        default:
            panic('invalid Address type');
    }
    for (let i = 0; i < buf.length; i++) {
        addr.id[i] = buf[i];
    }
    return addr;
}

export function addressToBytes(value: ScAddress): Uint8Array {
    switch (value.id[0]) {
        case ScAddressAlias:
            return value.id.slice(0, ScLengthAlias);
        case ScAddressEd25519:
            return value.id.slice(0, ScLengthEd25519);
        case ScAddressNFT:
            return value.id.slice(0, ScLengthNFT);
        case ScAddressEth:
            return value.id.slice(0, ScAddressEthLength);
        default:
            panic('unexpected Address type');
    }
    return new Uint8Array(0);
}

export function addressFromString(value: string): ScAddress {
    if (value.indexOf('0x') == 0) {
        const hexBytes = hexDecode(value);
        const b = new Uint8Array(hexBytes.length + 1);
        b[0] = ScAddressEth;
        b.set(hexBytes, 1);
        return addressFromBytes(b);
    }
    return bech32Decode(value);
}

export function addressToString(value: ScAddress): string {
    if (value.id[0] == ScAddressEth) {
        return hexEncode(value.id.slice(1, ScAddressEthLength));
    }
    return bech32Encode(value);
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableAddress {
    proxy: Proxy;

    constructor(proxy: Proxy) {
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
