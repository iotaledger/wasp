import * as wasmlib from '../index';

function vliCheck(i: i64) {
    let enc = new wasmlib.WasmEncoder();
    wasmlib.int64Encode(enc, i);
    let buf = enc.buf();
    let dec = new wasmlib.WasmDecoder(buf);
    let v = wasmlib.int64Decode(dec);
    expect(i == v).toBeTruthy();
}

function vluCheck(i: u64) {
    let enc = new wasmlib.WasmEncoder();
    wasmlib.uint64Encode(enc, i);
    let buf = enc.buf();
    let dec = new wasmlib.WasmDecoder(buf);
    let v = wasmlib.uint64Decode(dec);
    expect(i == v).toBeTruthy();
}

function checkString(testValue: string) {
    const buf = wasmlib.stringToBytes(testValue);
    expect(wasmlib.stringFromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.stringToString(testValue);
    expect(wasmlib.stringFromString(str) == testValue).toBeTruthy();
}

function checkInt8(testValue: i8) {
    const buf = wasmlib.int8ToBytes(testValue);
    expect(wasmlib.int8FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.int8ToString(testValue);
    expect(wasmlib.int8FromString(str) == testValue).toBeTruthy();
}

function checkInt16(testValue: i16) {
    const buf = wasmlib.int16ToBytes(testValue);
    expect(wasmlib.int16FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.int16ToString(testValue);
    expect(wasmlib.int16FromString(str) == testValue).toBeTruthy();
}

function checkInt32(testValue: i32) {
    const buf = wasmlib.int32ToBytes(testValue);
    expect(wasmlib.int32FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.int32ToString(testValue);
    expect(wasmlib.int32FromString(str) == testValue).toBeTruthy();
}

function checkInt64(testValue: i64) {
    const buf = wasmlib.int64ToBytes(testValue);
    expect(wasmlib.int64FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.int64ToString(testValue);
    expect(wasmlib.int64FromString(str) == testValue).toBeTruthy();
}

function checkUint8(testValue: u8) {
    const buf = wasmlib.uint8ToBytes(testValue);
    expect(wasmlib.uint8FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.uint8ToString(testValue);
    expect(wasmlib.uint8FromString(str) == testValue).toBeTruthy();
}

function checkUint16(testValue: u16) {
    const buf = wasmlib.uint16ToBytes(testValue);
    expect(wasmlib.uint16FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.uint16ToString(testValue);
    expect(wasmlib.uint16FromString(str) == testValue).toBeTruthy();
}

function checkUint32(testValue: u32) {
    const buf = wasmlib.uint32ToBytes(testValue);
    expect(wasmlib.uint32FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.uint32ToString(testValue);
    expect(wasmlib.uint32FromString(str) == testValue).toBeTruthy();
}

function checkUint64(testValue: u64) {
    const buf = wasmlib.uint64ToBytes(testValue);
    expect(wasmlib.uint64FromBytes(buf) == testValue).toBeTruthy();
    const str = wasmlib.uint64ToString(testValue);
    expect(wasmlib.uint64FromString(str) == testValue).toBeTruthy();
}

describe('conversions', function () {
    it('string conversion', () => {
        checkString("");
        checkString("?");
        checkString("Some weird test string");
    });
    it('int8 conversion', () => {
        checkInt8(0x7f);
        checkInt8(0x7e);
        checkInt8(123);
        checkInt8(1);
        checkInt8(0);
        checkInt8(-1);
        checkInt8(-123);
        checkInt8(-0x7f);
        checkInt8(-0x80);
    });
    it('int16 conversion', () => {
        checkInt16(0x7fff);
        checkInt16(0x7ffe);
        checkInt16(12345);
        checkInt16(1);
        checkInt16(0);
        checkInt16(-1);
        checkInt16(-12345);
        checkInt16(-0x7fff);
        checkInt16(-0x8000);
    });
    it('int32 conversion', () => {
        checkInt32(0x7fffffff);
        checkInt32(0x7ffffffe);
        checkInt32(123_456_789);
        checkInt32(1);
        checkInt32(0);
        checkInt32(-1);
        checkInt32(-123_456_789);
        checkInt32(-0x7fffffff);
        checkInt32(-0x80000000);
    });
    it('int64 conversion', () => {
        checkInt64(0x7fffffffffffffffn);
        checkInt64(0x7ffffffffffffffen);
        checkInt64(123_456_789_123_456_789n);
        checkInt64(1n);
        checkInt64(0n);
        checkInt64(-1n);
        checkInt64(-123_456_789_123_456_789n);
        checkInt64(-0x7fffffffffffffffn);
        checkInt64(-0x8000000000000000n);
    });
    it('uint8 conversion', () => {
        checkUint8(0);
        checkUint8(1);
        checkUint8(123);
        checkUint8(0xfe);
        checkUint8(0xff);
    });
    it('uint16 conversion', () => {
        checkUint16(0);
        checkUint16(1);
        checkUint16(12345);
        checkUint16(0xfffe);
        checkUint16(0xffff);
    });
    it('uint32 conversion', () => {
        checkUint32(0);
        checkUint32(1);
        checkUint32(123_456_789);
        checkUint32(0xfffffffe);
        checkUint32(0xffffffff);
    });
    it('uint64 conversion', () => {
        checkUint64(0n);
        checkUint64(1n);
        checkUint64(123_456_789_123_456_789n);
        checkUint64(0xfffffffffffffffen);
        checkUint64(0xffffffffffffffffn);
    });
    it('WasmCodec vli', () => {
        vliCheck(0x7fffffffffffffffn);
        vliCheck(0x7ffffffffffffffen);
        vliCheck(123_456_789_123_456_789n);
        vliCheck(1n);
        vliCheck(0n);
        vliCheck(-1n);
        vliCheck(-123_456_789_123_456_789n);
        vliCheck(-0x7fffffffffffffffn);
        vliCheck(-0x8000000000000000n);
        for (let i: i64 = -1600n; i < 1600n; i++) {
            vliCheck(i);
        }
        for (let i: i64 = 1n; i <= 1_000_000_000_000_000_000n; i *= 10n) {
            vliCheck(i);
        }
        for (let i: i64 = -1n; i >= -1_000_000_000_000_000_000n; i *= 10n) {
            vliCheck(i);
        }
        for (let i: i64 = 1n; i <= 1_000_000_000_000_000_000n; i *= 2n) {
            vliCheck(i);
        }
        for (let i: i64 = 1n; i <= 1_000_000_000_000_000_000n; i = i * 2n + 1n) {
            vliCheck(i);
        }
        for (let i: i64 = -1n; i >= -1_000_000_000_000_000_000n; i *= 2n) {
            vliCheck(i);
        }
        for (let i: i64 = -1n; i >= -1_000_000_000_000_000_000n; i = i * 2n - 1n) {
            vliCheck(i);
        }
    });
    it('WasmCodec vlu', () => {
        vluCheck(0n);
        vluCheck(1n);
        vluCheck(123_456_789_123_456_789n);
        vluCheck(0xfffffffffffffffen);
        vluCheck(0xffffffffffffffffn);
        for (let i: u64 = 0n; i < 3200n; i++) {
            vluCheck(i);
        }
        for (let i: u64 = 1n; i <= 1_000_000_000_000_000_000n; i *= 10n) {
            vluCheck(i);
        }
        for (let i: u64 = 1n; i <= 1_000_000_000_000_000_000n; i *= 2n) {
            vluCheck(i);
        }
        for (let i: u64 = 1n; i <= 1_000_000_000_000_000_000n; i = i * 2n + 1n) {
            vluCheck(i);
        }
    });
});
