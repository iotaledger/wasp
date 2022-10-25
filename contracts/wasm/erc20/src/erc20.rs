// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0


use wasmlib::*;

use crate::*;

pub fn func_approve(ctx: &ScFuncContext, f: &ApproveContext) {
}

pub fn func_init(ctx: &ScFuncContext, f: &InitContext) {
    if f.params.owner().exists() {
        f.state.owner().set_value(&f.params.owner().value());
        return;
    }
    f.state.owner().set_value(&ctx.request_sender());
}

pub fn func_transfer(ctx: &ScFuncContext, f: &TransferContext) {
}

pub fn func_transfer_from(ctx: &ScFuncContext, f: &TransferFromContext) {
}

pub fn view_allowance(ctx: &ScViewContext, f: &AllowanceContext) {
}

pub fn view_balance_of(ctx: &ScViewContext, f: &BalanceOfContext) {
}

pub fn view_total_supply(ctx: &ScViewContext, f: &TotalSupplyContext) {
}
