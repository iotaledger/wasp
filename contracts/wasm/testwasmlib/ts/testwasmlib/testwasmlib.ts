// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as coreblocklog from "wasmlib/coreblocklog"
import * as sc from "./index";

export function funcArrayAppend(ctx: wasmlib.ScFuncContext, f: sc.ArrayAppendContext): void {
    let name = f.params.name().value();
    let array = f.state.arrays().getStringArray(name);
    let value = f.params.value().value();
    array.appendString().setValue(value);
}

export function funcArrayClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayClearContext): void {
    let name = f.params.name().value();
    let array = f.state.arrays().getStringArray(name);
    array.clear();
}

export function funcArraySet(ctx: wasmlib.ScFuncContext, f: sc.ArraySetContext): void {
    let name = f.params.name().value();
    let array = f.state.arrays().getStringArray(name);
    let index = f.params.index().value();
    let value = f.params.value().value();
    array.getString(index).setValue(value);
}

export function funcMapClear(ctx: wasmlib.ScFuncContext, f: sc.MapClearContext): void {
    let name = f.params.name().value();
    let myMap = f.state.maps().getStringMap(name);
    myMap.clear();
}

export function funcMapSet(ctx: wasmlib.ScFuncContext, f: sc.MapSetContext): void {
    let name = f.params.name().value();
    let myMap = f.state.maps().getStringMap(name);
    let key = f.params.key().value();
    let value = f.params.value().value();
    myMap.getString(key).setValue(value);
}

export function funcParamTypes(ctx: wasmlib.ScFuncContext, f: sc.ParamTypesContext): void {
    if (f.params.address().exists()) {
        ctx.require(f.params.address().value().equals(ctx.accountID().address()), "mismatch: Address");
    }
    if (f.params.agentID().exists()) {
        ctx.require(f.params.agentID().value().equals(ctx.accountID()), "mismatch: AgentID");
    }
    if (f.params.bool().exists()) {
        ctx.require(f.params.bool().value(), "mismatch: Bool");
    }
    if (f.params.bytes().exists()) {
        const byteData = wasmtypes.stringToBytes("these are bytes");
        ctx.require(wasmtypes.bytesCompare(f.params.bytes().value(), byteData) == 0, "mismatch: Bytes");
    }
    if (f.params.chainID().exists()) {
        ctx.require(f.params.chainID().value().equals(ctx.chainID()), "mismatch: ChainID");
    }
    if (f.params.color().exists()) {
        const color = wasmlib.colorFromBytes(wasmtypes.stringToBytes("RedGreenBlueYellowCyanBlackWhitePurple"));
        ctx.require(f.params.color().value().equals(color), "mismatch: Color");
    }
    if (f.params.hash().exists()) {
        const hash = wasmtypes.hashFromBytes(wasmtypes.stringToBytes("0123456789abcdeffedcba9876543210"));
        ctx.require(f.params.hash().value().equals(hash), "mismatch: Hash");
    }
    if (f.params.hname().exists()) {
        ctx.require(f.params.hname().value().equals(ctx.accountID().hname()), "mismatch: Hname");
    }
    if (f.params.int8().exists()) {
        ctx.require(f.params.int8().value() == -123, "mismatch: Int8");
    }
    if (f.params.int16().exists()) {
        ctx.require(f.params.int16().value() == -12345, "mismatch: Int16");
    }
    if (f.params.int32().exists()) {
        ctx.require(f.params.int32().value() == -1234567890, "mismatch: Int32");
    }
    if (f.params.int64().exists()) {
        ctx.require(f.params.int64().value() == -1234567890123456789, "mismatch: Int64");
    }
    if (f.params.requestID().exists()) {
        const requestId = wasmtypes.requestIDFromBytes(wasmtypes.stringToBytes("abcdefghijklmnopqrstuvwxyz123456\x00\x00"));
        ctx.require(f.params.requestID().value().equals(requestId), "mismatch: RequestID");
    }
    if (f.params.string().exists()) {
        ctx.require(f.params.string().value() == "this is a string", "mismatch: String");
    }
    if (f.params.uint8().exists()) {
        ctx.require(f.params.uint8().value() == 123, "mismatch: Uint8");
    }
    if (f.params.uint16().exists()) {
        ctx.require(f.params.uint16().value() == 12345, "mismatch: Uint16");
    }
    if (f.params.uint32().exists()) {
        ctx.require(f.params.uint32().value() == 1234567890, "mismatch: Uint32");
    }
    if (f.params.uint64().exists()) {
        ctx.require(f.params.uint64().value() == 1234567890123456789, "mismatch: Uint64");
    }
}

export function funcRandom(ctx: wasmlib.ScFuncContext, f: sc.RandomContext): void {
    f.state.random().setValue(ctx.random(1000));
}

export function funcTriggerEvent(ctx: wasmlib.ScFuncContext, f: sc.TriggerEventContext): void {
    f.events.test(f.params.address().value(), f.params.name().value());
}

export function viewArrayLength(ctx: wasmlib.ScViewContext, f: sc.ArrayLengthContext): void {
    let name = f.params.name().value();
    let array = f.state.arrays().getStringArray(name);
    let length = array.length();
    f.results.length().setValue(length);
}

export function viewArrayValue(ctx: wasmlib.ScViewContext, f: sc.ArrayValueContext): void {
    let name = f.params.name().value();
    let array = f.state.arrays().getStringArray(name);
    let index = f.params.index().value();
    let value = array.getString(index).value();
    f.results.value().setValue(value);
}

export function viewBlockRecord(ctx: wasmlib.ScViewContext, f: sc.BlockRecordContext): void {
    let records = coreblocklog.ScFuncs.getRequestReceiptsForBlock(ctx);
    records.params.blockIndex().setValue(f.params.blockIndex().value());
    records.func.call();
    let recordIndex = f.params.recordIndex().value();
    ctx.log("index: " + recordIndex.toString());
    recordIndex = f.params.recordIndex().value();
    ctx.log("index: " + recordIndex.toString());
    const requestRecord = records.results.requestRecord();
    const length = requestRecord.length();
    ctx.log("length: " + length.toString());
    const length2 = requestRecord.length();
    ctx.log("length2: " + length2.toString());
    ctx.require(recordIndex < length, "invalid recordIndex");
    const buf = requestRecord.getBytes(recordIndex).value();
    f.results.record().setValue(buf);
}

export function viewBlockRecords(ctx: wasmlib.ScViewContext, f: sc.BlockRecordsContext): void {
    let records = coreblocklog.ScFuncs.getRequestReceiptsForBlock(ctx);
    records.params.blockIndex().setValue(f.params.blockIndex().value());
    records.func.call();
    f.results.count().setValue(records.results.requestRecord().length());
}

export function viewGetRandom(ctx: wasmlib.ScViewContext, f: sc.GetRandomContext): void {
    f.results.random().setValue(f.state.random().value());
}

export function viewIotaBalance(ctx: wasmlib.ScViewContext, f: sc.IotaBalanceContext): void {
    f.results.iotas().setValue(ctx.balances().balance(wasmtypes.IOTA));
}

export function viewMapValue(ctx: wasmlib.ScViewContext, f: sc.MapValueContext): void {
    let name = f.params.name().value();
    let myMap = f.state.maps().getStringMap(name);
    let key = f.params.key().value();
    let value = myMap.getString(key).value();
    f.results.value().setValue(value);
}

export function funcTakeAllowance(ctx: wasmlib.ScFuncContext, f: sc.TakeAllowanceContext): void {
}
