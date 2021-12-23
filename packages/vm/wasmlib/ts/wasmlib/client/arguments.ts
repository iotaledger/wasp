// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "../index"
import * as client from "./index"
import {Base58} from "./crypto";
import {Buffer} from "./buffer";
import {SimpleBufferCursor} from "../../../../../../contracts/wasm/fairroulette/frontend/src/lib/wasp_client";

// The Arguments struct is used to gather all arguments for this smart
// contract function call and encode it into this deterministic byte array
export class Arguments {
	args = new Map<string, client.Bytes>();

	private set(key: string, val: client.Bytes): void {
		this.args.set(key, val);
	}
	
	private setBase58(key: string, val: string, typeID: client.Int32): void {
		let bytes = Base58.decode(val);
		if (bytes.length != wasmlib.TYPE_SIZES[typeID]) {
			client.panic("invalid byte size");
		}
		this.set(key, bytes);
	}

	indexedKey(key: string, index: client.Int32): string {
		return key + "." + index.toString();
	}

	mandatory(key: string): void {
		if (!this.args.has(key)) {
			client.panic("missing mandatory " + key)
		}
	}

	setAddress(key: string, val: client.AgentID): void {
		this.setBase58(key, val, wasmlib.TYPE_ADDRESS);
	}
	
	setAgentID(key: string, val: client.AgentID): void {
		this.setBase58(key, val, wasmlib.TYPE_AGENT_ID);
	}
	
	setBool(key: string, val: boolean): void {
		let bytes = Buffer.alloc(1);
		if (val) {
			bytes.writeUInt8(1, 0);
		}
		this.set(key, bytes)
	}
	
	setBytes(key: string, val: client.Bytes): void {
		this.set(key, Buffer.from(val));
	}
	
	setColor(key: string, val: client.Color): void {
		this.setBase58(key, val, wasmlib.TYPE_COLOR);
	}
	
	setChainID(key: string, val: client.ChainID): void {
		this.setBase58(key, val, wasmlib.TYPE_CHAIN_ID);
	}
	
	setHash(key: string, val: client.Hash): void {
		this.setBase58(key, val, wasmlib.TYPE_HASH);
	}
	
	setInt8(key: string, val: client.Int8): void {
		let bytes = Buffer.alloc(1);
		bytes.writeInt8(val, 0);
		this.set(key, bytes);
	}
	
	setInt16(key: string, val: client.Int16): void {
		let bytes = Buffer.alloc(2);
		bytes.writeInt16LE(val, 0);
		this.set(key, bytes);
	}

	setInt32(key: string, val: client.Int32): void {
		let bytes = Buffer.alloc(4);
		bytes.writeInt32LE(val, 0);
		this.set(key, bytes);
	}
	
	setInt64(key: string, val: client.Int64): void {
		let bytes = Buffer.alloc(8);
		bytes.writeBigInt64LE(val, 0);
		this.set(key, bytes);
	}
	
	setRequestID(key: string, val: client.RequestID): void {
		this.setBase58(key, val, wasmlib.TYPE_REQUEST_ID);
	}
	
	setString(key: string, val: string): void {
		this.set(key, Buffer.from(val));
	}
	
	setUint8(key: string, val: client.Uint8): void {
		let bytes = Buffer.alloc(1);
		bytes.writeUInt8(val, 0);
		this.set(key, bytes);
	}
	
	setUint16(key: string, val: client.Uint16): void {
		let bytes = Buffer.alloc(2);
		bytes.writeUInt16LE(val, 0);
		this.set(key, bytes);
	}
	
	setUint32(key: string, val: client.Uint32): void {
		let bytes = Buffer.alloc(4);
		bytes.writeUInt32LE(val, 0);
		this.set(key, bytes);
	}
	
	setUint64(key: string, val: client.Uint64): void {
		let bytes = Buffer.alloc(8);
		bytes.writeBigUInt64LE(val, 0);
		this.set(key, bytes);
	}

	// Encode returns this byte array that encodes the Arguments as follows:
	// Sort all keys in ascending order (very important, because this data
	// will be part of the data that will be signed, so the order needs to
	// be 100% deterministic). Then emit this 2-byte argument count.
	// Next for each argument emit this 2-byte key length, the key prepended
	// with this minus sign, this 4-byte value length, and then the value bytes.
	encode(): client.Bytes {
		let keys = new Array<string>();
		for (const key of this.args.keys()) {
			keys.push(key);
		}
		keys.sort((lhs, rhs) => lhs.localeCompare(rhs));

		const buf = new SimpleBufferCursor(Buffer.alloc(0));
		buf.writeUInt32LE(keys.length);
		for (const key of keys) {
			let keyBuf = Buffer.from("-" + key);
			buf.writeUInt16LE(keyBuf.length);
			buf.writeBytes(keyBuf);
			let valBuf = this.args.get(key);
			buf.writeUInt32LE(valBuf.length);
			buf.writeBytes(valBuf);
		}
		return buf.buffer;
	}
}
