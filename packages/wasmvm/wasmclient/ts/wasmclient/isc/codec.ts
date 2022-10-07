// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as isc from "./index";
import { Bech32,Blake2b } from "@iota/crypto.js"

// Thank you, @iota/crypto.js, for making my life easy
export class Codec {
    public static bech32Prefix: string = "smr";

    public static bech32Decode(bech32: string): isc.Address | null {
        let dec = Bech32.decode(bech32);
        if (dec == undefined) {
            return null;
        }
        return dec.data;
    }

    public static bech32Encode(addr: isc.Address): string {
        return Bech32.encode(Codec.bech32Prefix, addr);
    }

    public static hNameEncode(name: string): isc.Hname {
        const len = name.length;
        const data = new Uint8Array(len);
        for (let i = 0; i < len; i++) {
            data[i] = name.charCodeAt(i);
        }
        let hash = Blake2b.sum256(data)

        // dumb Uint8Array cannot convert easily to u8[]
        // so do it the hard way
        let slice : u8[] = [0, 0, 0, 0];
        for (let i = 0; i < 4; i++) {
            slice[i] = hash[i];
        }

        // follow exact algorithm from packages/isc/hname.go
        let hName = wasmlib.uint32FromBytes(slice);
        if (hName == 0 || hName == 0xffff) {
            for (let i = 0; i < 4; i++) {
                slice[i] = hash[i + 4];
            }
            hName = wasmlib.uint32FromBytes(slice);
         }
        return hName;
    }
}