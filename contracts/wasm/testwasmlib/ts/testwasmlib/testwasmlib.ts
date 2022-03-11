// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as coreblocklog from "wasmlib/coreblocklog"
import * as sc from "./index";

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
        const color = wasmlib.colorFromBytes(wasmtypes.stringToBytes("RedGreenBlueYellowCyanBlackWhite"));
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

//////////////////// array of array \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfArraysAppend(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfArraysAppendContext): void {
    const index = f.params.index().value();
    const length = f.params.value().length();

    let sa: sc.ArrayOfMutableString;
    if (f.state.stringArrayOfArrays().length() <= index) {
        sa = f.state.stringArrayOfArrays().appendStringArray();
    } else {
        sa = f.state.stringArrayOfArrays().getStringArray(index);
    }

    for (let i = u32(0); i < length; i++) {
        const elt = f.params.value().getString(i).value();
        sa.appendString().setValue(elt);
    }
}

export function funcArrayOfArraysClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfArraysClearContext): void {
    const length = f.state.stringArrayOfArrays().length();
    for (let i = u32(0); i < length; i++) {
        const array = f.state.stringArrayOfArrays().getStringArray(i);
        array.clear();
    }
    f.state.stringArrayOfArrays().clear();
}

export function funcArrayOfArraysSet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfArraysSetContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();
    const array = f.state.stringArrayOfArrays().getStringArray(index0);
    const value = f.params.value().value();
    array.getString(index1).setValue(value);
}

export function viewArrayOfArraysLength(ctx: wasmlib.ScViewContext, f: sc.ArrayOfArraysLengthContext): void {
    const length = f.state.stringArrayOfArrays().length();
    f.results.length().setValue(length);
}

export function viewArrayOfArraysValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfArraysValueContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();

    const elt = f.state.stringArrayOfArrays().getStringArray(index0).getString(index1).value();
    f.results.value().setValue(elt);
}

//////////////////// array of map \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfMapsClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfMapsClearContext): void {
    const length = f.state.stringArrayOfArrays().length();
    for (let i = u32(0); i < length; i++) {
        const mmap = f.state.stringArrayOfMaps().getStringMap(i);
        mmap.clear();
    }
    f.state.stringArrayOfMaps().clear();
}

export function funcArrayOfMapsSet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfMapsSetContext): void {
    const index = f.params.index().value();
    const value = f.params.value().value();
    const key = f.params.key().value();
    if (f.state.stringArrayOfMaps().length() <= index) {
        const mmap = f.state.stringArrayOfMaps().appendStringMap();
        mmap.getString(key).setValue(value);
        return
    }
    const mmap = f.state.stringArrayOfMaps().getStringMap(index);
    mmap.getString(key).setValue(value);
}

export function viewArrayOfMapsValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfMapsValueContext): void {
    const index = f.params.index().value();
    const key = f.params.key().value();
    const mmap = f.state.stringArrayOfMaps().getStringMap(index);
    f.results.value().setValue(mmap.getString(key).value());
}

//////////////////// map of array \\\\\\\\\\\\\\\\\\\\

export function funcMapOfArraysAppend(ctx: wasmlib.ScFuncContext, f: sc.MapOfArraysAppendContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfArrays().getStringArray(name);
    const value = f.params.value().value();
    array.appendString().setValue(value);
}

export function funcMapOfArraysClear(ctx: wasmlib.ScFuncContext, f: sc.MapOfArraysClearContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfArrays().getStringArray(name);
    array.clear();
}

export function funcMapOfArraysSet(ctx: wasmlib.ScFuncContext, f: sc.MapOfArraysSetContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfArrays().getStringArray(name);
    const index = f.params.index().value();
    const value = f.params.value().value();
    array.getString(index).setValue(value);
}

export function viewMapOfArraysLength(ctx: wasmlib.ScViewContext, f: sc.MapOfArraysLengthContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfArrays().getStringArray(name);
    const length = array.length();
    f.results.length().setValue(length);
}

export function viewMapOfArraysValue(ctx: wasmlib.ScViewContext, f: sc.MapOfArraysValueContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfArrays().getStringArray(name);
    const index = f.params.index().value();
    const value = array.getString(index).value();
    f.results.value().setValue(value);
}

//////////////////// map of map \\\\\\\\\\\\\\\\\\\\

export function funcMapOfMapsClear(ctx: wasmlib.ScFuncContext, f: sc.MapOfMapsClearContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfMaps().getStringMap(name);
    mmap.clear();
}

export function funcMapOfMapsSet(ctx: wasmlib.ScFuncContext, f: sc.MapOfMapsSetContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfMaps().getStringMap(name);
    const key = f.params.key().value();
    const value = f.params.value().value();
    mmap.getString(key).setValue(value);
}

export function viewMapOfMapsValue(ctx: wasmlib.ScViewContext, f: sc.MapOfMapsValueContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfMaps().getStringMap(name);
    const key = f.params.key().value();
    f.results.value().setValue(mmap.getString(key).value());
}
