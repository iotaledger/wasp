// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_param_types(ctx: &ScFuncContext, f: &ParamTypesContext) {
    if f.params.address().exists() {
        ctx.require(
            f.params.address().value() == ctx.account_id().address(),
            "mismatch: Address",
        );
    }
    if f.params.agent_id().exists() {
        ctx.require(
            f.params.agent_id().value() == ctx.account_id(),
            "mismatch: AgentID",
        );
    }
    if f.params.bool().exists() {
        ctx.require(f.params.bool().value(), "mismatch: Bool");
    }
    if f.params.bytes().exists() {
        let byte_data = "these are bytes".as_bytes();
        ctx.require(f.params.bytes().value() == byte_data, "mismatch: Bytes");
    }
    if f.params.chain_id().exists() {
        ctx.require(
            f.params.chain_id().value() == ctx.current_chain_id(),
            "mismatch: ChainID",
        );
    }
    if f.params.hash().exists() {
        let hash = hash_from_bytes("0123456789abcdeffedcba9876543210".as_bytes());
        ctx.require(f.params.hash().value() == hash, "mismatch: Hash");
    }
    if f.params.hname().exists() {
        ctx.require(
            f.params.hname().value() == ctx.account_id().hname(),
            "mismatch: Hname",
        );
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
        ctx.require(
            f.params.int64().value() == -1234567890123456789,
            "mismatch: Int64",
        );
    }
    if f.params.nft_id().exists() {
        let nft_id = nft_id_from_bytes("abcdefghijklmnopqrstuvwxyz123456".as_bytes());
        ctx.require(f.params.nft_id().value() == nft_id, "mismatch: NftID");
    }
    if f.params.request_id().exists() {
        let request_id =
            request_id_from_bytes("abcdefghijklmnopqrstuvwxyz123456\x00\x00".as_bytes());
        ctx.require(
            f.params.request_id().value() == request_id,
            "mismatch: RequestID",
        );
    }
    if f.params.string().exists() {
        ctx.require(
            f.params.string().value() == "this is a string",
            "mismatch: String",
        );
    }
    if f.params.token_id().exists() {
        let token_id = token_id_from_bytes("abcdefghijklmnopqrstuvwxyz1234567890AB".as_bytes());
        ctx.require(f.params.token_id().value() == token_id, "mismatch: TokenID");
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
        ctx.require(
            f.params.uint64().value() == 1234567890123456789,
            "mismatch: Uint64",
        );
    }
}

pub fn func_random(ctx: &ScFuncContext, f: &RandomContext) {
    f.state.random().set_value(ctx.random(1000));
}

pub fn func_take_allowance(ctx: &ScFuncContext, _f: &TakeAllowanceContext) {
    ctx.transfer_allowed(
        &ctx.account_id(),
        &ScTransfer::from_balances(&ctx.allowance()),
        false,
    );
}

pub fn func_take_balance(ctx: &ScFuncContext, f: &TakeBalanceContext) {
    f.results.iotas().set_value(ctx.balances().iotas());
}

pub fn func_trigger_event(_ctx: &ScFuncContext, f: &TriggerEventContext) {
    f.events
        .test(&f.params.address().value(), &f.params.name().value());
}

pub fn view_block_record(ctx: &ScViewContext, f: &BlockRecordContext) {
    let records = coreblocklog::ScFuncs::get_request_receipts_for_block(ctx);
    records
        .params
        .block_index()
        .set_value(f.params.block_index().value());
    records.func.call();
    let record_index = f.params.record_index().value();
    ctx.require(
        record_index < records.results.request_record().length(),
        "invalid recordIndex",
    );
    f.results.record().set_value(
        &records
            .results
            .request_record()
            .get_bytes(record_index)
            .value(),
    );
}

