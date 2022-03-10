// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_param_types(ctx: &ScFuncContext, f: &ParamTypesContext) {
    if f.params.address().exists() {
        ctx.require(f.params.address().value() == ctx.account_id().address(), "mismatch: Address");
    }
    if f.params.agent_id().exists() {
        ctx.require(f.params.agent_id().value() == ctx.account_id(), "mismatch: AgentID");
    }
    if f.params.bool().exists() {
        ctx.require(f.params.bool().value(), "mismatch: Bool");
    }
    if f.params.bytes().exists() {
        let byte_data = "these are bytes".as_bytes();
        ctx.require(f.params.bytes().value() == byte_data, "mismatch: Bytes");
    }
    if f.params.chain_id().exists() {
        ctx.require(f.params.chain_id().value() == ctx.chain_id(), "mismatch: ChainID");
    }
    if f.params.color().exists() {
        let color = color_from_bytes("RedGreenBlueYellowCyanBlackWhitePurple".as_bytes());
        ctx.require(f.params.color().value() == color, "mismatch: Color");
    }
    if f.params.hash().exists() {
        let hash = hash_from_bytes("0123456789abcdeffedcba9876543210".as_bytes());
        ctx.require(f.params.hash().value() == hash, "mismatch: Hash");
    }
    if f.params.hname().exists() {
        ctx.require(f.params.hname().value() == ctx.account_id().hname(), "mismatch: Hname");
    }
    if f.params.int8().exists() {
        ctx.require(f.params.int8().value() == -123, "mismatch: Int8");
    }
    if f.params.int16().exists() {
        ctx.require(f.params.int16().value() == -12345, "mismatch: Int16");
    }
    if f.params.int32().exists() {
        ctx.require(f.params.int32().value() == -1234567890, "mismatch: Int32");
    }
    if f.params.int64().exists() {
        ctx.require(f.params.int64().value() == -1234567890123456789, "mismatch: Int64");
    }
    if f.params.request_id().exists() {
        let request_id = request_id_from_bytes("abcdefghijklmnopqrstuvwxyz123456\x00\x00".as_bytes());
        ctx.require(f.params.request_id().value() == request_id, "mismatch: RequestID");
    }
    if f.params.string().exists() {
        ctx.require(f.params.string().value() == "this is a string", "mismatch: String");
    }
    if f.params.uint8().exists() {
        ctx.require(f.params.uint8().value() == 123, "mismatch: Uint8");
    }
    if f.params.uint16().exists() {
        ctx.require(f.params.uint16().value() == 12345, "mismatch: Uint16");
    }
    if f.params.uint32().exists() {
        ctx.require(f.params.uint32().value() == 1234567890, "mismatch: Uint32");
    }
    if f.params.uint64().exists() {
        ctx.require(f.params.uint64().value() == 1234567890123456789, "mismatch: Uint64");
    }
}

pub fn func_random(ctx: &ScFuncContext, f: &RandomContext) {
    f.state.random().set_value(ctx.random(1000));
}

pub fn func_take_allowance(ctx: &ScFuncContext, _f: &TakeAllowanceContext) {
    ctx.transfer_allowed(&ctx.account_id(), &ScTransfers::from_balances(ctx.allowance()), false);
}

pub fn func_take_balance(ctx: &ScFuncContext, f: &TakeBalanceContext) {
    f.results.iotas().set_value(ctx.balances().balance(&ScColor::IOTA));
}

pub fn func_trigger_event(_ctx: &ScFuncContext, f: &TriggerEventContext) {
    f.events.test(&f.params.address().value(), &f.params.name().value());
}

pub fn view_block_record(ctx: &ScViewContext, f: &BlockRecordContext) {
    let records = coreblocklog::ScFuncs::get_request_receipts_for_block(ctx);
    records.params.block_index().set_value(f.params.block_index().value());
    records.func.call();
    let record_index = f.params.record_index().value();
    ctx.require(record_index < records.results.request_record().length(), "invalid recordIndex");
    f.results.record().set_value(&records.results.request_record().get_bytes(record_index).value());
}

pub fn view_block_records(ctx: &ScViewContext, f: &BlockRecordsContext) {
    let records = coreblocklog::ScFuncs::get_request_receipts_for_block(ctx);
    records.params.block_index().set_value(f.params.block_index().value());
    records.func.call();
    f.results.count().set_value(records.results.request_record().length());
}

pub fn view_get_random(_ctx: &ScViewContext, f: &GetRandomContext) {
    f.results.random().set_value(f.state.random().value());
}

pub fn view_iota_balance(ctx: &ScViewContext, f: &IotaBalanceContext) {
    f.results.iotas().set_value(ctx.balances().balance(&ScColor::IOTA));
}

