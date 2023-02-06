// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Bech32, Blake2b} from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import * as isc from './';

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
    chainId: string;
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
    public static bech32Decode(bech32: string): [string, wasmlib.ScAddress, isc.Error] {
        const dec = Bech32.decode(bech32);
        if (dec == undefined) {
            return ['', new wasmlib.ScAddress(), 'invalid bech32 string: ' + bech32];
        }
        return [dec.humanReadablePart, wasmlib.addressFromBytes(dec.data), null];
    }

    public static bech32Encode(hrp: string, addr: wasmlib.ScAddress): string {
        return Bech32.encode(hrp, wasmlib.addressToBytes(addr));
    }

    public static hNameBytes(name: string): Uint8Array {
        const data = wasmlib.stringToBytes(name);
        const hash = Blake2b.sum256(data);

        // follow exact algorithm from packages/isc/hname.go
        let slice = hash.slice(0, 4);
        const hName = wasmlib.uint32FromBytes(slice);
        if (hName == 0 || hName == 0xffff) {
            slice = hash.slice(4, 8);
        }
        return slice;
    }

    public static jsonDecode(dict: JsonResp): Uint8Array {
        const enc = new wasmlib.WasmEncoder();
        const items = dict.Items;
        enc.fixedBytes(wasmlib.uint32ToBytes(items.length as u32), wasmlib.ScUint32Length);
        for (let i = 0; i < items.length; i++) {
            const item = items[i];
            const key = wasmlib.hexDecode(item.key);
            const val = wasmlib.hexDecode(item.value);
            enc.fixedBytes(wasmlib.uint16ToBytes(key.length as u16), wasmlib.ScUint16Length);
            enc.fixedBytes(key, key.length as u32);
            enc.fixedBytes(wasmlib.uint32ToBytes(val.length as u32), wasmlib.ScUint32Length);
            enc.fixedBytes(val, val.length as u32);
        }
        return enc.buf();
    }

    public static jsonEncode(buf: Uint8Array): JsonReq {
        const dict = new JsonReq();
        const dec = new wasmlib.WasmDecoder(buf);
        const size = wasmlib.uint32FromBytes(dec.fixedBytes(wasmlib.ScUint32Length));
        for (let i: u32 = 0; i < size; i++) {
            const keyBuf = dec.fixedBytes(wasmlib.ScUint16Length);
            const keyLen = wasmlib.uint16FromBytes(keyBuf);
            const key = dec.fixedBytes(keyLen as u32);
            const valBuf = dec.fixedBytes(wasmlib.ScUint32Length);
            const valLen = wasmlib.uint32FromBytes(valBuf);
            const val = dec.fixedBytes(valLen);
            const item = new JsonItem();
            item.key = wasmlib.hexEncode(key);
            item.value = wasmlib.hexEncode(val);
            dict.Items.push(item);
        }
        return dict;
    }
}