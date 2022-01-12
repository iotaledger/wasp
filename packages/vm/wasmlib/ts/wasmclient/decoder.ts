// The Results struct is used to gather all arguments for a smart
// contract function call and encode it into a deterministic byte array
import * as wasmclient from "./index";
import {Base58} from "./crypto";
import {Buffer} from "./buffer";

export class Decoder {
    private static checkDefault(bytes: Buffer | undefined, typeID: wasmclient.Int32): Buffer {
        const size = wasmclient.TYPE_SIZES[typeID];
        if (bytes === undefined) {
            return Buffer.alloc(size);
        }
        if (size != 0 && bytes.length != size) {
            wasmclient.panic("invalid type size");
        }
        return bytes;
    }

    private static toBase58(bytes: Buffer | undefined, typeID: wasmclient.Int32): string {
        return Base58.encode(Decoder.checkDefault(bytes, typeID));
    }

    protected toAddress(bytes: Buffer | undefined): wasmclient.Address {
        return Decoder.toBase58(bytes, wasmclient.TYPE_ADDRESS);
    }

    protected toAgentID(bytes: Buffer | undefined): wasmclient.AgentID {
        return Decoder.toBase58(bytes, wasmclient.TYPE_AGENT_ID);
    }

    protected toBytes(bytes: Buffer | undefined): wasmclient.Bytes {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_BYTES)
    }

    protected toBool(bytes: Buffer | undefined): wasmclient.Bool {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_BOOL)[0] != 0;
    }

    protected toChainID(bytes: Buffer | undefined): wasmclient.ChainID {
        return Decoder.toBase58(bytes, wasmclient.TYPE_CHAIN_ID);
    }

    protected toColor(bytes: Buffer | undefined): wasmclient.Color {
        let color = Decoder.toBase58(bytes, wasmclient.TYPE_COLOR);
        if (color == "11111111111111111111111111111111") {
            color = "IOTA";
        }
        return color;
    }

    protected toHash(bytes: Buffer | undefined): wasmclient.Hash {
        return Decoder.toBase58(bytes, wasmclient.TYPE_HASH);
    }

    protected toHname(bytes: Buffer | undefined): wasmclient.Hname {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_HNAME).readUInt32LE(0);
    }

    protected toInt8(bytes: Buffer | undefined): wasmclient.Int8 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT8).readInt8(0);
    }

    protected toInt16(bytes: Buffer | undefined): wasmclient.Int16 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT16).readInt16LE(0);
    }

    protected toInt32(bytes: Buffer | undefined): wasmclient.Int32 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT32).readInt32LE(0);
    }

    protected toInt64(bytes: Buffer | undefined): wasmclient.Int64 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT64).readBigInt64LE(0);
    }

    protected toRequestID(bytes: Buffer | undefined): wasmclient.RequestID {
        return Decoder.toBase58(bytes, wasmclient.TYPE_REQUEST_ID);
    }

    protected toString(bytes: Buffer | undefined): wasmclient.String {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_STRING).toString();
    }

    protected toUint8(bytes: Buffer | undefined): wasmclient.Uint8 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT8).readUInt8(0);
    }

    protected toUint16(bytes: Buffer | undefined): wasmclient.Uint16 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT16).readUInt16LE(0);
    }

    protected toUint32(bytes: Buffer | undefined): wasmclient.Uint32 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT32).readUInt32LE(0);
    }

    protected toUint64(bytes: Buffer | undefined): wasmclient.Uint64 {
        return Decoder.checkDefault(bytes, wasmclient.TYPE_INT64).readBigUInt64LE(0);
    }
}