pub fn view_block_records(ctx: &ScViewContext, f: &BlockRecordsContext) {
    let records = coreblocklog::ScFuncs::get_request_receipts_for_block(ctx);
    records
        .params
        .block_index()
        .set_value(f.params.block_index().value());
    records.func.call();
    f.results
        .count()
        .set_value(records.results.request_record().length());
}

pub fn view_get_random(_ctx: &ScViewContext, f: &GetRandomContext) {
    f.results.random().set_value(f.state.random().value());
}

pub fn view_iota_balance(ctx: &ScViewContext, f: &IotaBalanceContext) {
    f.results.iotas().set_value(ctx.balances().iotas());
}

//////////////////// array of StringArray \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_string_array_append(
    _ctx: &ScFuncContext,
    f: &ArrayOfStringArrayAppendContext,
) {
    let index = f.params.index().value();
    let val_len = f.params.value().length();

    let sa: ArrayOfMutableString;
    if f.state.array_of_string_array().length() <= index {
        sa = f.state.array_of_string_array().append_string_array();
    } else {
        sa = f.state.array_of_string_array().get_string_array(index);
    }

    for i in 0..val_len {
        let elt = f.params.value().get_string(i).value();
        sa.append_string().set_value(&elt);
    }
}

pub fn func_array_of_string_array_clear(_ctx: &ScFuncContext, f: &ArrayOfStringArrayClearContext) {
    let length = f.state.array_of_string_array().length();
    for i in 0..length {
        let array = f.state.array_of_string_array().get_string_array(i);
        array.clear();
    }
    f.state.array_of_string_array().clear();
}

pub fn func_array_of_string_array_set(_ctx: &ScFuncContext, f: &ArrayOfStringArraySetContext) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();
    let array = f.state.array_of_string_array().get_string_array(index0);
    let value = f.params.value().value();
    array.get_string(index1).set_value(&value);
}

pub fn view_array_of_string_array_length(
    _ctx: &ScViewContext,
    f: &ArrayOfStringArrayLengthContext,
) {
    let length = f.state.array_of_string_array().length();
    f.results.length().set_value(length);
}

pub fn view_array_of_string_array_value(_ctx: &ScViewContext, f: &ArrayOfStringArrayValueContext) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();

    let elt = f
        .state
        .array_of_string_array()
        .get_string_array(index0)
        .get_string(index1)
        .value();
    f.results.value().set_value(&elt);
}

//////////////////// array of StringMap \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_string_map_clear(_ctx: &ScFuncContext, f: &ArrayOfStringMapClearContext) {
    let length = f.state.array_of_string_array().length();
    for i in 0..length {
        let mmap = f.state.array_of_string_map().get_string_map(i);
        mmap.clear();
    }
    f.state.array_of_string_map().clear();
}

pub fn func_array_of_string_map_set(_ctx: &ScFuncContext, f: &ArrayOfStringMapSetContext) {
    let index = f.params.index().value();
    let value = f.params.value().value();
    let key = f.params.key().value();
    if f.state.array_of_string_map().length() <= index {
        let mmap = f.state.array_of_string_map().append_string_map();
        mmap.get_string(&key).set_value(&value);
        return;
    }
    let mmap = f.state.array_of_string_map().get_string_map(index);
    mmap.get_string(&key).set_value(&value);
}

pub fn view_array_of_string_map_value(_ctx: &ScViewContext, f: &ArrayOfStringMapValueContext) {
    let index = f.params.index().value();
    let key = f.params.key().value();
    let mmap = f.state.array_of_string_map().get_string_map(index);
    f.results.value().set_value(&mmap.get_string(&key).value());
}

//////////////////// StringMap of StringArray \\\\\\\\\\\\\\\\\\\\

pub fn func_string_map_of_string_array_append(
    _ctx: &ScFuncContext,
    f: &StringMapOfStringArrayAppendContext,
) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_string_array().get_string_array(&name);
    let value = f.params.value().value();
    array.append_string().set_value(&value);
}

