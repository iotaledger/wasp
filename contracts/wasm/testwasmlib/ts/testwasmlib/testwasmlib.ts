// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as coreblocklog from "wasmlib/coreblocklog"
import * as sc from "./index";

export function funcParamTypes(ctx: wasmlib.ScFuncContext, f: sc.ParamTypesContext): void {
    if ((f.params.address().exists())) {
        ctx.require(f.params.address().value().equals(ctx.accountID().address()), "mismatch: Address");
    }
    if ((f.params.agentID().exists())) {
        ctx.require(f.params.agentID().value().equals(ctx.accountID()), "mismatch: AgentID");
    }
    if ((f.params.bigInt().exists())) {
        let bigIntData = wasmtypes.bigIntFromString("100000000000000000000");
        ctx.require(f.params.bigInt().value().cmp(bigIntData) == 0, "mismatch: BigInt");
    }
    if ((f.params.bool().exists())) {
        ctx.require(f.params.bool().value(), "mismatch: Bool");
    }
    if ((f.params.bytes().exists())) {
        const byteData = wasmtypes.stringToBytes("these are bytes");
        ctx.require(wasmtypes.bytesCompare(f.params.bytes().value(), byteData) == 0, "mismatch: Bytes");
    }
    if ((f.params.chainID().exists())) {
        ctx.require(f.params.chainID().value().equals(ctx.currentChainID()), "mismatch: ChainID");
    }
    if ((f.params.hash().exists())) {
        const hash = wasmtypes.hashFromBytes(wasmtypes.stringToBytes("0123456789abcdeffedcba9876543210"));
        ctx.require(f.params.hash().value().equals(hash), "mismatch: Hash");
    }
    if ((f.params.hname().exists())) {
        ctx.require(f.params.hname().value().equals(ctx.accountID().hname()), "mismatch: Hname");
    }
    if ((f.params.int8().exists())) {
        ctx.require(f.params.int8().value() == -123, "mismatch: Int8");
    }
    if ((f.params.int16().exists())) {
        ctx.require(f.params.int16().value() == -12345, "mismatch: Int16");
    }
    if ((f.params.int32().exists())) {
        ctx.require(f.params.int32().value() == -1234567890, "mismatch: Int32");
    }
    if ((f.params.int64().exists())) {
        ctx.require(f.params.int64().value() == -1234567890123456789, "mismatch: Int64");
    }
    if ((f.params.nftID().exists())) {
        const color = wasmlib.nftIDFromBytes(wasmtypes.stringToBytes("abcdefghijklmnopqrstuvwxyz123456"));
        ctx.require(f.params.nftID().value().equals(color), "mismatch: NftID");
    }
    if ((f.params.requestID().exists())) {
        const requestId = wasmtypes.requestIDFromBytes(wasmtypes.stringToBytes("abcdefghijklmnopqrstuvwxyz123456\x00\x00"));
        ctx.require(f.params.requestID().value().equals(requestId), "mismatch: RequestID");
    }
    if ((f.params.string().exists())) {
        ctx.require(f.params.string().value() == "this is a string", "mismatch: String");
    }
    if ((f.params.tokenID().exists())) {
        const color = wasmlib.tokenIDFromBytes(wasmtypes.stringToBytes("abcdefghijklmnopqrstuvwxyz1234567890AB"));
        ctx.require(f.params.tokenID().value().equals(color), "mismatch: TokenID");
    }
    if ((f.params.uint8().exists())) {
        ctx.require(f.params.uint8().value() == 123, "mismatch: Uint8");
    }
    if ((f.params.uint16().exists())) {
        ctx.require(f.params.uint16().value() == 12345, "mismatch: Uint16");
    }
    if ((f.params.uint32().exists())) {
        ctx.require(f.params.uint32().value() == 1234567890, "mismatch: Uint32");
    }
    if ((f.params.uint64().exists())) {
        ctx.require(f.params.uint64().value() == 1234567890123456789, "mismatch: Uint64");
    }
}

export function funcRandom(ctx: wasmlib.ScFuncContext, f: sc.RandomContext): void {
    f.state.random().setValue(ctx.random(1000));
}

export function funcTakeAllowance(ctx: wasmlib.ScFuncContext, f: sc.TakeAllowanceContext): void {
    ctx.transferAllowed(ctx.accountID(), wasmlib.ScTransfer.fromBalances(ctx.allowance()), false);
}