//////////////////// array of array \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_arrays_append(_ctx: &ScFuncContext, f: &ArrayOfArraysAppendContext) {
    let index = f.params.index().value();
    let length = f.params.value().length();

    let sa: ArrayOfMutableString;
    if f.state.string_array_of_arrays().length() <= index {
        sa = f.state.string_array_of_arrays().append_string_array();
    } else {
        sa = f.state.string_array_of_arrays().get_string_array(index);
    }

    for i in 0..length {
        let elt = f.params.value().get_string(i).value();
        sa.append_string().set_value(&elt);
    }
}

pub fn func_array_of_arrays_clear(_ctx: &ScFuncContext, f: &ArrayOfArraysClearContext) {
    let length = f.state.string_array_of_arrays().length();
    for i in 0..length {
        let array = f.state.string_array_of_arrays().get_string_array(i);
        array.clear();
    }
    f.state.string_array_of_arrays().clear();
}

pub fn func_array_of_arrays_set(_ctx: &ScFuncContext, f: &ArrayOfArraysSetContext) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();
    let array = f.state.string_array_of_arrays().get_string_array(index0);
    let value = f.params.value().value();
    array.get_string(index1).set_value(&value);
}

pub fn view_array_of_arrays_length(_ctx: &ScViewContext, f: &ArrayOfArraysLengthContext) {
    let length = f.state.string_array_of_arrays().length();
    f.results.length().set_value(length);
}

pub fn view_array_of_arrays_value(_ctx: &ScViewContext, f: &ArrayOfArraysValueContext) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();

    let elt = f.state.string_array_of_arrays().get_string_array(index0).get_string(index1).value();
    f.results.value().set_value(&elt);
}

//////////////////// array of map \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_maps_clear(_ctx: &ScFuncContext, f: &ArrayOfMapsClearContext) {
    let length = f.state.string_array_of_arrays().length();
    for i in 0..length {
        let mmap = f.state.string_array_of_maps().get_string_map(i);
        mmap.clear();
    }
    f.state.string_array_of_maps().clear();
}

pub fn func_array_of_maps_set(_ctx: &ScFuncContext, f: &ArrayOfMapsSetContext) {
    let index = f.params.index().value();
    let value = f.params.value().value();
    let key = f.params.key().value();
    if f.state.string_array_of_maps().length() <= index {
        let mmap = f.state.string_array_of_maps().append_string_map();
        mmap.get_string(&key).set_value(&value);
        return
    }
    let mmap = f.state.string_array_of_maps().get_string_map(index);
    mmap.get_string(&key).set_value(&value);
}

pub fn view_array_of_maps_value(_ctx: &ScViewContext, f: &ArrayOfMapsValueContext) {
    let index = f.params.index().value();
    let key = f.params.key().value();
    let mmap = f.state.string_array_of_maps().get_string_map(index);
    f.results.value().set_value(&mmap.get_string(&key).value());
}

//////////////////// map of array \\\\\\\\\\\\\\\\\\\\

pub fn func_map_of_arrays_append(_ctx: &ScFuncContext, f: &MapOfArraysAppendContext) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_arrays().get_string_array(&name);
    let value = f.params.value().value();
    array.append_string().set_value(&value);
}

pub fn func_map_of_arrays_clear(_ctx: &ScFuncContext, f: &MapOfArraysClearContext) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_arrays().get_string_array(&name);
    array.clear();
}

pub fn func_map_of_arrays_set(_ctx: &ScFuncContext, f: &MapOfArraysSetContext) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_arrays().get_string_array(&name);
    let index = f.params.index().value();
    let value = f.params.value().value();
    array.get_string(index).set_value(&value);
}

pub fn view_map_of_arrays_length(_ctx: &ScViewContext, f: &MapOfArraysLengthContext) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_arrays().get_string_array(&name);
    let length = array.length();
    f.results.length().set_value(length);
}

pub fn view_map_of_arrays_value(_ctx: &ScViewContext, f: &MapOfArraysValueContext) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_arrays().get_string_array(&name);
    let index = f.params.index().value();
    let value = array.get_string(index).value();
    f.results.value().set_value(&value);
}

//////////////////// map of map \\\\\\\\\\\\\\\\\\\\

pub fn func_map_of_maps_clear(_ctx: &ScFuncContext, f: &MapOfMapsClearContext) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_maps().get_string_map(&name);
    mmap.clear();
}

pub fn func_map_of_maps_set(_ctx: &ScFuncContext, f: &MapOfMapsSetContext) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_maps().get_string_map(&name);
    let key = f.params.key().value();
    let value = f.params.value().value();
    mmap.get_string(&key).set_value(&value);
}

pub fn view_map_of_maps_value(_ctx: &ScViewContext, f: &MapOfMapsValueContext) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_maps().get_string_map(&name);
    let key = f.params.key().value();
    f.results.value().set_value(&mmap.get_string(&key).value());
}
