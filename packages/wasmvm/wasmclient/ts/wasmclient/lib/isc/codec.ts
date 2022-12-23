// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Bech32, Blake2b } from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import * as isc from "./";
import {
    hexDecode,
    hexEncode,
    ScUint16Length,
    ScUint32Length,
    uint16FromBytes, uint16ToBytes,
    uint32FromBytes, uint32ToBytes,
    WasmDecoder,
    WasmEncoder
} from "wasmlib";

export class JsonItem {
    Key: string = "";
    Value: string = "";
}

export class JsonReq {
    Items: JsonItem[] = [];
}

export class JsonResp {
    Items: JsonItem[] = [];
    Message: string = "";
    StatusCode: u16 = 0;
}

// Thank you, @iota/crypto.js, for making my life easy
export class Codec {
    public static bech32Decode(bech32: string): [string, wasmlib.ScAddress, isc.Error] {
        const dec = Bech32.decode(bech32);
        if (dec == undefined) {
            return ["", new wasmlib.ScAddress(), "invalid bech32 string: " + bech32];
        }
        return [dec.humanReadablePart, wasmlib.addressFromBytes(dec.data), null];
    }

    public static bech32Encode(hrp: string, addr: wasmlib.ScAddress): string {
        return Bech32.encode(hrp, wasmlib.addressToBytes(addr));
    }

    public static hNameBytes(name: string): Uint8Array {
        const data = wasmlib.stringToBytes(name)
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
        const enc = new WasmEncoder();
        let items = dict.Items;
        enc.fixedBytes(uint32ToBytes(items.length as u32), ScUint32Length);
        for (let i = 0; i < items.length; i++) {
            const item = items[i];
            const key = hexDecode(item.Key);
            const val = hexDecode(item.Value);
            enc.fixedBytes(uint16ToBytes(key.length as u16), ScUint16Length);
            enc.fixedBytes(key, key.length as u32);
            enc.fixedBytes(uint32ToBytes(val.length as u32), ScUint32Length);
            enc.fixedBytes(val, val.length as u32);
        }
        return enc.buf();
    }

    public static jsonEncode(buf: Uint8Array): JsonReq {
        let dict = new JsonReq();
        const dec = new WasmDecoder(buf);
        const size = uint32FromBytes(dec.fixedBytes(ScUint32Length));
        for (let i: u32 = 0; i < size; i++) {
            const keyBuf = dec.fixedBytes(ScUint16Length);
            const keyLen = uint16FromBytes(keyBuf);
            const key = dec.fixedBytes(keyLen as u32);
            const valBuf = dec.fixedBytes(ScUint32Length);
            const valLen = uint32FromBytes(valBuf);
            const val = dec.fixedBytes(valLen);
            const item = new JsonItem();
            item.Key = hexEncode(key);
            item.Value = hexEncode(val);
            dict.Items.push(item)
        }
        return dict;
    }
}