// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmclient from "./index"
import {Base58} from "./crypto";
import {Buffer} from "./buffer";

// The Arguments struct is used to gather all arguments for this smart
// contract function call and encode it into this deterministic byte array
export class Encoder {
    private fromBase58(val: string, typeID: wasmclient.Int32): Buffer {
        const bytes = Base58.decode(val);
        if (bytes.length != wasmclient.TYPE_SIZES[typeID]) {
            wasmclient.panic("invalid byte size");
        }
        return bytes;
    }

    fromAddress(val: wasmclient.AgentID): Buffer {
        return this.fromBase58(val, wasmclient.TYPE_ADDRESS);
    }

    fromAgentID(val: wasmclient.AgentID): Buffer {
        return this.fromBase58(val, wasmclient.TYPE_AGENT_ID);
    }

    fromBool(val: boolean): Buffer {
        const bytes = Buffer.alloc(1);
        if (val) {
            bytes.writeUInt8(1, 0);
        }
        return bytes;
    }

    fromBytes(val: wasmclient.Bytes): Buffer {
        return val;
    }

    fromChainID(val: wasmclient.ChainID): Buffer {
        return this.fromBase58(val, wasmclient.TYPE_CHAIN_ID);
    }

    fromColor(val: wasmclient.Color): Buffer {
        if (val == "IOTA") {
            val = "11111111111111111111111111111111";
        }
        return this.fromBase58(val, wasmclient.TYPE_COLOR);
    }

    fromHash(val: wasmclient.Hash): Buffer {
        return this.fromBase58(val, wasmclient.TYPE_HASH);
    }

    fromHname(val: wasmclient.Hname): Buffer {
        return this.fromUint32(val);
    }

    fromInt8(val: wasmclient.Int8): Buffer {
        const bytes = Buffer.alloc(1);
        bytes.writeInt8(val, 0);
        return bytes;
    }

    fromInt16(val: wasmclient.Int16): Buffer {
        const bytes = Buffer.alloc(2);
        bytes.writeInt16LE(val, 0);
        return bytes;
    }

    fromInt32(val: wasmclient.Int32): Buffer {
        const bytes = Buffer.alloc(4);
        bytes.writeInt32LE(val, 0);
        return bytes;
    }

    fromInt64(val: wasmclient.Int64): Buffer {
        const bytes = Buffer.alloc(8);
        bytes.writeBigInt64LE(val, 0);
        return bytes;
    }

    fromRequestID(val: wasmclient.RequestID): Buffer {
        return this.fromBase58(val, wasmclient.TYPE_REQUEST_ID);
    }

    fromString(val: string): Buffer {
        return Buffer.from(val);
    }

    fromUint8(val: wasmclient.Uint8): Buffer {
        const bytes = Buffer.alloc(1);
        bytes.writeUInt8(val, 0);
        return bytes;
    }

    fromUint16(val: wasmclient.Uint16): Buffer {
        const bytes = Buffer.alloc(2);
        bytes.writeUInt16LE(val, 0);
        return bytes;
    }

    fromUint32(val: wasmclient.Uint32): Buffer {
        const bytes = Buffer.alloc(4);
        bytes.writeUInt32LE(val, 0);
        return bytes;
    }

    fromUint64(val: wasmclient.Uint64): Buffer {
        const bytes = Buffer.alloc(8);
        bytes.writeBigUInt64LE(val, 0);
        return bytes;
    }
}
