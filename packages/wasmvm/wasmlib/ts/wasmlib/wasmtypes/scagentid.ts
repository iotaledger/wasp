// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

export const ScAgentIDNil: u8 = 0;
export const ScAgentIDAddress: u8 = 1;
export const ScAgentIDContract: u8 = 2;
export const ScAgentIDEthereum: u8 = 3;

import {panic} from '../sandbox';
import {
    addressFromBytes,
    addressFromString,
    addressToBytes,
    addressToString,
    ScAddress,
    ScAddressAlias,
    ScAddressEth,
    ScLengthAlias,
    ScLengthEd25519,
    ScLengthEth,
} from './scaddress';
import {chainIDFromBytes, ScChainIDLength} from './scchainid';
import {hnameFromBytes, hnameFromString, hnameToBytes, hnameToString, ScHname, ScHnameLength} from './schname';
import {concat, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

const nilAgentIDString = '-';

export class ScAgentID {
    kind: u8;
    _address: ScAddress;
    _hname: ScHname;
    eth: ScAddress;

    constructor(address: ScAddress, hname: ScHname) {
        this.kind = ScAgentIDContract;
        this._address = address;
        this._hname = hname;
        this.eth = new ScAddress();
    }

    public static forEthereum(chainAddress: ScAddress, ethAddress: ScAddress): ScAgentID {
        if (chainAddress.id[0] != ScAddressAlias) {
            panic("invalid eth AgentID: chain address");
        }
        if (ethAddress.id[0] != ScAddressEth) {
            panic("invalid eth AgentID: eth address");
        }
        const agentID = new ScAgentID(chainAddress, new ScHname(0));
        agentID.kind = ScAgentIDEthereum;
        agentID.eth = ethAddress;
        return agentID;
    }

    public static fromAddress(address: ScAddress): ScAgentID {
        const agentID = new ScAgentID(address, new ScHname(0));
        switch (address.id[0]) {
            case ScAddressAlias: {
                break;
            }
            case ScAddressEth: {
                panic("invalid eth AgentID: need chain address");
                break;
            }
            default: {
                agentID.kind = ScAgentIDAddress;
                break;
            }
        }
        return agentID;
    }

    public equals(other: ScAgentID): bool {
        return this._address.equals(other._address) &&
            this._hname.equals(other._hname);
    }

    public address(): ScAddress {
        return this._address;
    }

    public ethAddress(): ScAddress {
        return this.eth;
    }

    public hname(): ScHname {
        return this._hname;
    }

    public isAddress(): bool {
        return this.kind == ScAgentIDAddress;
    }

    public isContract(): bool {
        return this.kind == ScAgentIDContract;
    }

    // convert to byte array representation
    public toBytes(): Uint8Array {
        return agentIDToBytes(this);
    }

    // human-readable string representation
    public toString(): string {
        return agentIDToString(this);
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function agentIDDecode(dec: WasmDecoder): ScAgentID {
    return agentIDFromBytes(dec.bytes());
}

export function agentIDEncode(enc: WasmEncoder, value: ScAgentID): void {
    enc.bytes(agentIDToBytes(value));
}

export function agentIDFromBytes(buf: Uint8Array | null): ScAgentID {
    if (buf === null || buf.length == 0) {
        const agentID = ScAgentID.fromAddress(addressFromBytes(null));
        agentID.kind = ScAgentIDNil;
        return agentID;
    }
    switch (buf[0]) {
        case ScAgentIDAddress: {
            buf = buf.subarray(1);
            if (buf.length != ScLengthAlias && buf.length != ScLengthEd25519) {
                panic('invalid AgentID length: address agentID');
            }
            return ScAgentID.fromAddress(addressFromBytes(buf));
        }
        case ScAgentIDContract: {
            buf = buf.subarray(1);
            if (buf.length != ScChainIDLength + ScHnameLength) {
                panic('invalid AgentID length: contract agentID');
            }
            const chainID = chainIDFromBytes(buf.subarray(0, ScChainIDLength));
            const hname = hnameFromBytes(buf.subarray(ScChainIDLength));
            return new ScAgentID(chainID.address(), hname);
        }
        case ScAgentIDEthereum: {
            buf = buf.subarray(1);
            if (buf.length != ScChainIDLength + ScLengthEth) {
                panic('invalid AgentID length: eth agentID');
            }
            const chainID = chainIDFromBytes(buf.subarray(0, ScChainIDLength));
            const ethAddress = addressFromBytes(buf.subarray(ScChainIDLength));
            return ScAgentID.forEthereum(chainID.address(), ethAddress);
        }
        case ScAgentIDNil:
            break;
        default: {
            panic('AgentIDFromBytes: invalid AgentID type');
            break;
        }
    }
    return agentIDFromBytes(null);
}

export function agentIDToBytes(value: ScAgentID): Uint8Array {
    let buf = new Uint8Array(1);
    buf[0] = value.kind;
    switch (value.kind) {
        case ScAgentIDAddress:
            return concat(buf, addressToBytes(value._address));
        case ScAgentIDContract: {
            buf = addressToBytes(value._address);
            buf[0] = value.kind;
            return concat(buf, hnameToBytes(value._hname));
        }
        case ScAgentIDEthereum: {
            buf = addressToBytes(value._address);
            buf[0] = value.kind;
            return concat(buf, addressToBytes(value.eth));
        }
        case ScAgentIDNil:
            return buf;
        default: {
            panic('AgentIDToBytes: invalid AgentID type');
            break;
        }
    }
    return buf;
}

export function agentIDFromString(value: string): ScAgentID {
    if (value == nilAgentIDString) {
        return agentIDFromBytes(null);
    }

    const parts = value.split('@');
    switch (parts.length) {
        case 1:
            return ScAgentID.fromAddress(addressFromString(parts[0]));
        case 2:
            if (!value.startsWith('0x')) {
                return new ScAgentID(addressFromString(parts[1]), hnameFromString(parts[0]));
            }
            return ScAgentID.forEthereum(addressFromString(parts[1]), addressFromString(parts[0]));
        default:
            panic('invalid AgentID string');
            return agentIDFromBytes(null);
    }
}

export function agentIDToString(value: ScAgentID): string {
    switch (value.kind) {
        case ScAgentIDAddress:
            return addressToString(value.address());
        case ScAgentIDContract: {
            return hnameToString(value.hname()) + '@' + addressToString(value.address());
        }
        case ScAgentIDEthereum:
            return addressToString(value.ethAddress()) + '@' + addressToString(value.address());
        case ScAgentIDNil:
            return nilAgentIDString;
        default: {
            panic('AgentIDToString: invalid AgentID type');
            return '';
        }
    }
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