export function funcTakeBalance(ctx: wasmlib.ScFuncContext, f: sc.TakeBalanceContext): void {
    f.results.tokens().setValue(ctx.balances().baseTokens());
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

export function viewTokenBalance(ctx: wasmlib.ScViewContext, f: sc.TokenBalanceContext): void {
    f.results.tokens().setValue(ctx.balances().baseTokens());
}

//////////////////// array of StringArray \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfStringArrayAppend(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfStringArrayAppendContext): void {
    const index = f.params.index().value();
    const valLen = f.params.value().length();

    let sa: sc.ArrayOfMutableString;
    if (f.state.arrayOfStringArray().length() <= index) {
        sa = f.state.arrayOfStringArray().appendStringArray();
    } else {
        sa = f.state.arrayOfStringArray().getStringArray(index);
    }

    for (let i = u32(0); i < valLen; i++) {
        const elt = f.params.value().getString(i).value();
        sa.appendString().setValue(elt);
    }
}

export function funcArrayOfStringArrayClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfStringArrayClearContext): void {
    const length = f.state.arrayOfStringArray().length();
    for (let i = u32(0); i < length; i++) {
        const array = f.state.arrayOfStringArray().getStringArray(i);
        array.clear();
    }
    f.state.arrayOfStringArray().clear();
}

export function funcArrayOfStringArraySet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfStringArraySetContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();
    const array = f.state.arrayOfStringArray().getStringArray(index0);
    const value = f.params.value().value();
    array.getString(index1).setValue(value);
}

export function viewArrayOfStringArrayLength(ctx: wasmlib.ScViewContext, f: sc.ArrayOfStringArrayLengthContext): void {
    const length = f.state.arrayOfStringArray().length();
    f.results.length().setValue(length);
}

export function viewArrayOfStringArrayValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfStringArrayValueContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();

    const elt = f.state.arrayOfStringArray().getStringArray(index0).getString(index1).value();
    f.results.value().setValue(elt);
}

//////////////////// array of StringMap \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfStringMapClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfStringMapClearContext): void {
    const length = f.state.arrayOfStringArray().length();
    for (let i = u32(0); i < length; i++) {
        const mmap = f.state.arrayOfStringMap().getStringMap(i);
        mmap.clear();
    }
    f.state.arrayOfStringMap().clear();
}

export function funcArrayOfStringMapSet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfStringMapSetContext): void {
    const index = f.params.index().value();
    const value = f.params.value().value();
    const key = f.params.key().value();
    if (f.state.arrayOfStringMap().length() <= index) {
        const mmap = f.state.arrayOfStringMap().appendStringMap();
        mmap.getString(key).setValue(value);
        return
    }
    const mmap = f.state.arrayOfStringMap().getStringMap(index);
    mmap.getString(key).setValue(value);
}

export function viewArrayOfStringMapValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfStringMapValueContext): void {
    const index = f.params.index().value();
    const key = f.params.key().value();
    const mmap = f.state.arrayOfStringMap().getStringMap(index);
    f.results.value().setValue(mmap.getString(key).value());
}

//////////////////// StringMap of StringArray \\\\\\\\\\\\\\\\\\\\

export function funcStringMapOfStringArrayAppend(ctx: wasmlib.ScFuncContext, f: sc.StringMapOfStringArrayAppendContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfStringArray().getStringArray(name);
    const value = f.params.value().value();
    array.appendString().setValue(value);
}

export function funcStringMapOfStringArrayClear(ctx: wasmlib.ScFuncContext, f: sc.StringMapOfStringArrayClearContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfStringArray().getStringArray(name);
    array.clear();
}

export function funcStringMapOfStringArraySet(ctx: wasmlib.ScFuncContext, f: sc.StringMapOfStringArraySetContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfStringArray().getStringArray(name);
    const index = f.params.index().value();
    const value = f.params.value().value();
    array.getString(index).setValue(value);
}

export function viewStringMapOfStringArrayLength(ctx: wasmlib.ScViewContext, f: sc.StringMapOfStringArrayLengthContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfStringArray().getStringArray(name);
    const length = array.length();
    f.results.length().setValue(length);
}

export function viewStringMapOfStringArrayValue(ctx: wasmlib.ScViewContext, f: sc.StringMapOfStringArrayValueContext): void {
    const name = f.params.name().value();
    const array = f.state.stringMapOfStringArray().getStringArray(name);
    const index = f.params.index().value();
    const value = array.getString(index).value();
    f.results.value().setValue(value);
}

