// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::contract::*;
use crate::structs::*;

const DURATION_DEFAULT: u32 = 60;
const DURATION_MIN: u32 = 1;
const DURATION_MAX: u32 = 120;
const MAX_DESCRIPTION_LENGTH: usize = 150;
const OWNER_MARGIN_DEFAULT: u64 = 50;
const OWNER_MARGIN_MIN: u64 = 5;
const OWNER_MARGIN_MAX: u64 = 100;

pub fn func_finalize_auction(ctx: &ScFuncContext, f: &FinalizeAuctionContext) {
    let token = f.params.token().value();
    let current_auction = f.state.auctions().get_auction(&token);
    ctx.require(current_auction.exists(), "Missing auction info");
    let auction = current_auction.value();
    if auction.highest_bid == 0 {
        ctx.log(&("No one bid on ".to_string() + &token.to_string()));
        let mut owner_fee = auction.minimum_bid * auction.owner_margin / 1000;
        if owner_fee == 0 {
            owner_fee = 1;
        }
        // finalizeAuction request token was probably not confirmed yet
        transfer_iotas(ctx, &ctx.contract_creator(), owner_fee - 1);
        transfer_tokens(ctx, &auction.creator, &auction.token, auction.num_tokens);
        transfer_iotas(ctx, &auction.creator, auction.deposit - owner_fee);
        return;
    }

    let mut owner_fee = auction.highest_bid * auction.owner_margin / 1000;
    if owner_fee == 0 {
        owner_fee = 1;
    }

    // return staked bids to losers
    let bids = f.state.bids().get_bids(&token);
    let bidder_list = f.state.bidder_list().get_bidder_list(&token);
    let size = bidder_list.length();
    for i in 0..size {
        let loser = bidder_list.get_agent_id(i).value();
        if loser != auction.highest_bidder {
            let bid = bids.get_bid(&loser).value();
            transfer_iotas(ctx, &loser, bid.amount);
        }
    }

    // finalizeAuction request token was probably not confirmed yet
    transfer_iotas(ctx, &ctx.contract_creator(), owner_fee - 1);
    transfer_tokens(ctx, &auction.highest_bidder, &auction.token, auction.num_tokens);
    transfer_iotas(ctx, &auction.creator, auction.deposit + auction.highest_bid - owner_fee);
}

pub fn func_place_bid(ctx: &ScFuncContext, f: &PlaceBidContext) {
    let mut bid_amount = ctx.allowance().iotas();
    ctx.require(bid_amount > 0, "Missing bid amount");

    let token = f.params.token().value();
    let current_auction = f.state.auctions().get_auction(&token);
    ctx.require(current_auction.exists(), "Missing auction info");

    let mut auction = current_auction.value();
    let bids = f.state.bids().get_bids(&token);
    let bidder_list = f.state.bidder_list().get_bidder_list(&token);
    let caller = ctx.caller();
    let current_bid = bids.get_bid(&caller);
    if current_bid.exists() {
        ctx.log(&("Upped bid from: ".to_string() + &caller.to_string()));
        let mut bid = current_bid.value();
        bid_amount += bid.amount;
        bid.amount = bid_amount;
        bid.timestamp = ctx.timestamp();
        current_bid.set_value(&bid);
    } else {
        ctx.require(bid_amount >= auction.minimum_bid, "Insufficient bid amount");
        ctx.log(&("New bid from: ".to_string() + &caller.to_string()));
        let index = bidder_list.length();
        bidder_list.append_agent_id().set_value(&caller);
        let bid = Bid {
            index: index,
            amount: bid_amount,
            timestamp: ctx.timestamp(),
        };
        current_bid.set_value(&bid);
    }
    if bid_amount > auction.highest_bid {
        ctx.log("New highest bidder");
        auction.highest_bid = bid_amount;
        auction.highest_bidder = caller;
        current_auction.set_value(&auction);
    }
}

