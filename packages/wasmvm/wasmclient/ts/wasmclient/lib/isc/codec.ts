// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Bech32, Blake2b } from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';
import * as isc from "./";

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
}