//////////////////// StringMap of StringMap \\\\\\\\\\\\\\\\\\\\

export function funcStringMapOfStringMapClear(ctx: wasmlib.ScFuncContext, f: sc.StringMapOfStringMapClearContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfStringMap().getStringMap(name);
    mmap.clear();
}

export function funcStringMapOfStringMapSet(ctx: wasmlib.ScFuncContext, f: sc.StringMapOfStringMapSetContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfStringMap().getStringMap(name);
    const key = f.params.key().value();
    const value = f.params.value().value();
    mmap.getString(key).setValue(value);
}

export function viewStringMapOfStringMapValue(ctx: wasmlib.ScViewContext, f: sc.StringMapOfStringMapValueContext): void {
    const name = f.params.name().value();
    const mmap = f.state.stringMapOfStringMap().getStringMap(name);
    const key = f.params.key().value();
    f.results.value().setValue(mmap.getString(key).value());
}

//////////////////// array of AddressArray \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfAddressArrayAppend(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfAddressArrayAppendContext): void {
    const index = f.params.index().value();
    const valLen = f.params.valueAddr().length();

    let sa: sc.ArrayOfMutableAddress;
    if (f.state.arrayOfStringArray().length() <= index) {
        sa = f.state.arrayOfAddressArray().appendAddressArray();
    } else {
        sa = f.state.arrayOfAddressArray().getAddressArray(index);
    }

    for (let i = u32(0); i < valLen; i++) {
        const elt = f.params.valueAddr().getAddress(i).value();
        sa.appendAddress().setValue(elt);
    }
}

export function funcArrayOfAddressArrayClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfAddressArrayClearContext): void {
    const length = f.state.arrayOfAddressArray().length();
    for (let i = u32(0); i < length; i++) {
        const array = f.state.arrayOfAddressArray().getAddressArray(i);
        array.clear();
    }
    f.state.arrayOfAddressArray().clear();
}

export function funcArrayOfAddressArraySet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfAddressArraySetContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();
    const array = f.state.arrayOfAddressArray().getAddressArray(index0);
    const value = f.params.valueAddr().value();
    array.getAddress(index1).setValue(value);
}

export function viewArrayOfAddressArrayLength(ctx: wasmlib.ScViewContext, f: sc.ArrayOfAddressArrayLengthContext): void {
    const length = f.state.arrayOfAddressArray().length();
    f.results.length().setValue(length);
}

export function viewArrayOfAddressArrayValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfAddressArrayValueContext): void {
    const index0 = f.params.index0().value();
    const index1 = f.params.index1().value();

    const elt = f.state.arrayOfAddressArray().getAddressArray(index0).getAddress(index1).value();
    f.results.valueAddr().setValue(elt);
}

//////////////////// array of AddressMap \\\\\\\\\\\\\\\\\\\\

export function funcArrayOfAddressMapClear(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfAddressMapClearContext): void {
    const length = f.state.arrayOfAddressArray().length();
    for (let i = u32(0); i < length; i++) {
        const mmap = f.state.arrayOfAddressMap().getAddressMap(i);
        mmap.clear();
    }
    f.state.arrayOfAddressMap().clear();
}

export function funcArrayOfAddressMapSet(ctx: wasmlib.ScFuncContext, f: sc.ArrayOfAddressMapSetContext): void {
    const index = f.params.index().value();
    const value = f.params.valueAddr().value();
    const key = f.params.keyAddr().value();
    if (f.state.arrayOfAddressMap().length() <= index) {
        const mmap = f.state.arrayOfAddressMap().appendAddressMap();
        mmap.getAddress(key).setValue(value);
        return
    }
    const mmap = f.state.arrayOfAddressMap().getAddressMap(index);
    mmap.getAddress(key).setValue(value);
}

export function viewArrayOfAddressMapValue(ctx: wasmlib.ScViewContext, f: sc.ArrayOfAddressMapValueContext): void {
    const index = f.params.index().value();
    const key = f.params.keyAddr().value();
    const mmap = f.state.arrayOfAddressMap().getAddressMap(index);
    f.results.valueAddr().setValue(mmap.getAddress(key).value());
}

//////////////////// AddressMap of AddressArray \\\\\\\\\\\\\\\\\\\\

