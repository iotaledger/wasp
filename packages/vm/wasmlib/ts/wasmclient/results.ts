// The Results struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
import * as wasmclient from "./index";
import {Base58} from "./crypto";
import {Buffer} from "./buffer";

export class ViewResults {
    res: wasmclient.Results;

    constructor(res: wasmclient.Results) {
        this.res = res;
    }
}

export class Results {
    res = new Map<string, wasmclient.Bytes>();
    keys = new Map<string, wasmclient.Bytes>();

    exists(key: string): wasmclient.Bool {
        return this.res.has(key);
    }

    forEach(callbackfn: (base58Key: string, valueKey: string) => void): void {
        this.keys.forEach((val, key, map) => {
            callbackfn(Base58.encode(val), key);
        })
    }

    private get(key: string, typeID: wasmclient.Int32): wasmclient.Bytes {
        const size = wasmclient.TYPE_SIZES[typeID];
        const bytes = this.res.get(key);
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
        return this.getBase58(key, wasmclient.TYPE_ADDRESS);
    }

    getAgentID(key: string): wasmclient.AgentID {
        return this.getBase58(key, wasmclient.TYPE_AGENT_ID);
    }

    getBytes(key: string): wasmclient.Bytes {
        return this.get(key, wasmclient.TYPE_BYTES)
    }

    getBool(key: string): wasmclient.Bool {
        return this.get(key, wasmclient.TYPE_BOOL)[0] != 0;
    }

    getChainID(key: string): wasmclient.ChainID {
        return this.getBase58(key, wasmclient.TYPE_CHAIN_ID);
    }

    getColor(key: string): wasmclient.Color {
        return this.getBase58(key, wasmclient.TYPE_COLOR);
    }

    getHash(key: string): wasmclient.Hash {
        return this.getBase58(key, wasmclient.TYPE_HASH);
    }

    getHname(key: string): wasmclient.Hname {
        return this.get(key, wasmclient.TYPE_HNAME).readUInt32LE(0);
    }

    getInt8(key: string): wasmclient.Int8 {
        return this.get(key, wasmclient.TYPE_INT8).readInt8(0);
    }

    getInt16(key: string): wasmclient.Int16 {
        return this.get(key, wasmclient.TYPE_INT16).readInt16LE(0);
    }

    getInt32(key: string): wasmclient.Int32 {
        return this.get(key, wasmclient.TYPE_INT32).readInt32LE(0);
    }

    getInt64(key: string): wasmclient.Int64 {
        return this.get(key, wasmclient.TYPE_INT64).readBigInt64LE(0);
    }

    getRequestID(key: string): wasmclient.RequestID {
        return this.getBase58(key, wasmclient.TYPE_REQUEST_ID);
    }

    getString(key: string): wasmclient.String {
        return this.get(key, wasmclient.TYPE_STRING).toString();
    }

    getUint8(key: string): wasmclient.Uint8 {
        return this.get(key, wasmclient.TYPE_INT8).readUInt8(0);
    }

    getUint16(key: string): wasmclient.Uint16 {
        return this.get(key, wasmclient.TYPE_INT16).readUInt16LE(0);
    }

    getUint32(key: string): wasmclient.Uint32 {
        return this.get(key, wasmclient.TYPE_INT32).readUInt32LE(0);
    }

    getUint64(key: string): wasmclient.Uint64 {
        return this.get(key, wasmclient.TYPE_INT64).readBigUInt64LE(0);
    }

    // TODO Decode() from view call response into map
}
