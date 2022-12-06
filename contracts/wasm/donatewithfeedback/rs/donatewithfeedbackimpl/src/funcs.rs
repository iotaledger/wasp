// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;
use donatewithfeedback::*;
use crate::*;

pub fn func_donate(ctx: &ScFuncContext, f: &DonateContext) {
    let amount = ctx.allowance().base_tokens();
    let transfer = ScTransfer::base_tokens(amount);
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);
    let mut donation = Donation {
        amount: amount,
        donator: ctx.caller(),
        error: String::new(),
        feedback: f.params.feedback().value(),
        timestamp: ctx.timestamp(),
    };
    if donation.amount == 0 || donation.feedback.len() == 0 {
        donation.error = "error: empty feedback or donated amount = 0".to_string();
    }
    let log = f.state.log();
    log.append_donation().set_value(&donation);

    let largest_donation = f.state.max_donation();
    let total_donated = f.state.total_donation();
    if donation.amount > largest_donation.value() {
        largest_donation.set_value(donation.amount);
    }
    total_donated.set_value(total_donated.value() + donation.amount);
}

pub fn func_withdraw(ctx: &ScFuncContext, f: &WithdrawContext) {
    let balance = ctx.balances().base_tokens();
    let mut amount = f.params.amount().value();
    if amount == 0 || amount > balance {
        amount = balance;
    }
    if amount == 0 {
        ctx.log("dwf.withdraw: nothing to withdraw");
        return;
    }

    let sc_owner = f.state.owner().value().address();
    ctx.send(&sc_owner, &ScTransfer::base_tokens(amount));
}

pub fn view_donation(_ctx: &ScViewContext, f: &DonationContext) {
    let nr = f.params.nr().value();
    let donation = f.state.log().get_donation(nr).value();
    f.results.amount().set_value(donation.amount);
    f.results.donator().set_value(&donation.donator);
    f.results.error().set_value(&donation.error);
    f.results.feedback().set_value(&donation.feedback);
    f.results.timestamp().set_value(donation.timestamp);
}

pub fn view_donation_info(_ctx: &ScViewContext, f: &DonationInfoContext) {
    f.results.max_donation().set_value(f.state.max_donation().value());
    f.results.total_donation().set_value(f.state.total_donation().value());
    f.results.count().set_value(f.state.log().length());
}

pub fn func_init(ctx: &ScFuncContext, f: &InitContext) {
    if f.params.owner().exists() {
        f.state.owner().set_value(&f.params.owner().value());
        return;
    }
    f.state.owner().set_value(&ctx.request_sender());
}
