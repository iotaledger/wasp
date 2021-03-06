// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;

pub fn func_param_types(ctx: &ScFuncContext, params: &FuncParamTypesParams) {
    if params.address.exists() {
        ctx.require(params.address.value() == ctx.account_id().address(), "mismatch: Address");
    }
    if params.agent_id.exists() {
        ctx.require(params.agent_id.value() == ctx.account_id(), "mismatch: AgentID");
    }
    if params.bytes.exists() {
        let bytes = "these are bytes".as_bytes();
        ctx.require(params.bytes.value() == bytes, "mismatch: Bytes");
    }
    if params.chain_id.exists() {
        ctx.require(params.chain_id.value() == ctx.chain_id(), "mismatch: ChainID");
    }
    if params.color.exists() {
        let color = ScColor::from_bytes("RedGreenBlueYellowCyanBlackWhite".as_bytes());
        ctx.require(params.color.value() == color, "mismatch: Color");
    }
    if params.hash.exists() {
        let hash = ScHash::from_bytes("0123456789abcdeffedcba9876543210".as_bytes());
        ctx.require(params.hash.value() == hash, "mismatch: Hash");
    }
    if params.hname.exists() {
        ctx.require(params.hname.value() == ctx.account_id().hname(), "mismatch: Hname");
    }
    if params.int16.exists() {
        ctx.require(params.int16.value() == 12345, "mismatch: Int16");
    }
    if params.int32.exists() {
        ctx.require(params.int32.value() == 1234567890, "mismatch: Int32");
    }
    if params.int64.exists() {
        ctx.require(params.int64.value() == 1234567890123456789, "mismatch: Int64");
    }
    if params.request_id.exists() {
        let request_id = ScRequestID::from_bytes("abcdefghijklmnopqrstuvwxyz123456\x00\x00".as_bytes());
        ctx.require(params.request_id.value() == request_id, "mismatch: RequestID");
    }
    if params.string.exists() {
        ctx.require(params.string.value() == "this is a string", "mismatch: String");
    }
}
