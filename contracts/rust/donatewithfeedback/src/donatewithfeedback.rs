// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::types::*;

pub fn func_donate(ctx: &ScFuncContext) {
    ctx.log("dwf.donate");
    let p = ctx.params();
    let mut donation = Donation {
        amount: ctx.incoming().balance(&ScColor::IOTA),
        donator: ctx.caller(),
        error: String::new(),
        feedback: p.get_string(PARAM_FEEDBACK).value(),
        timestamp: ctx.timestamp(),
    };
    if donation.amount == 0 || donation.feedback.len() == 0 {
        donation.error = "error: empty feedback or donated amount = 0".to_string();
        if donation.amount > 0 {
            ctx.transfer_to_address(&donation.donator.address(), ScTransfers::iotas(donation.amount));
            donation.amount = 0;
        }
    }
    let state = ctx.state();
    let log = state.get_bytes_array(VAR_LOG);
    log.get_bytes(log.length()).set_value(&donation.to_bytes());

    let largest_donation = state.get_int64(VAR_MAX_DONATION);
    let total_donated = state.get_int64(VAR_TOTAL_DONATION);
    if donation.amount > largest_donation.value() {
        largest_donation.set_value(donation.amount);
    }
    total_donated.set_value(total_donated.value() + donation.amount);
    ctx.log("dwf.donate ok");
}

pub fn func_withdraw(ctx: &ScFuncContext) {
    ctx.log("dwf.withdraw");

    // only SC creator can withdraw donated funds
    ctx.require(ctx.caller() == ctx.contract_creator(), "no permission");

    let balance = ctx.balances().balance(&ScColor::IOTA);
    let p = ctx.params();
    let mut amount = p.get_int64(PARAM_AMOUNT).value();
    if amount == 0 || amount > balance {
        amount = balance;
    }
    if amount == 0 {
        ctx.log("dwf.withdraw: nothing to withdraw");
        return;
    }

    let sc_creator = ctx.contract_creator().address();
    ctx.transfer_to_address(&sc_creator, ScTransfers::iotas(amount));

    ctx.log("dwf.withdraw ok");
}

pub fn view_donation(ctx: &ScViewContext) {
    ctx.log("dwf.donation");
    let params = ctx.params();
    let nr = params.get_int64(PARAM_NR).value() as i32;
    let state = ctx.state();
    let results = ctx.results();
    let donation = Donation::from_bytes(&state.get_bytes_array(VAR_LOG).get_bytes(nr).value());
    results.get_int64(RESULT_AMOUNT).set_value(donation.amount);
    results.get_agent_id(RESULT_DONATOR).set_value(&donation.donator);
    results.get_string(RESULT_ERROR).set_value(&donation.error);
    results.get_string(RESULT_FEEDBACK).set_value(&donation.feedback);
    results.get_int64(RESULT_TIMESTAMP).set_value(donation.timestamp);

    ctx.log("dwf.donation ok");
}

pub fn view_donation_info(ctx: &ScViewContext) {
    ctx.log("dwf.donation_info");

    let state = ctx.state();
    let results = ctx.results();
    results.get_int64(RESULT_MAX_DONATION).set_value(state.get_int64(VAR_MAX_DONATION).value());
    results.get_int64(RESULT_TOTAL_DONATION).set_value(state.get_int64(VAR_TOTAL_DONATION).value());
    results.get_int64(RESULT_COUNT).set_value(state.get_bytes_array(VAR_LOG).length() as i64);

    ctx.log("dwf.donation_info ok");
}
