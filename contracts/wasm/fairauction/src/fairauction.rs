// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::contract::*;
use crate::structs::*;
use crate::*;

const DURATION_DEFAULT: u32 = 60;
const DURATION_MIN: u32 = 1;
const DURATION_MAX: u32 = 120;
const MAX_DESCRIPTION_LENGTH: usize = 150;
const OWNER_MARGIN_DEFAULT: u64 = 50;
const OWNER_MARGIN_MIN: u64 = 5;
const OWNER_MARGIN_MAX: u64 = 100;

pub fn func_start_auction(ctx: &ScFuncContext, f: &StartAuctionContext) {
    let allowance = ctx.allowance();
    let nfts = allowance.nft_ids();
    ctx.require(nfts.len() == 1, "single NFT allowance expected");
    let auction_nft = nfts[0];

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

    //TODO need at least 1 iota to run SC
    let mut margin = minimum_bid * owner_margin / 1000;
    if margin == 0 {
        margin = 1;
    }
    let deposit = allowance.iotas();
    if deposit < margin {
        ctx.panic("Insufficient deposit");
    }

    let current_auction = f.state.auctions().get_auction(&auction_nft);
    if current_auction.exists() {
        ctx.panic("Auction for this nft already exists");
    }

    let auction = Auction {
        creator: ctx.caller(),
        deposit: deposit,
        description: description,
        duration: duration,
        highest_bid: 0,
        highest_bidder: ctx.caller(),
        minimum_bid: minimum_bid,
        owner_margin: owner_margin,
        nft: auction_nft,
        when_started: ctx.timestamp(),
    };
    current_auction.set_value(&auction);

    // take custody of deposit and NFT
    let mut transfer = ScTransfer::iotas(deposit);
    transfer.add_nft(&auction_nft);
    ctx.transfer_allowed(&ctx.account_id(), &transfer, false);

    let fa = ScFuncs::finalize_auction(ctx);
    fa.params.nft().set_value(&auction.nft);
    fa.func.delay(duration * 60).post();
}

pub fn func_place_bid(ctx: &ScFuncContext, f: &PlaceBidContext) {
    let mut bid_amount = ctx.allowance().iotas();
    ctx.require(bid_amount > 0, "Missing bid amount");

    let nft = f.params.nft().value();
    let current_auction = f.state.auctions().get_auction(&nft);
    ctx.require(current_auction.exists(), "Missing auction info");

    let mut auction = current_auction.value();
    let bids = f.state.bids().get_bids(&nft);
    let bidder_list = f.state.bidder_list().get_bidder_list(&nft);
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

pub fn func_finalize_auction(ctx: &ScFuncContext, f: &FinalizeAuctionContext) {
    let nft = f.params.nft().value();
    let current_auction = f.state.auctions().get_auction(&nft);
    ctx.require(current_auction.exists(), "Missing auction info");
    let auction = current_auction.value();
    if auction.highest_bid == 0 {
        ctx.log(&("No one bid on ".to_string() + &nft.to_string()));
        let mut owner_fee = auction.minimum_bid * auction.owner_margin / 1000;
        if owner_fee == 0 {
            owner_fee = 1;
        }
        // finalizeAuction request nft was probably not confirmed yet
        transfer_tokens(ctx, &f.state.owner().value(), owner_fee - 1);
        transfer_nft(ctx, &auction.creator, &auction.nft);
        transfer_tokens(ctx, &auction.creator, auction.deposit - owner_fee);
        return;
    }

    let mut owner_fee = auction.highest_bid * auction.owner_margin / 1000;
    if owner_fee == 0 {
        owner_fee = 1;
    }

    // return staked bids to losers
    let bids = f.state.bids().get_bids(&nft);
    let bidder_list = f.state.bidder_list().get_bidder_list(&nft);
    let size = bidder_list.length();
    for i in 0..size {
        let loser = bidder_list.get_agent_id(i).value();
        if loser != auction.highest_bidder {
            let bid = bids.get_bid(&loser).value();
            transfer_tokens(ctx, &loser, bid.amount);
        }
    }

    // finalizeAuction request nft was probably not confirmed yet
    transfer_tokens(ctx, &f.state.owner().value(), owner_fee - 1);
    transfer_nft(ctx, &auction.highest_bidder, &auction.nft);
    transfer_tokens(
        ctx,
        &auction.creator,
        auction.deposit + auction.highest_bid - owner_fee,
    );
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

pub fn view_get_auction_info(ctx: &ScViewContext, f: &GetAuctionInfoContext) {
    let nft = f.params.nft().value();
    let current_auction = f.state.auctions().get_auction(&nft);
    ctx.require(current_auction.exists(), "Missing auction info");

    let auction = current_auction.value();
    f.results.creator().set_value(&auction.creator);
    f.results.deposit().set_value(auction.deposit);
    f.results.description().set_value(&auction.description);
    f.results.duration().set_value(auction.duration);
    f.results.highest_bid().set_value(auction.highest_bid);
    f.results
        .highest_bidder()
        .set_value(&auction.highest_bidder);
    f.results.minimum_bid().set_value(auction.minimum_bid);
    f.results.owner_margin().set_value(auction.owner_margin);
    f.results.nft().set_value(&auction.nft);
    f.results.when_started().set_value(auction.when_started);

    let bidder_list = f.state.bidder_list().get_bidder_list(&nft);
    f.results.bidders().set_value(bidder_list.length());
}

fn transfer_tokens(ctx: &ScFuncContext, agent: &ScAgentID, amount: u64) {
    if agent.is_address() {
        // send back to original Tangle address
        ctx.send(&agent.address(), &ScTransfer::iotas(amount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(&agent.address(), &ScTransfer::iotas(amount));
}

fn transfer_nft(ctx: &ScFuncContext, agent: &ScAgentID, nft: &ScNftID) {
    if agent.is_address() {
        // send back to original Tangle address
        ctx.send(&agent.address(), &ScTransfer::nft(nft));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(&agent.address(), &ScTransfer::nft(nft));
}

pub fn func_init(ctx: &ScFuncContext, f: &InitContext) {
    if f.params.owner().exists() {
        f.state.owner().set_value(&f.params.owner().value());
        return;
    }
    f.state.owner().set_value(&ctx.request_sender());
}