export function funcAddressMapOfAddressArrayAppend(ctx: wasmlib.ScFuncContext, f: sc.AddressMapOfAddressArrayAppendContext): void {
    const addr = f.params.nameAddr().value();
    const array = f.state.addressMapOfAddressArray().getAddressArray(addr);
    const value = f.params.valueAddr().value();
    array.appendAddress().setValue(value);
}

export function funcAddressMapOfAddressArrayClear(ctx: wasmlib.ScFuncContext, f: sc.AddressMapOfAddressArrayClearContext): void {
    const addr = f.params.nameAddr().value();
    const array = f.state.addressMapOfAddressArray().getAddressArray(addr);
    array.clear();
}

export function funcAddressMapOfAddressArraySet(ctx: wasmlib.ScFuncContext, f: sc.AddressMapOfAddressArraySetContext): void {
    const addr = f.params.nameAddr().value();
    const array = f.state.addressMapOfAddressArray().getAddressArray(addr);
    const index = f.params.index().value();
    const value = f.params.valueAddr().value();
    array.getAddress(index).setValue(value);
}

export function viewAddressMapOfAddressArrayLength(ctx: wasmlib.ScViewContext, f: sc.AddressMapOfAddressArrayLengthContext): void {
    const addr = f.params.nameAddr().value();
    const array = f.state.addressMapOfAddressArray().getAddressArray(addr);
    const length = array.length();
    f.results.length().setValue(length);
}

export function viewAddressMapOfAddressArrayValue(ctx: wasmlib.ScViewContext, f: sc.AddressMapOfAddressArrayValueContext): void {
    const addr = f.params.nameAddr().value();
    const array = f.state.addressMapOfAddressArray().getAddressArray(addr);
    const index = f.params.index().value();
    const value = array.getAddress(index).value();
    f.results.valueAddr().setValue(value);
}

//////////////////// AddressMap of AddressMap \\\\\\\\\\\\\\\\\\\\

export function funcAddressMapOfAddressMapClear(ctx: wasmlib.ScFuncContext, f: sc.AddressMapOfAddressMapClearContext): void {
    const name = f.params.nameAddr().value();
    const myMap = f.state.addressMapOfAddressMap().getAddressMap(name);
    myMap.clear();
}

export function funcAddressMapOfAddressMapSet(ctx: wasmlib.ScFuncContext, f: sc.AddressMapOfAddressMapSetContext): void {
    const name = f.params.nameAddr().value();
    const myMap = f.state.addressMapOfAddressMap().getAddressMap(name);
    const key = f.params.keyAddr().value();
    const value = f.params.valueAddr().value();
    myMap.getAddress(key).setValue(value);
}

export function viewAddressMapOfAddressMapValue(ctx: wasmlib.ScViewContext, f: sc.AddressMapOfAddressMapValueContext): void {
    const name = f.params.nameAddr().value();
    const myMap = f.state.addressMapOfAddressMap().getAddressMap(name);
    const key = f.params.keyAddr().value();
    f.results.valueAddr().setValue(myMap.getAddress(key).value());
}

export function viewBigIntAdd(ctx: wasmlib.ScViewContext, f: sc.BigIntAddContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.add(rhs);
    f.results.res().setValue(res);
}

export function viewBigIntDiv(ctx: wasmlib.ScViewContext, f: sc.BigIntDivContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.div(rhs);
    f.results.res().setValue(res);
}

export function viewBigIntDivMod(ctx: wasmlib.ScViewContext, f: sc.BigIntDivModContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.divMod(rhs);
    f.results.quo().setValue(res[0]);
    f.results.remainder().setValue(res[1]);
}

export function viewBigIntMod(ctx: wasmlib.ScViewContext, f: sc.BigIntModContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.modulo(rhs);
    f.results.res().setValue(res);
}

export function viewBigIntMul(ctx: wasmlib.ScViewContext, f: sc.BigIntMulContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.mul(rhs);
    f.results.res().setValue(res);
}

export function viewBigIntSub(ctx: wasmlib.ScViewContext, f: sc.BigIntSubContext): void {
    const lhs = f.params.lhs().value();
    const rhs = f.params.rhs().value();
    const res = lhs.sub(rhs);
    f.results.res().setValue(res);
}

export function viewBigIntShl(ctx: wasmlib.ScViewContext, f: sc.BigIntShlContext): void {
    const lhs = f.params.lhs().value();
    const shift = f.params.shift().value();
    const res = lhs.shl(shift);
    f.results.res().setValue(res);
}