pub fn func_set_owner_margin(_ctx: &ScFuncContext, f: &SetOwnerMarginContext) {
    let mut owner_margin = f.params.owner_margin().value();
    if owner_margin < OWNER_MARGIN_MIN {
        owner_margin = OWNER_MARGIN_MIN;
    }
    if owner_margin > OWNER_MARGIN_MAX {
        owner_margin = OWNER_MARGIN_MAX;
    }
    f.state.owner_margin().set_value(owner_margin);
}

pub fn func_start_auction(ctx: &ScFuncContext, f: &StartAuctionContext) {
    let token = f.params.token().value();
    let num_tokens = ctx.allowance().balance(&token);
    if num_tokens.is_zero() {
        ctx.panic("Missing auction tokens");
    }

    let minimum_bid = f.params.minimum_bid().value();

    // duration in minutes
    let mut duration = f.params.duration().value();
    if duration == 0 {
        duration = DURATION_DEFAULT;
    }
    if duration < DURATION_MIN {
        duration = DURATION_MIN;
    }
    if duration > DURATION_MAX {
        duration = DURATION_MAX;
    }

    let mut description = f.params.description().value();
    if description == "" {
        description = "N/A".to_string();
    }
    if description.len() > MAX_DESCRIPTION_LENGTH {
        let ss: String = description.chars().take(MAX_DESCRIPTION_LENGTH).collect();
        description = ss + "[...]";
    }

    let mut owner_margin = f.state.owner_margin().value();
    if owner_margin == 0 {
        owner_margin = OWNER_MARGIN_DEFAULT;
    }

    // need at least 1 iota to run SC
    let mut margin = minimum_bid * owner_margin / 1000;
    if margin == 0 {
        margin = 1;
    }
    let deposit = ctx.allowance().iotas();
    if deposit < margin {
        ctx.panic("Insufficient deposit");
    }

    let current_auction = f.state.auctions().get_auction(&token);
    if current_auction.exists() {
        ctx.panic("Auction for this token already exists");
    }

    let auction = Auction {
        creator: ctx.caller(),
        deposit: deposit,
        description: description,
        duration: duration,
        highest_bid: 0,
        highest_bidder: agent_id_from_bytes(&[]),
        minimum_bid: minimum_bid,
        num_tokens: num_tokens.uint64(),
        owner_margin: owner_margin,
        token: token,
        when_started: ctx.timestamp(),
    };
    current_auction.set_value(&auction);

    let fa = ScFuncs::finalize_auction(ctx);
    fa.params.token().set_value(&auction.token);
    fa.func.delay(duration * 60).post();
}

pub fn view_get_info(ctx: &ScViewContext, f: &GetInfoContext) {
    let token = f.params.token().value();
    let current_auction = f.state.auctions().get_auction(&token);
    ctx.require(current_auction.exists(), "Missing auction info");

    let auction = current_auction.value();
    f.results.creator().set_value(&auction.creator);
    f.results.deposit().set_value(auction.deposit);
    f.results.description().set_value(&auction.description);
    f.results.duration().set_value(auction.duration);
    f.results.highest_bid().set_value(auction.highest_bid);
    f.results.highest_bidder().set_value(&auction.highest_bidder);
    f.results.minimum_bid().set_value(auction.minimum_bid);
    f.results.num_tokens().set_value(auction.num_tokens);
    f.results.owner_margin().set_value(auction.owner_margin);
    f.results.token().set_value(&auction.token);
    f.results.when_started().set_value(auction.when_started);

    let bidder_list = f.state.bidder_list().get_bidder_list(&token);
    f.results.bidders().set_value(bidder_list.length());
}

fn transfer_iotas(ctx: &ScFuncContext, agent: &ScAgentID, amount: u64) {
    if agent.is_address() {
        // send back to original Tangle address
        ctx.send(&agent.address(), &ScTransfer::iotas(amount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(&agent.address(), &ScTransfer::iotas(amount));
}

fn transfer_tokens(ctx: &ScFuncContext, agent: &ScAgentID, token: &ScTokenID, amount: u64) {
    let big_amount = ScBigInt::from_uint64(amount);
    if agent.is_address() {
        // send back to original Tangle address
        ctx.send(&agent.address(), &ScTransfer::tokens(token, &big_amount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(&agent.address(), &ScTransfer::tokens(token, &big_amount));
}
