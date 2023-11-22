// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Bech32, Blake2b} from '@iota/crypto.js';
import create from 'keccak';
import * as wasmlib from 'wasmlib';
import {panic} from 'wasmlib';
import * as iscclient from './';

export type Error = string | null;

export class JsonItem {
    key = '';
    value = '';
}

export class JsonReq {
    Items: JsonItem[] = [];
}

export interface APICallViewRequest {
    contractHName: string;
    functionHName: string;
    arguments: JsonReq;
}

export interface APIOffLedgerRequest {
    chainId: string;
    request: string;
}

export class JsonResp {
    Items: JsonItem[] = [];
    Message = '';
    StatusCode: u16 = 0;
}

// Thank you, @iota/crypto.js, for making my life easy
export class Codec {
    public static bech32Decode(bech32: string): [string, wasmlib.ScAddress, iscclient.Error] {
        const dec = Bech32.decode(bech32);
        if (dec == undefined) {
            return ['', new wasmlib.ScAddress(), 'invalid bech32 string: ' + bech32];
        }
        return [dec.humanReadablePart, wasmlib.addressFromBytes(dec.data), null];
    }

    public static bech32Encode(hrp: string, addr: wasmlib.ScAddress): string {
        return Bech32.encode(hrp, wasmlib.addressToBytes(addr));
    }

    public static hashKeccak(buf: Uint8Array): Uint8Array {
        const buffer = Buffer.alloc(buf.length);
        for (let i = 0; i < buf.length; ++i) {
            buffer[i] = buf[i];
        }
        return new Uint8Array(create('keccak256').update(buffer).digest());
    }

    public static hashName(name: string): Uint8Array {
        const data = wasmlib.stringToBytes(name);
        const hash = Blake2b.sum256(data);
        for (let i = 0; i < hash.length; i += wasmlib.ScHnameLength) {
            const slice = hash.slice(i, i + wasmlib.ScHnameLength);
            const hName = wasmlib.uint32FromBytes(slice);
            if (hName != 0) {
                return slice;
            }
        }
        // astronomically unlikely to end up here
        return wasmlib.uint32ToBytes(1);
    }

    public static hrpForClient(): string {
        return hrpForClient;
    }

    public static jsonDecode(dict: JsonResp): Uint8Array {
        const enc = new wasmlib.WasmEncoder();
        const items = dict.Items;
        enc.vluEncode(items.length as u32);
        for (let i = 0; i < items.length; i++) {
            const item = items[i];
            const key = wasmlib.hexDecode(item.key);
            const val = wasmlib.hexDecode(item.value);
            enc.vluEncode(key.length as u32);
            enc.fixedBytes(key, key.length as u32);
            enc.vluEncode(val.length as u32);
            enc.fixedBytes(val, val.length as u32);
        }
        return enc.buf();
    }

    public static jsonEncode(buf: Uint8Array): JsonReq {
        const dict = new JsonReq();
        const dec = new wasmlib.WasmDecoder(buf);
        const size = dec.vluDecode(32);
        for (let i: u32 = 0; i < size; i++) {
            const keyLen = dec.vluDecode(32);
            const key = dec.fixedBytes(keyLen as u32);
            const valLen = dec.vluDecode(32);
            const val = dec.fixedBytes(valLen);
            const item = new JsonItem();
            item.key = wasmlib.hexEncode(key);
            item.value = wasmlib.hexEncode(val);
            dict.Items.push(item);
        }
        return dict;
    }
}

let hrpForClient = '';

export function clientBech32Decode(bech32: string): wasmlib.ScAddress {
    const [hrp, addr, err] = iscclient.Codec.bech32Decode(bech32);
    if (err != null) {
        panic(err);
    }
    if (hrp != hrpForClient) {
        panic('invalid protocol prefix: ' + hrp);
    }
    return addr;
}

export function clientBech32Encode(addr: wasmlib.ScAddress): string {
    return iscclient.Codec.bech32Encode(hrpForClient, addr);
}

export function clientHashKeccak(buf: Uint8Array): wasmlib.ScHash {
    return wasmlib.hashFromBytes(iscclient.Codec.hashKeccak(buf));
}

export function clientHashName(name: string): wasmlib.ScHname {
    const hName = new wasmlib.ScHname(0);
    hName.id = iscclient.Codec.hashName(name);
    return hName;
}

export function setSandboxWrappers(chainID: string): Error {
    wasmlib.sandboxWrappers(clientBech32Decode, clientBech32Encode, clientHashKeccak, clientHashName);

    // set the network prefix for the current network
    const [hrp, _addr, err] = iscclient.Codec.bech32Decode(chainID);
    if (err != null) {
        return err;
    }
    if (hrpForClient != hrp && hrpForClient != '') {
        panic('WasmClient can only connect to one Tangle network per app');
    }
    hrpForClient = hrp;
    return null;
}