pub fn func_string_map_of_string_array_clear(
    _ctx: &ScFuncContext,
    f: &StringMapOfStringArrayClearContext,
) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_string_array().get_string_array(&name);
    array.clear();
}

pub fn func_string_map_of_string_array_set(
    _ctx: &ScFuncContext,
    f: &StringMapOfStringArraySetContext,
) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_string_array().get_string_array(&name);
    let index = f.params.index().value();
    let value = f.params.value().value();
    array.get_string(index).set_value(&value);
}

pub fn view_string_map_of_string_array_length(
    _ctx: &ScViewContext,
    f: &StringMapOfStringArrayLengthContext,
) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_string_array().get_string_array(&name);
    let length = array.length();
    f.results.length().set_value(length);
}

pub fn view_string_map_of_string_array_value(
    _ctx: &ScViewContext,
    f: &StringMapOfStringArrayValueContext,
) {
    let name = f.params.name().value();
    let array = f.state.string_map_of_string_array().get_string_array(&name);
    let index = f.params.index().value();
    let value = array.get_string(index).value();
    f.results.value().set_value(&value);
}

//////////////////// StringMap of StringMap \\\\\\\\\\\\\\\\\\\\

pub fn func_string_map_of_string_map_clear(
    _ctx: &ScFuncContext,
    f: &StringMapOfStringMapClearContext,
) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_string_map().get_string_map(&name);
    mmap.clear();
}

pub fn func_string_map_of_string_map_set(_ctx: &ScFuncContext, f: &StringMapOfStringMapSetContext) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_string_map().get_string_map(&name);
    let key = f.params.key().value();
    let value = f.params.value().value();
    mmap.get_string(&key).set_value(&value);
}

pub fn view_string_map_of_string_map_value(
    _ctx: &ScViewContext,
    f: &StringMapOfStringMapValueContext,
) {
    let name = f.params.name().value();
    let mmap = f.state.string_map_of_string_map().get_string_map(&name);
    let key = f.params.key().value();
    f.results.value().set_value(&mmap.get_string(&key).value());
}

//////////////////// array of AddressArray \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_address_array_append(
    _ctx: &ScFuncContext,
    f: &ArrayOfAddressArrayAppendContext,
) {
    let index = f.params.index().value();
    let val_len = f.params.value_addr().length();

    let sa: ArrayOfMutableAddress;
    if f.state.array_of_string_array().length() <= index {
        sa = f.state.array_of_address_array().append_address_array();
    } else {
        sa = f.state.array_of_address_array().get_address_array(index);
    }

    for i in 0..val_len {
        let elt = f.params.value_addr().get_address(i).value();
        sa.append_address().set_value(&elt);
    }
}

pub fn func_array_of_address_array_clear(
    _ctx: &ScFuncContext,
    f: &ArrayOfAddressArrayClearContext,
) {
    let length = f.state.array_of_address_array().length();
    for i in 0..length {
        let array = f.state.array_of_address_array().get_address_array(i);
        array.clear();
    }
    f.state.array_of_address_array().clear();
}

pub fn func_array_of_address_array_set(_ctx: &ScFuncContext, f: &ArrayOfAddressArraySetContext) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();
    let array = f.state.array_of_address_array().get_address_array(index0);
    let value = f.params.value_addr().value();
    array.get_address(index1).set_value(&value);
}

pub fn view_array_of_address_array_length(
    _ctx: &ScViewContext,
    f: &ArrayOfAddressArrayLengthContext,
) {
    let length = f.state.array_of_address_array().length();
    f.results.length().set_value(length);
}

pub fn view_array_of_address_array_value(
    _ctx: &ScViewContext,
    f: &ArrayOfAddressArrayValueContext,
) {
    let index0 = f.params.index0().value();
    let index1 = f.params.index1().value();

    let elt = f
        .state
        .array_of_address_array()
        .get_address_array(index0)
        .get_address(index1)
        .value();
    f.results.value_addr().set_value(&elt);
}

