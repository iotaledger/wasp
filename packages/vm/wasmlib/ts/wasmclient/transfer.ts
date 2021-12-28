// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Base58} from "./crypto";
import {Buffer} from "./buffer";

export class Transfer {
    private xfer = new Map<Buffer, wasmclient.Uint64>();

    static iotas(amount: wasmclient.Uint64): Transfer {
        const iotaColorBase58 = "11111111111111111111111111111111";
        return Transfer.tokens(iotaColorBase58, amount);
    }

    static tokens(color: string, amount: wasmclient.Uint64): Transfer {
        let transfers = new Transfer();
        transfers.set(color, amount);
        return transfers;
    }

    set(color: string, amount: wasmclient.Uint64) {
        this.xfer.set(Base58.decode(color), amount);
    }

    // Encode returns a byte array that encodes the Transfer as follows:
    // Sort all nonzero transfers in ascending color order (very important,
    // because this data will be part of the data that will be signed,
    // so the order needs to be 100% deterministic). Then emit the 4-byte
    // transfer count. Next for each color emit the 32-byte color value,
    // and then the 8-byte amount.
    encode(): wasmclient.Bytes {
        let keys = new Array<Buffer>();
        for (const [key,val] of this.xfer) {
            // filter out zero transfers
            if (val != BigInt(0)) {
                keys.push(key);
            }
        }
        keys.sort((lhs, rhs) => lhs.compare(rhs));

        let buf = Buffer.alloc(0);
        buf.writeUInt32LE(keys.length, 0);
        for (const key of keys) {
            buf = Buffer.concat([buf, key]);
            buf.writeBigUInt64LE(this.xfer.get(key), buf.length);
        }
        return buf;
    }
}