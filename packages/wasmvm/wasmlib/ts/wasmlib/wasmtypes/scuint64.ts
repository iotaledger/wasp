// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import {panic} from '../sandbox';
import {uintFromString64, WasmDecoder, WasmEncoder} from './codec';
import {Proxy} from './proxy';
import {uint32FromBytes} from './scuint32';

export const ScUint64Length = 8;

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export function uint64Decode(dec: WasmDecoder): u64 {
    return dec.vluDecode64(64);
}

export function uint64Encode(enc: WasmEncoder, value: u64): void {
    enc.vluEncode64(value);
}

export function uint64FromBytes(buf: Uint8Array): u64 {
    if (buf.length == 0) {
        return 0n;
    }
    if (buf.length != ScUint64Length) {
        panic('invalid Uint64 length');
    }
    // let ret: u64 = buf[7];
    // ret = (ret << 8) | buf[6];
    // ret = (ret << 8) | buf[5];
    // ret = (ret << 8) | buf[4];
    // ret = (ret << 8) | buf[3];
    // ret = (ret << 8) | buf[2];
    // ret = (ret << 8) | buf[1];
    // return (ret << 8) | buf[0];
    const u32high = uint32FromBytes(buf.subarray(4));
    const u32low = uint32FromBytes(buf.subarray(0, 4));
    return BigInt(u32high) * 0x100000000n + BigInt(u32low);
}

export function uint64ToBytes(value: u64): Uint8Array {
    const buf = new Uint8Array(ScUint64Length);
    const u32low = Number(value % 0x100000000n);
    const u32High = Number(value / 0x100000000n % 0x100000000n);
    buf[0] = u32low as u8;
    buf[1] = (u32low >> 8) as u8;
    buf[2] = (u32low >> 16) as u8;
    buf[3] = (u32low >> 24) as u8;
    buf[4] = u32High as u8;
    buf[5] = (u32High >> 8) as u8;
    buf[6] = (u32High >> 16) as u8;
    buf[7] = (u32High >> 24) as u8;
    return buf;
}

export function uint64FromString(value: string): u64 {
    return uintFromString64(value, 64);
}

export function uint64ToString(value: u64): string {
    return value.toString();
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScImmutableUint64 {
    proxy: Proxy;

    constructor(proxy: Proxy) {
        this.proxy = proxy;
    }

    exists(): bool {
        return this.proxy.exists();
    }

    toString(): string {
        return uint64ToString(this.value());
    }

    value(): u64 {
        return uint64FromBytes(this.proxy.get());
    }
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

export class ScMutableUint64 extends ScImmutableUint64 {
    delete(): void {
        this.proxy.delete();
    }

    setValue(value: u64): void {
        this.proxy.set(uint64ToBytes(value));
    }
}