//////////////////// array of AddressMap \\\\\\\\\\\\\\\\\\\\

pub fn func_array_of_address_map_clear(_ctx: &ScFuncContext, f: &ArrayOfAddressMapClearContext) {
    let length = f.state.array_of_address_array().length();
    for i in 0..length {
        let mmap = f.state.array_of_address_map().get_address_map(i);
        mmap.clear();
    }
    f.state.array_of_address_map().clear();
}

pub fn func_array_of_address_map_set(_ctx: &ScFuncContext, f: &ArrayOfAddressMapSetContext) {
    let index = f.params.index().value();
    let value = f.params.value_addr().value();
    let key = f.params.key_addr().value();
    if f.state.array_of_address_map().length() <= index {
        let mmap = f.state.array_of_address_map().append_address_map();
        mmap.get_address(&key).set_value(&value);
        return;
    }
    let mmap = f.state.array_of_address_map().get_address_map(index);
    mmap.get_address(&key).set_value(&value);
}

pub fn view_array_of_address_map_value(_ctx: &ScViewContext, f: &ArrayOfAddressMapValueContext) {
    let index = f.params.index().value();
    let key = f.params.key_addr().value();
    let mmap = f.state.array_of_address_map().get_address_map(index);
    f.results
        .value_addr()
        .set_value(&mmap.get_address(&key).value());
}

//////////////////// AddressMap of AddressArray \\\\\\\\\\\\\\\\\\\\

pub fn func_address_map_of_address_array_append(
    _ctx: &ScFuncContext,
    f: &AddressMapOfAddressArrayAppendContext,
) {
    let addr = f.params.name_addr().value();
    let array = f
        .state
        .address_map_of_address_array()
        .get_address_array(&addr);
    let value = f.params.value_addr().value();
    array.append_address().set_value(&value);
}

pub fn func_address_map_of_address_array_clear(
    _ctx: &ScFuncContext,
    f: &AddressMapOfAddressArrayClearContext,
) {
    let addr = f.params.name_addr().value();
    let array = f
        .state
        .address_map_of_address_array()
        .get_address_array(&addr);
    array.clear();
}

pub fn func_address_map_of_address_array_set(
    _ctx: &ScFuncContext,
    f: &AddressMapOfAddressArraySetContext,
) {
    let addr = f.params.name_addr().value();
    let array = f
        .state
        .address_map_of_address_array()
        .get_address_array(&addr);
    let index = f.params.index().value();
    let value = f.params.value_addr().value();
    array.get_address(index).set_value(&value);
}

pub fn view_address_map_of_address_array_length(
    _ctx: &ScViewContext,
    f: &AddressMapOfAddressArrayLengthContext,
) {
    let addr = f.params.name_addr().value();
    let array = f
        .state
        .address_map_of_address_array()
        .get_address_array(&addr);
    let length = array.length();
    f.results.length().set_value(length);
}

pub fn view_address_map_of_address_array_value(
    _ctx: &ScViewContext,
    f: &AddressMapOfAddressArrayValueContext,
) {
    let addr = f.params.name_addr().value();
    let array = f
        .state
        .address_map_of_address_array()
        .get_address_array(&addr);
    let index = f.params.index().value();
    let value = array.get_address(index).value();
    f.results.value_addr().set_value(&value);
}

//////////////////// AddressMap of AddressMap \\\\\\\\\\\\\\\\\\\\

pub fn func_address_map_of_address_map_clear(
    _ctx: &ScFuncContext,
    f: &AddressMapOfAddressMapClearContext,
) {
    let name = f.params.name_addr().value();
    let my_map = f.state.address_map_of_address_map().get_address_map(&name);
    my_map.clear();
}

pub fn func_address_map_of_address_map_set(
    _ctx: &ScFuncContext,
    f: &AddressMapOfAddressMapSetContext,
) {
    let name = f.params.name_addr().value();
    let my_map = f.state.address_map_of_address_map().get_address_map(&name);
    let key = f.params.key_addr().value();
    let value = f.params.value_addr().value();
    my_map.get_address(&key).set_value(&value);
}

