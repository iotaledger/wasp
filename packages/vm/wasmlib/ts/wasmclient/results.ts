// The Results struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
import * as wasmclient from "./index";
import * as wasmlib from "../wasmlib";
import {Base58} from "./crypto";
import {Buffer} from "./buffer";

export class Results {
    res = new Map<string, wasmclient.Bytes>();

    exists(key: string): wasmclient.Bool {
        return this.res.has(key);
    }

    private get(key: string, typeID: wasmclient.Int32): wasmclient.Bytes {
        let size = wasmlib.TYPE_SIZES[typeID];
        let bytes = this.res.get(key);
        if (bytes !== undefined) {
            if (size != 0 && bytes.length != size) {
                wasmclient.panic("invalid type size");
            }
            return bytes;
        }
        // return default all-zero bytes value
        return Buffer.alloc(size);
    }

    private getBase58(key: string, typeID: wasmclient.Int32): string {
        return Base58.encode(this.get(key, typeID));
    }

    getAddress(key: string): wasmclient.Address {
        return this.getBase58(key, wasmlib.TYPE_ADDRESS);
    }

    getAgentID(key: string): wasmclient.AgentID {
        return this.getBase58(key, wasmlib.TYPE_AGENT_ID);
    }

    getBytes(key: string): wasmclient.Bytes {
        return this.get(key, wasmlib.TYPE_BYTES)
    }

    getBool(key: string): wasmclient.Bool {
        return this.get(key, wasmlib.TYPE_BOOL)[0] != 0;
    }

    getChainID(key: string): wasmclient.ChainID {
        return this.getBase58(key, wasmlib.TYPE_CHAIN_ID);
    }

    getColor(key: string): wasmclient.Color {
        return this.getBase58(key, wasmlib.TYPE_COLOR);
    }

    getHash(key: string): wasmclient.Hash {
        return this.getBase58(key, wasmlib.TYPE_HASH);
    }

    getHname(key: string): wasmclient.Hname {
		return this.get(key, wasmlib.TYPE_HNAME).readUInt32LE(0);
    }

    getInt8(key: string): wasmclient.Int8 {
        return this.get(key, wasmlib.TYPE_INT8).readInt8(0);
    }

    getInt16(key: string): wasmclient.Int16 {
        return this.get(key, wasmlib.TYPE_INT16).readInt16LE(0);
    }

    getInt32(key: string): wasmclient.Int32 {
        return this.get(key, wasmlib.TYPE_INT32).readInt32LE(0);
    }

    getInt64(key: string): wasmclient.Int64 {
        return this.get(key, wasmlib.TYPE_INT64).readBigInt64LE(0);
    }

    getRequestID(key: string): wasmclient.RequestID {
        return this.getBase58(key, wasmlib.TYPE_REQUEST_ID);
    }

    getString(key: string): wasmclient.String {
        return this.get(key, wasmlib.TYPE_STRING).toString();
    }

    getUint8(key: string): wasmclient.Uint8 {
        return this.get(key, wasmlib.TYPE_INT8).readUInt8(0);
    }

    getUint16(key: string): wasmclient.Uint16 {
        return this.get(key, wasmlib.TYPE_INT16).readUInt16LE(0);
    }

    getUint32(key: string): wasmclient.Uint32 {
        return this.get(key, wasmlib.TYPE_INT32).readUInt32LE(0);
    }

    getUint64(key: string): wasmclient.Uint64 {
        return this.get(key, wasmlib.TYPE_INT64).readBigUInt64LE(0);
    }

	// TODO Decode() from view call response into map
}
