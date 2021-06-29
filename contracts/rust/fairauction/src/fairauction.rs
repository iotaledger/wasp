// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::types::*;

const DURATION_DEFAULT: i32 = 60;
const DURATION_MIN: i32 = 1;
const DURATION_MAX: i32 = 120;
const MAX_DESCRIPTION_LENGTH: usize = 150;
const OWNER_MARGIN_DEFAULT: i64 = 50;
const OWNER_MARGIN_MIN: i64 = 5;
const OWNER_MARGIN_MAX: i64 = 100;

pub fn func_finalize_auction(ctx: &ScFuncContext) {
    ctx.log("fairauction.finalize");
    // only SC itself can invoke this function
    ctx.require(ctx.caller() == ctx.account_id(), "no permission");

    let p = ctx.params();
    let param_color = p.get_color(PARAM_COLOR);
    ctx.require(param_color.exists(), "missing mandatory color");

    let color = param_color.value();
    let state = ctx.state();
    let current_auction = state.get_map(STATE_AUCTIONS).get_bytes(&color);
    ctx.require(current_auction.exists(), "Missing auction info");
    let auction = Auction::from_bytes(&current_auction.value());
    if auction.highest_bid < 0 {
        ctx.log(&("No one bid on ".to_string() + &color.to_string()));
        let mut owner_fee = auction.minimum_bid * auction.owner_margin / 1000;
        if owner_fee == 0 {
            owner_fee = 1;
        }
        // finalizeAuction request token was probably not confirmed yet
        transfer_tokens(ctx, &ctx.contract_creator(), &ScColor::IOTA, owner_fee - 1);
        transfer_tokens(ctx, &auction.creator, &auction.color, auction.num_tokens);
        transfer_tokens(ctx, &auction.creator, &ScColor::IOTA, auction.deposit - owner_fee);
        return;
    }

    let mut owner_fee = auction.highest_bid * auction.owner_margin / 1000;
    if owner_fee == 0 {
        owner_fee = 1;
    }

    // return staked bids to losers
    let bids = state.get_map(STATE_BIDS).get_map(&color);
    let bidder_list = state.get_map(STATE_BIDDER_LIST).get_agent_id_array(&color);
    let size = bidder_list.length();
    for i in 0..size {
        let loser = bidder_list.get_agent_id(i).value();
        if loser != auction.highest_bidder {
            let bid = Bid::from_bytes(&bids.get_bytes(&loser).value());
            transfer_tokens(ctx, &loser, &ScColor::IOTA, bid.amount);
        }
    }

    // finalizeAuction request token was probably not confirmed yet
    transfer_tokens(ctx, &ctx.contract_creator(), &ScColor::IOTA, owner_fee - 1);
    transfer_tokens(ctx, &auction.highest_bidder, &auction.color, auction.num_tokens);
    transfer_tokens(ctx, &auction.creator, &ScColor::IOTA, auction.deposit + auction.highest_bid - owner_fee);
    ctx.log("fairauction.finalize ok");
}

pub fn func_place_bid(ctx: &ScFuncContext) {
    ctx.log("fairauction.placeBid");
    let p = ctx.params();
    let param_color = p.get_color(PARAM_COLOR);
    ctx.require(param_color.exists(), "missing mandatory color");

    let mut bid_amount = ctx.incoming().balance(&ScColor::IOTA);
    ctx.require(bid_amount > 0, "Missing bid amount");

    let color = param_color.value();
    let state = ctx.state();
    let current_auction = state.get_map(STATE_AUCTIONS).get_bytes(&color);
    ctx.require(current_auction.exists(), "Missing auction info");

    let mut auction = Auction::from_bytes(&current_auction.value());
    let bids = state.get_map(STATE_BIDS).get_map(&color);
    let bidder_list = state.get_map(STATE_BIDDER_LIST).get_agent_id_array(&color);
    let caller = ctx.caller();
    let current_bid = bids.get_bytes(&caller);
    if current_bid.exists() {
        ctx.log(&("Upped bid from: ".to_string() + &caller.to_string()));
        let mut bid = Bid::from_bytes(&current_bid.value());
        bid_amount += bid.amount;
        bid.amount = bid_amount;
        bid.timestamp = ctx.timestamp();
        current_bid.set_value(&bid.to_bytes());
    } else {
        ctx.require(bid_amount >= auction.minimum_bid, "Insufficient bid amount");
        ctx.log(&("New bid from: ".to_string() + &caller.to_string()));
        let index = bidder_list.length();
        bidder_list.get_agent_id(index).set_value(&caller);
        let bid = Bid {
            index: index,
            amount: bid_amount,
            timestamp: ctx.timestamp(),
        };
        current_bid.set_value(&bid.to_bytes());
    }
    if bid_amount > auction.highest_bid {
        ctx.log("New highest bidder");
        auction.highest_bid = bid_amount;
        auction.highest_bidder = caller;
        current_auction.set_value(&auction.to_bytes());
    }
    ctx.log("fairauction.placeBid ok");
}

pub fn func_set_owner_margin(ctx: &ScFuncContext) {
    ctx.log("fairauction.setOwnerMargin");
    // only SC creator can set owner margin
    ctx.require(ctx.caller() == ctx.contract_creator(), "no permission");

    let p = ctx.params();
    let param_owner_margin = p.get_int64(PARAM_OWNER_MARGIN);

    ctx.require(param_owner_margin.exists(), "missing mandatory ownerMargin");

    let mut owner_margin = param_owner_margin.value();
    if owner_margin < OWNER_MARGIN_MIN {
        owner_margin = OWNER_MARGIN_MIN;
    }
    if owner_margin > OWNER_MARGIN_MAX {
        owner_margin = OWNER_MARGIN_MAX;
    }
    ctx.state().get_int64(STATE_OWNER_MARGIN).set_value(owner_margin);
    ctx.log("fairauction.setOwnerMargin ok");
}