pub fn view_address_map_of_address_map_value(
    _ctx: &ScViewContext,
    f: &AddressMapOfAddressMapValueContext,
) {
    let name = f.params.name_addr().value();
    let my_map = f.state.address_map_of_address_map().get_address_map(&name);
    let key = f.params.key_addr().value();
    f.results
        .value_addr()
        .set_value(&my_map.get_address(&key).value());
}

pub fn view_big_int_add(_ctx: &ScViewContext, f: &BigIntAddContext) {
    let lhs = f.params.lhs().value();
    let rhs = f.params.rhs().value();
    let res = lhs.add(&rhs);
    f.results.res().set_value(&res);
}

pub fn view_big_int_div(_ctx: &ScViewContext, f: &BigIntDivContext) {
    let lhs = f.params.lhs().value();
    let rhs = f.params.rhs().value();
    let res = lhs.div(&rhs);
    f.results.res().set_value(&res);
}

pub fn view_big_int_mod(_ctx: &ScViewContext, f: &BigIntModContext) {
    let lhs = f.params.lhs().value();
    let rhs = f.params.rhs().value();
    let res = lhs.modulo(&rhs);
    f.results.res().set_value(&res);
}

pub fn view_big_int_mul(_ctx: &ScViewContext, f: &BigIntMulContext) {
    let lhs = f.params.lhs().value();
    let rhs = f.params.rhs().value();
    let res = lhs.mul(&rhs);
    f.results.res().set_value(&res);
}

pub fn view_big_int_sub(_ctx: &ScViewContext, f: &BigIntSubContext) {
    let lhs = f.params.lhs().value();
    let rhs = f.params.rhs().value();
    let res = lhs.sub(&rhs);
    f.results.res().set_value(&res);
}

pub fn view_big_int_shl(_ctx: &ScViewContext, f: &BigIntShlContext) {
    let lhs = f.params.lhs().value();
    let shift = f.params.shift().value();
    let res = lhs.shl(shift);
    f.results.res().set_value(&res);
}

pub fn view_big_int_shr(_ctx: &ScViewContext, f: &BigIntShrContext) {
    let lhs = f.params.lhs().value();
    let shift = f.params.shift().value();
    let res = lhs.shr(shift);
    f.results.res().set_value(&res);
}

pub fn view_check_agent_id(ctx: &ScViewContext, f: &CheckAgentIDContext) {
    let sc_agent_id = f.params.sc_agent_id().value();
    let agent_bytes = f.params.agent_bytes().value();
    let agent_string = f.params.agent_string().value();
    ctx.require(
        sc_agent_id == agent_id_from_bytes(&agent_id_to_bytes(&sc_agent_id)),
        "bytes conversion failed",
    );
    ctx.require(
        sc_agent_id == agent_id_from_string(&agent_id_to_string(&sc_agent_id)),
        "string conversion failed",
    );
    ctx.require(sc_agent_id.to_bytes() == agent_bytes, "bytes mismatch");
    ctx.require(sc_agent_id.to_string() == agent_string, "string mismatch");
}

pub fn view_check_address(ctx: &ScViewContext, f: &CheckAddressContext) {
    let sc_address = f.params.sc_address().value();
    let address_bytes = f.params.address_bytes().value();
    let address_string = f.params.address_string().value();
    ctx.require(
        sc_address == address_from_bytes(&address_to_bytes(&sc_address)),
        "bytes conversion failed",
    );
    ctx.require(
        sc_address == address_from_string(&address_to_string(&sc_address)),
        "string conversion failed",
    );
    ctx.require(sc_address.to_bytes() == address_bytes, "bytes mismatch");
    ctx.require(sc_address.to_string() == address_string, "string mismatch");
}