export function viewBigIntShr(ctx: wasmlib.ScViewContext, f: sc.BigIntShrContext): void {
    const lhs = f.params.lhs().value();
    const shift = f.params.shift().value();
    const res = lhs.shr(shift);
    f.results.res().setValue(res);
}

export function viewCheckAgentID(ctx: wasmlib.ScViewContext, f: sc.CheckAgentIDContext): void {
    const scAgentID = f.params.scAgentID().value();
    const agentBytes = f.params.agentBytes().value();
    const agentString = f.params.agentString().value();
    ctx.require(scAgentID.equals(wasmtypes.agentIDFromBytes(wasmtypes.agentIDToBytes(scAgentID))), "bytes conversion failed");
    ctx.require(scAgentID.equals(wasmtypes.agentIDFromString(wasmtypes.agentIDToString(scAgentID))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(scAgentID.toBytes(), agentBytes) == 0, "bytes mismatch");
    ctx.require(scAgentID.toString() == agentString, "string mismatch");
}

export function viewCheckAddress(ctx: wasmlib.ScViewContext, f: sc.CheckAddressContext): void {
    const address = f.params.scAddress().value();
    const addressBytes = f.params.addressBytes().value();
    const addressString = f.params.addressString().value();
    ctx.require(address.equals(wasmtypes.addressFromBytes(wasmtypes.addressToBytes(address))), "bytes conversion failed");
    ctx.require(address.equals(wasmtypes.addressFromString(wasmtypes.addressToString(address))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(address.toBytes(), addressBytes) == 0, "bytes mismatch");
    ctx.require(address.toString() == addressString, "string mismatch");
}

export function viewCheckEthAddressAndAgentID(ctx: wasmlib.ScViewContext, f: sc.CheckEthAddressAndAgentIDContext): void {
    const ethAddress = f.params.ethAddress().value();
    const scAddressEth = wasmtypes.addressFromString(ethAddress);
    ctx.require(scAddressEth.equals(wasmtypes.addressFromBytes(wasmtypes.addressToBytes(scAddressEth))), "eth address bytes conversion failed");
    ctx.require(scAddressEth.equals(wasmtypes.addressFromString(wasmtypes.addressToString(scAddressEth))), "eth address string conversion failed");
    const scAgentIDEth = wasmtypes.ScAgentID.fromAddress(scAddressEth);
    ctx.require(scAgentIDEth.equals(wasmtypes.agentIDFromBytes(wasmtypes.agentIDToBytes(scAgentIDEth))), "eth agentID bytes conversion failed");
    ctx.require(scAgentIDEth.equals(wasmtypes.agentIDFromString(wasmtypes.agentIDToString(scAgentIDEth))), "eth agentID string conversion failed");
}

export function viewCheckHash(ctx: wasmlib.ScViewContext, f: sc.CheckHashContext): void {
    const hash = f.params.scHash().value();
    const hashBytes = f.params.hashBytes().value();
    const hashString = f.params.hashString().value();
    ctx.require(hash.equals(wasmtypes.hashFromBytes(wasmtypes.hashToBytes(hash))), "bytes conversion failed");
    ctx.require(hash.equals(wasmtypes.hashFromString(wasmtypes.hashToString(hash))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(hash.toBytes(), hashBytes) == 0, "bytes mismatch");
    ctx.require(hash.toString() == hashString, "string mismatch");
}

export function viewCheckNftID(ctx: wasmlib.ScViewContext, f: sc.CheckNftIDContext): void {
    const nftID = f.params.scNftID().value();
    const nftIDBytes = f.params.nftIDBytes().value();
    const nftIDString = f.params.nftIDString().value();
    ctx.require(nftID.equals(wasmtypes.nftIDFromBytes(wasmtypes.nftIDToBytes(nftID))), "bytes conversion failed");
    ctx.require(nftID.equals(wasmtypes.nftIDFromString(wasmtypes.nftIDToString(nftID))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(nftID.toBytes(), nftIDBytes) == 0, "bytes mismatch");
    ctx.require(nftID.toString() == nftIDString, "string mismatch");
}

export function viewCheckRequestID(ctx: wasmlib.ScViewContext, f: sc.CheckRequestIDContext): void {
    const RequestID = f.params.scRequestID().value();
    const RequestIDBytes = f.params.requestIDBytes().value();
    const RequestIDString = f.params.requestIDString().value();
    ctx.require(RequestID.equals(wasmtypes.requestIDFromBytes(wasmtypes.requestIDToBytes(RequestID))), "bytes conversion failed");
    ctx.require(RequestID.equals(wasmtypes.requestIDFromString(wasmtypes.requestIDToString(RequestID))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(RequestID.toBytes(), RequestIDBytes) == 0, "bytes mismatch");
    ctx.require(RequestID.toString() == RequestIDString, "string mismatch");
}

export function viewCheckTokenID(ctx: wasmlib.ScViewContext, f: sc.CheckTokenIDContext): void {
    const tokenID = f.params.scTokenID().value();
    const tokenIDBytes = f.params.tokenIDBytes().value();
    const tokenIDString = f.params.tokenIDString().value();
    ctx.require(tokenID.equals(wasmtypes.tokenIDFromBytes(wasmtypes.tokenIDToBytes(tokenID))), "bytes conversion failed");
    ctx.require(tokenID.equals(wasmtypes.tokenIDFromString(wasmtypes.tokenIDToString(tokenID))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(tokenID.toBytes(), tokenIDBytes) == 0, "bytes mismatch");
    ctx.require(tokenID.toString() == tokenIDString, "string mismatch");
}

export function viewCheckBigInt(ctx: wasmlib.ScViewContext, f: sc.CheckBigIntContext): void {
    const bigInt = f.params.scBigInt().value();
    const bigIntBytes = f.params.bigIntBytes().value();
    const bigIntString = f.params.bigIntString().value();
    ctx.require(bigInt.equals(wasmtypes.bigIntFromBytes(wasmtypes.bigIntToBytes(bigInt))), "bytes conversion failed");
    ctx.require(bigInt.equals(wasmtypes.bigIntFromString(wasmtypes.bigIntToString(bigInt))), "string conversion failed");
    ctx.require(wasmtypes.bytesCompare(bigInt.toBytes(), bigIntBytes) == 0, "bytes mismatch");
    ctx.require(bigInt.toString() == bigIntString, "string mismatch");
}

export function viewCheckIntAndUint(ctx: wasmlib.ScViewContext, f: sc.CheckIntAndUintContext): void {
    let int8 = i8.MAX_VALUE;
	ctx.require(int8 == wasmtypes.int8FromBytes(wasmtypes.int8ToBytes(int8)), "bytes conversion failed");
	ctx.require(int8 == wasmtypes.int8FromString(wasmtypes.int8ToString(int8)), "string conversion failed");
	int8 = i8.MIN_VALUE;
	ctx.require(int8 == wasmtypes.int8FromBytes(wasmtypes.int8ToBytes(int8)), "bytes conversion failed");
	ctx.require(int8 == wasmtypes.int8FromString(wasmtypes.int8ToString(int8)), "string conversion failed");
	int8 = 1;
	ctx.require(int8 == wasmtypes.int8FromBytes(wasmtypes.int8ToBytes(int8)), "bytes conversion failed");
	ctx.require(int8 == wasmtypes.int8FromString(wasmtypes.int8ToString(int8)), "string conversion failed");
	int8 = 0;
	ctx.require(int8 == wasmtypes.int8FromBytes(wasmtypes.int8ToBytes(int8)), "bytes conversion failed");
	ctx.require(int8 == wasmtypes.int8FromString(wasmtypes.int8ToString(int8)), "string conversion failed");
	int8 = -1;
	ctx.require(int8 == wasmtypes.int8FromBytes(wasmtypes.int8ToBytes(int8)), "bytes conversion failed");
	ctx.require(int8 == wasmtypes.int8FromString(wasmtypes.int8ToString(int8)), "string conversion failed");
	let uint8 = u8.MIN_VALUE;
	ctx.require(uint8 == wasmtypes.uint8FromBytes(wasmtypes.uint8ToBytes(uint8)), "bytes conversion failed");
	ctx.require(uint8 == wasmtypes.uint8FromString(wasmtypes.uint8ToString(uint8)), "string conversion failed");
    uint8--;
    ctx.require(uint8 == u8.MAX_VALUE, "unxpected max u8")
	ctx.require(uint8 == wasmtypes.uint8FromBytes(wasmtypes.uint8ToBytes(uint8)), "bytes conversion failed");
	ctx.require(uint8 == wasmtypes.uint8FromString(wasmtypes.uint8ToString(uint8)), "string conversion failed");

    let int16 = i16.MAX_VALUE;
	ctx.require(int16 == wasmtypes.int16FromBytes(wasmtypes.int16ToBytes(int16)), "bytes conversion failed");
	ctx.require(int16 == wasmtypes.int16FromString(wasmtypes.int16ToString(int16)), "string conversion failed");
	int16 = i16.MIN_VALUE;
	ctx.require(int16 == wasmtypes.int16FromBytes(wasmtypes.int16ToBytes(int16)), "bytes conversion failed");
	ctx.require(int16 == wasmtypes.int16FromString(wasmtypes.int16ToString(int16)), "string conversion failed");
	int16 = 1;
	ctx.require(int16 == wasmtypes.int16FromBytes(wasmtypes.int16ToBytes(int16)), "bytes conversion failed");
	ctx.require(int16 == wasmtypes.int16FromString(wasmtypes.int16ToString(int16)), "string conversion failed");
	int16 = 0;
	ctx.require(int16 == wasmtypes.int16FromBytes(wasmtypes.int16ToBytes(int16)), "bytes conversion failed");
	ctx.require(int16 == wasmtypes.int16FromString(wasmtypes.int16ToString(int16)), "string conversion failed");
	int16 = -1;
	ctx.require(int16 == wasmtypes.int16FromBytes(wasmtypes.int16ToBytes(int16)), "bytes conversion failed");
	ctx.require(int16 == wasmtypes.int16FromString(wasmtypes.int16ToString(int16)), "string conversion failed");
	let uint16 = u16.MIN_VALUE;
	ctx.require(uint16 == wasmtypes.uint16FromBytes(wasmtypes.uint16ToBytes(uint16)), "bytes conversion failed");
	ctx.require(uint16 == wasmtypes.uint16FromString(wasmtypes.uint16ToString(uint16)), "string conversion failed");
    uint16--;
    ctx.require(uint16 == u16.MAX_VALUE, "unxpected max u16")
	ctx.require(uint16 == wasmtypes.uint16FromBytes(wasmtypes.uint16ToBytes(uint16)), "bytes conversion failed");
	ctx.require(uint16 == wasmtypes.uint16FromString(wasmtypes.uint16ToString(uint16)), "string conversion failed");

	let int32 = i32.MAX_VALUE;
	ctx.require(int32 == wasmtypes.int32FromBytes(wasmtypes.int32ToBytes(int32)), "bytes conversion failed");
	ctx.require(int32 == wasmtypes.int32FromString(wasmtypes.int32ToString(int32)), "string conversion failed");
	int32 = i32.MIN_VALUE;
	ctx.require(int32 == wasmtypes.int32FromBytes(wasmtypes.int32ToBytes(int32)), "bytes conversion failed");
	ctx.require(int32 == wasmtypes.int32FromString(wasmtypes.int32ToString(int32)), "string conversion failed");
	int32 = 1;
	ctx.require(int32 == wasmtypes.int32FromBytes(wasmtypes.int32ToBytes(int32)), "bytes conversion failed");
	ctx.require(int32 == wasmtypes.int32FromString(wasmtypes.int32ToString(int32)), "string conversion failed");
	int32 = 0;
	ctx.require(int32 == wasmtypes.int32FromBytes(wasmtypes.int32ToBytes(int32)), "bytes conversion failed");
	ctx.require(int32 == wasmtypes.int32FromString(wasmtypes.int32ToString(int32)), "string conversion failed");
	int32 = -1;
	ctx.require(int32 == wasmtypes.int32FromBytes(wasmtypes.int32ToBytes(int32)), "bytes conversion failed");
	ctx.require(int32 == wasmtypes.int32FromString(wasmtypes.int32ToString(int32)), "string conversion failed");
	let uint32 = u32.MIN_VALUE;
	ctx.require(uint32 == wasmtypes.uint32FromBytes(wasmtypes.uint32ToBytes(uint32)), "bytes conversion failed");
	ctx.require(uint32 == wasmtypes.uint32FromString(wasmtypes.uint32ToString(uint32)), "string conversion failed");
    uint32--;
    ctx.require(uint32 == u32.MAX_VALUE, "unxpected max u32")
	ctx.require(uint32 == wasmtypes.uint32FromBytes(wasmtypes.uint32ToBytes(uint32)), "bytes conversion failed");
	ctx.require(uint32 == wasmtypes.uint32FromString(wasmtypes.uint32ToString(uint32)), "string conversion failed");

	let int64 = i64.MAX_VALUE;
	ctx.require(int64 == wasmtypes.int64FromBytes(wasmtypes.int64ToBytes(int64)), "bytes conversion failed");
	ctx.require(int64 == wasmtypes.int64FromString(wasmtypes.int64ToString(int64)), "string conversion failed");
	int64 = i64.MIN_VALUE;
	ctx.require(int64 == wasmtypes.int64FromBytes(wasmtypes.int64ToBytes(int64)), "bytes conversion failed");
	ctx.require(int64 == wasmtypes.int64FromString(wasmtypes.int64ToString(int64)), "string conversion failed");
	int64 = 1;
	ctx.require(int64 == wasmtypes.int64FromBytes(wasmtypes.int64ToBytes(int64)), "bytes conversion failed");
	ctx.require(int64 == wasmtypes.int64FromString(wasmtypes.int64ToString(int64)), "string conversion failed");
	int64 = 0;
	ctx.require(int64 == wasmtypes.int64FromBytes(wasmtypes.int64ToBytes(int64)), "bytes conversion failed");
	ctx.require(int64 == wasmtypes.int64FromString(wasmtypes.int64ToString(int64)), "string conversion failed");
	int64 = -1;
	ctx.require(int64 == wasmtypes.int64FromBytes(wasmtypes.int64ToBytes(int64)), "bytes conversion failed");
	ctx.require(int64 == wasmtypes.int64FromString(wasmtypes.int64ToString(int64)), "string conversion failed");
	let uint64 = u64.MIN_VALUE;
	ctx.require(uint64 == wasmtypes.uint64FromBytes(wasmtypes.uint64ToBytes(uint64)), "bytes conversion failed");
	ctx.require(uint64 == wasmtypes.uint64FromString(wasmtypes.uint64ToString(uint64)), "string conversion failed");
	uint64--;
    ctx.require(uint64 == u64.MAX_VALUE, "unxpected max u64")
	ctx.require(uint64 == wasmtypes.uint64FromBytes(wasmtypes.uint64ToBytes(uint64)), "bytes conversion failed");
	ctx.require(uint64 == wasmtypes.uint64FromString(wasmtypes.uint64ToString(uint64)), "string conversion failed");
}

export function viewCheckBool(ctx: wasmlib.ScViewContext, f: sc.CheckBoolContext): void {
	ctx.require(wasmtypes.boolFromBytes(wasmtypes.boolToBytes(true)), "bytes conversion failed");
	ctx.require(wasmtypes.boolFromString(wasmtypes.boolToString(true)), "string conversion failed");
	ctx.require(!wasmtypes.boolFromBytes(wasmtypes.boolToBytes(false)), "bytes conversion failed");
	ctx.require(!wasmtypes.boolFromString(wasmtypes.boolToString(false)), "string conversion failed");
}

export function viewCheckBytes(ctx: wasmlib.ScViewContext, f: sc.CheckBytesContext): void {
    let byteData = f.params.bytes().value();
	ctx.require(wasmtypes.bytesCompare(byteData, wasmtypes.bytesFromBytes(wasmtypes.bytesToBytes(byteData))) == 0, "bytes conversion failed");
	ctx.require(wasmtypes.bytesCompare(byteData, wasmtypes.bytesFromString(wasmtypes.bytesToString(byteData))) == 0, "string conversion failed");
}

export function viewCheckHname(ctx: wasmlib.ScViewContext, f: sc.CheckHnameContext): void {
    let scHname = f.params.scHname().value();
	let hnameBytes = f.params.hnameBytes().value();
	let hnameString = f.params.hnameString().value();
	ctx.require(scHname.equals(wasmtypes.hnameFromBytes(wasmtypes.hnameToBytes(scHname))), "bytes conversion failed");
	ctx.require(scHname.equals(wasmtypes.hnameFromString(wasmtypes.hnameToString(scHname))), "string conversion failed");
	ctx.require(wasmtypes.bytesCompare(hnameBytes, wasmtypes.hnameToBytes(scHname)) == 0, "not equal to input bytes");
	ctx.require(hnameString == wasmtypes.hnameToString(scHname), "not equal to input string");
}

export function viewCheckString(ctx: wasmlib.ScViewContext, f: sc.CheckStringContext): void {
    let stringData = f.params.string().value();
	ctx.require(stringData == wasmtypes.stringFromBytes(wasmtypes.stringToBytes(stringData)), "bytes conversion failed");
	ctx.require(stringData == wasmtypes.stringToString(wasmtypes.stringFromString(stringData)), "string conversion failed");
}
