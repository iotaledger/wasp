// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::structs::*;

pub fn func_mint_supply(ctx: &ScFuncContext, f: &MintSupplyContext) {
    let minted = ctx.minted();
    let minted_colors = minted.token_ids();
    ctx.require(minted_colors.len() == 1, "need single minted color");
    let minted_color = minted_colors.get(0).unwrap();
    let current_token = f.state.registry().get_token(&minted_color);
    if current_token.exists() {
        // should never happen, because transaction id is unique
        ctx.panic("TokenRegistry: registry for color already exists");
    }
    let mut token = Token {
        supply: minted.balance(&minted_color).uint64(),
        minted_by: ctx.caller(),
        owner: ctx.caller(),
        created: ctx.timestamp(),
        updated: ctx.timestamp(),
        description: f.params.description().value(),
        user_defined: f.params.user_defined().value(),
    };
    if token.description.is_empty() {
        token.description += "no dscr";
    }
    current_token.set_value(&token);
    let token_list = f.state.token_list();
    token_list.append_token_id().set_value(&minted_color);
}

pub fn func_transfer_ownership(_ctx: &ScFuncContext, _f: &TransferOwnershipContext) {
    // TODO
}

pub fn func_update_metadata(_ctx: &ScFuncContext, _f: &UpdateMetadataContext) {
    // TODO
}

pub fn view_get_info(_ctx: &ScViewContext, _f: &GetInfoContext) {
    // TODO
}
