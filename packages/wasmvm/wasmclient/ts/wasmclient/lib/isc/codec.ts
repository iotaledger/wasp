// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import { Bech32, Blake2b } from '@iota/crypto.js';
import * as wasmlib from 'wasmlib';

// Thank you, @iota/crypto.js, for making my life easy
export class Codec {
    //TODO get this from Hornat node config
    public static bech32Prefix = 'smr';

    public static bech32Decode(bech32: string): wasmlib.ScAddress | null {
        const dec = Bech32.decode(bech32);
        if (dec == undefined) {
            return null;
        }
        const buf = wasmlib.bytesFromUint8Array(dec.data);
        return wasmlib.addressFromBytes(buf);
    }

    public static bech32Encode(addr: wasmlib.ScAddress): string {
        const buf = wasmlib.addressToBytes(addr);
        const iscAddr = wasmlib.bytesToUint8Array(buf);
        return Bech32.encode(Codec.bech32Prefix, iscAddr);
    }

    public static hNameBytes(name: string): Uint8Array {
        const data = wasmlib.stringToBytes(name)
        const hash = Blake2b.sum256(data);

        // follow exact algorithm from packages/isc/hname.go
        let slice = wasmlib.bytesFromUint8Array(hash.slice(0, 4));
        const hName = wasmlib.uint32FromBytes(slice);
        if (hName == 0 || hName == 0xffff) {
            slice = wasmlib.bytesFromUint8Array(hash.slice(4, 8));
        }
        return slice;
    }
}