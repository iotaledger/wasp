// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {Bech32, Blake2b} from "@iota/crypto.js"
import * as isc from "./index";
import * as wasmlib from "wasmlib"

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

    public static hNameBytes(name: string): u8[] {
        const data = Uint8Array.wrap(String.UTF8.encode(name));
        let hash = Blake2b.sum256(data)

        // follow exact algorithm from packages/isc/hname.go
        let slice = wasmlib.bytesFromUint8Array(hash.slice(0, 4));
        let hName = wasmlib.uint32FromBytes(slice);
        if (hName == 0 || hName == 0xffff) {
            slice = wasmlib.bytesFromUint8Array(hash.slice(4, 8));
        }
        return slice;
    }
}