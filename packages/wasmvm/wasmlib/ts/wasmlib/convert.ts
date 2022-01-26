// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// converts to and from little endian bytes
import {panic} from "./host";

export class Convert {
    static equals(lhs: u8[], rhs: u8[]): boolean {
        if (lhs.length != rhs.length) {
            return false;
        }
        for (let i = lhs.length - 1; i >= 0; i--) {
            if (lhs[i] != rhs[i]) {
                return false;
            }
        }
        return true;
    }

    static fromI16(val: i16): u8[] {
        return [
            val as u8,
            (val >> 8) as u8,
        ];
    }

    static fromI32(val: i32): u8[] {
        return [
            val as u8,
            (val >> 8) as u8,
            (val >> 16) as u8,
            (val >> 24) as u8,
        ];
    }

    static fromI64(val: i64): u8[] {
        return [
            val as u8,
            (val >> 8) as u8,
            (val >> 16) as u8,
            (val >> 24) as u8,
            (val >> 32) as u8,
            (val >> 40) as u8,
            (val >> 48) as u8,
            (val >> 56) as u8,
        ];
    }

    static fromString(val: string): u8[] {
        let arrayBuffer = String.UTF8.encode(val);
        let u8Array = Uint8Array.wrap(arrayBuffer)
        let ret: u8[] = new Array(u8Array.length);
        for (let i = 0; i < ret.length; i++) {
            ret[i] = u8Array[i];
        }
        return ret;
    }

    static toI16(bytes: u8[]): i16 {
        if (bytes.length != 2) {
            panic("expected i16 (2 bytes)")
        }

        let ret: i16 = bytes[1];
        return (ret << 8) | bytes[0];
    }

    static toI32(bytes: u8[]): i32 {
        if (bytes.length != 4) {
            panic("expected i32 (4 bytes)")
        }

        let ret: i32 = bytes[3];
        ret = (ret << 8) | bytes[2];
        ret = (ret << 8) | bytes[1];
        return (ret << 8) | bytes[0];
    }

    static toI64(bytes: u8[]): i64 {
        if (bytes.length != 8) {
            panic("expected i64 (8 bytes)")
        }

        let ret: i64 = bytes[7];
        ret = (ret << 8) | bytes[6];
        ret = (ret << 8) | bytes[5];
        ret = (ret << 8) | bytes[4];
        ret = (ret << 8) | bytes[3];
        ret = (ret << 8) | bytes[2];
        ret = (ret << 8) | bytes[1];
        return (ret << 8) | bytes[0];
    }

    static toString(bytes: u8[]): string {
        return String.UTF8.decodeUnsafe(bytes.dataStart, bytes.length);
    }
}