pub fn func_start_auction(ctx: &ScFuncContext) {
    ctx.log("fairauction.startAuction");
    let p = ctx.params();
    let param_color = p.get_color(PARAM_COLOR);
    let param_description = p.get_string(PARAM_DESCRIPTION);
    let param_duration = p.get_int32(PARAM_DURATION);
    let param_minimum_bid = p.get_int64(PARAM_MINIMUM_BID);

    ctx.require(param_color.exists(), "missing mandatory color");
    ctx.require(param_minimum_bid.exists(), "missing mandatory minimumBid");

    let color = param_color.value();
    if color == ScColor::IOTA || color == ScColor::MINT {
        ctx.panic("Reserved auction token color");
    }
    let num_tokens = ctx.incoming().balance(&color);
    if num_tokens == 0 {
        ctx.panic("Missing auction tokens");
    }

    let minimum_bid = param_minimum_bid.value();

    // duration in minutes
    let mut duration = param_duration.value();
    if duration == 0 {
        duration = DURATION_DEFAULT;
    }
    if duration < DURATION_MIN {
        duration = DURATION_MIN;
    }
    if duration > DURATION_MAX {
        duration = DURATION_MAX;
    }

    let mut description = param_description.value();
    if description == "" {
        description = "N/A".to_string();
    }
    if description.len() > MAX_DESCRIPTION_LENGTH {
        let ss: String = description.chars().take(MAX_DESCRIPTION_LENGTH).collect();
        description = ss + "[...]";
    }

    let state = ctx.state();
    let mut owner_margin = state.get_int64(STATE_OWNER_MARGIN).value();
    if owner_margin == 0 {
        owner_margin = OWNER_MARGIN_DEFAULT;
    }

    // need at least 1 iota to run SC
    let mut margin = minimum_bid * owner_margin / 1000;
    if margin == 0 {
        margin = 1;
    }
    let deposit = ctx.incoming().balance(&ScColor::IOTA);
    if deposit < margin {
        ctx.panic("Insufficient deposit");
    }

    let current_auction = state.get_map(STATE_AUCTIONS).get_bytes(&color);
    if current_auction.exists() {
        ctx.panic("Auction for this token color already exists");
    }

    let auction = Auction {
        creator: ctx.caller(),
        color: color,
        deposit: deposit,
        description: description,
        duration: duration,
        highest_bid: -1,
        highest_bidder: ScAgentId::from_bytes(&[0; 37]),
        minimum_bid: minimum_bid,
        num_tokens: num_tokens,
        owner_margin: owner_margin,
        when_started: ctx.timestamp(),
    };
    current_auction.set_value(&auction.to_bytes());

    let finalize_params = ScMutableMap::new();
    finalize_params.get_color(PARAM_COLOR).set_value(&auction.color);
    let transfer = ScTransfers::iotas(1);
    ctx.post_self(HFUNC_FINALIZE_AUCTION, Some(finalize_params), transfer, duration * 60);
    ctx.log("fairauction.startAuction ok");
}

pub fn view_get_info(ctx: &ScViewContext) {
    ctx.log("fairauction.getInfo");
    let p = ctx.params();
    let param_color = p.get_color(PARAM_COLOR);

    ctx.require(param_color.exists(), "missing mandatory color");
    let color = param_color.value();
    let state = ctx.state();
    let current_auction = state.get_map(STATE_AUCTIONS).get_bytes(&color);
    ctx.require(current_auction.exists(), "Missing auction info");

    let auction = Auction::from_bytes(&current_auction.value());
    let results = ctx.results();
    results.get_color(RESULT_COLOR).set_value(&auction.color);
    results.get_agent_id(RESULT_CREATOR).set_value(&auction.creator);
    results.get_int64(RESULT_DEPOSIT).set_value(auction.deposit);
    results.get_string(RESULT_DESCRIPTION).set_value(&auction.description);
    results.get_int32(RESULT_DURATION).set_value(auction.duration);
    results.get_int64(RESULT_HIGHEST_BID).set_value(auction.highest_bid);
    results.get_agent_id(RESULT_HIGHEST_BIDDER).set_value(&auction.highest_bidder);
    results.get_int64(RESULT_MINIMUM_BID).set_value(auction.minimum_bid);
    results.get_int64(RESULT_NUM_TOKENS).set_value(auction.num_tokens);
    results.get_int64(RESULT_OWNER_MARGIN).set_value(auction.owner_margin);
    results.get_int64(RESULT_WHEN_STARTED).set_value(auction.when_started);

    let bidder_list = state.get_map(STATE_BIDDER_LIST).get_agent_id_array(&color);
    results.get_int64(RESULT_BIDDERS).set_value(bidder_list.length() as i64);
    ctx.log("fairauction.getInfo ok");
}

fn transfer_tokens(ctx: &ScFuncContext, agent: &ScAgentId, color: &ScColor, amount: i64) {
    if agent.is_address() {
        // send back to original Tangle address
        ctx.transfer_to_address(&agent.address(), ScTransfers::new(color, amount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.transfer_to_address(&agent.address(), ScTransfers::new(color, amount));
}
