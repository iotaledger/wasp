// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "./index";

const DURATION_DEFAULT: u32 = 60;
const DURATION_MIN: u32 = 1;
const DURATION_MAX: u32 = 120;
const MAX_DESCRIPTION_LENGTH: i32 = 150;
const OWNER_MARGIN_DEFAULT: u64 = 50;
const OWNER_MARGIN_MIN: u64 = 5;
const OWNER_MARGIN_MAX: u64 = 100;

export function funcFinalizeAuction(ctx: wasmlib.ScFuncContext, f: sc.FinalizeAuctionContext): void {
    let token = f.params.token().value();
    let currentAuction = f.state.auctions().getAuction(token);
    ctx.require(currentAuction.exists(), "Missing auction info");
    let auction = currentAuction.value();
    if (auction.highestBid == 0) {
        ctx.log("No one bid on " + token.toString());
        let ownerFee = auction.minimumBid * auction.ownerMargin / 1000;
        if (ownerFee == 0) {
            ownerFee = 1;
        }
        // finalizeAuction request token was probably not confirmed yet
        transferIotas(ctx, ctx.contractCreator(), ownerFee - 1);
        transferTokens(ctx, auction.creator, auction.token, auction.numTokens);
        transferIotas(ctx, auction.creator, auction.deposit - ownerFee);
        return;
    }

    let ownerFee = auction.highestBid * auction.ownerMargin / 1000;
    if (ownerFee == 0) {
        ownerFee = 1;
    }

    // return staked bids to losers
    let bids = f.state.bids().getBids(token);
    let bidderList = f.state.bidderList().getBidderList(token);
    let size = bidderList.length();
    for (let i: u32 = 0; i < size; i++) {
        let loser = bidderList.getAgentID(i).value();
        if (!loser.equals(auction.highestBidder)) {
            let bid = bids.getBid(loser).value();
            transferIotas(ctx, loser, bid.amount);
        }
    }

    // finalizeAuction request token was probably not confirmed yet
    transferIotas(ctx, ctx.contractCreator(), ownerFee - 1);
    transferTokens(ctx, auction.highestBidder, auction.token, auction.numTokens);
    transferIotas(ctx, auction.creator, auction.deposit + auction.highestBid - ownerFee);
}

export function funcPlaceBid(ctx: wasmlib.ScFuncContext, f: sc.PlaceBidContext): void {
    let bidAmount = ctx.allowance().iotas();
    ctx.require(bidAmount > 0, "Missing bid amount");

    let token = f.params.token().value();
    let currentAuction = f.state.auctions().getAuction(token);
    ctx.require(currentAuction.exists(), "Missing auction info");

    let auction = currentAuction.value();
    let bids = f.state.bids().getBids(token);
    let bidderList = f.state.bidderList().getBidderList(token);
    let caller = ctx.caller();
    let currentBid = bids.getBid(caller);
    if (currentBid.exists()) {
        ctx.log("Upped bid from: " + caller.toString());
        let bid = currentBid.value();
        bidAmount += bid.amount;
        bid.amount = bidAmount;
        bid.timestamp = ctx.timestamp();
        currentBid.setValue(bid);
    } else {
        ctx.require(bidAmount >= auction.minimumBid, "Insufficient bid amount");
        ctx.log("New bid from: " + caller.toString());
        let index = bidderList.length();
        bidderList.appendAgentID().setValue(caller);
        let bid = new sc.Bid();
        bid.index = index;
        bid.amount = bidAmount;
        bid.timestamp = ctx.timestamp();
        currentBid.setValue(bid);
    }
    if (bidAmount > auction.highestBid) {
        ctx.log("New highest bidder");
        auction.highestBid = bidAmount;
        auction.highestBidder = caller;
        currentAuction.setValue(auction);
    }
}

export function funcSetOwnerMargin(ctx: wasmlib.ScFuncContext, f: sc.SetOwnerMarginContext): void {
    let ownerMargin = f.params.ownerMargin().value();
    if (ownerMargin < OWNER_MARGIN_MIN) {
        ownerMargin = OWNER_MARGIN_MIN;
    }
    if (ownerMargin > OWNER_MARGIN_MAX) {
        ownerMargin = OWNER_MARGIN_MAX;
    }
    f.state.ownerMargin().setValue(ownerMargin);
}

export function funcStartAuction(ctx: wasmlib.ScFuncContext, f: sc.StartAuctionContext): void {
    let token = f.params.token().value();
    let numTokens = ctx.allowance().balance(token);
    if (numTokens.isZero()) {
        ctx.panic("Missing auction tokens");
    }

    let minimumBid = f.params.minimumBid().value();

    // duration in minutes
    let duration = f.params.duration().value();
    if (duration == 0) {
        duration = DURATION_DEFAULT;
    }
    if (duration < DURATION_MIN) {
        duration = DURATION_MIN;
    }
    if (duration > DURATION_MAX) {
        duration = DURATION_MAX;
    }

    let description = f.params.description().value();
    if (description == "") {
        description = "N/A".toString();
    }
    if (description.length > MAX_DESCRIPTION_LENGTH) {
        description = description.slice(0,MAX_DESCRIPTION_LENGTH) + "[...]";
    }

    let ownerMargin = f.state.ownerMargin().value();
    if (ownerMargin == 0) {
        ownerMargin = OWNER_MARGIN_DEFAULT;
    }

    // need at least 1 iota to run SC
    let margin = minimumBid * ownerMargin / 1000;
    if (margin == 0) {
        margin = 1;
    }
    let deposit = ctx.allowance().iotas();
    if (deposit < margin) {
        ctx.panic("Insufficient deposit");
    }

    let currentAuction = f.state.auctions().getAuction(token);
    if (currentAuction.exists()) {
        ctx.panic("Auction for this token token already exists");
    }

    let auction = new sc.Auction();
    auction.creator = ctx.caller();
    auction.token = token;
    auction.deposit = deposit;
    auction.description = description;
    auction.duration = duration;
    auction.highestBid = 0;
    auction.highestBidder = wasmtypes.agentIDFromBytes([]);
    auction.minimumBid = minimumBid;
    auction.numTokens = numTokens.uint64();
    auction.ownerMargin = ownerMargin;
    auction.whenStarted = ctx.timestamp();
    currentAuction.setValue(auction);

    let fa = sc.ScFuncs.finalizeAuction(ctx);
    fa.params.token().setValue(auction.token);
    fa.func.delay(duration * 60).post();
}

export function viewGetInfo(ctx: wasmlib.ScViewContext, f: sc.GetInfoContext): void {
    let token = f.params.token().value();
    let currentAuction = f.state.auctions().getAuction(token);
    ctx.require(currentAuction.exists(), "Missing auction info");

    let auction = currentAuction.value();
    f.results.token().setValue(auction.token);
    f.results.creator().setValue(auction.creator);
    f.results.deposit().setValue(auction.deposit);
    f.results.description().setValue(auction.description);
    f.results.duration().setValue(auction.duration);
    f.results.highestBid().setValue(auction.highestBid);
    f.results.highestBidder().setValue(auction.highestBidder);
    f.results.minimumBid().setValue(auction.minimumBid);
    f.results.numTokens().setValue(auction.numTokens);
    f.results.ownerMargin().setValue(auction.ownerMargin);
    f.results.whenStarted().setValue(auction.whenStarted);

    let bidderList = f.state.bidderList().getBidderList(token);
    f.results.bidders().setValue(bidderList.length());
}

function transferIotas(ctx: wasmlib.ScFuncContext, agent: wasmlib.ScAgentID, amount: u64): void {
    if (agent.isAddress()) {
        // send back to original Tangle address
        ctx.send(agent.address(), wasmlib.ScTransfer.iotas(amount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(agent.address(), wasmlib.ScTransfer.iotas(amount));
}

function transferTokens(ctx: wasmlib.ScFuncContext, agent: wasmlib.ScAgentID, token: wasmlib.ScTokenID, amount: u64): void {
    const bigAmount = wasmtypes.ScBigInt.fromUint64(amount);
    if (agent.isAddress()) {
        // send back to original Tangle address
        ctx.send(agent.address(), wasmlib.ScTransfer.tokens(token, bigAmount));
        return;
    }

    // TODO not an address, deposit into account on chain
    ctx.send(agent.address(), wasmlib.ScTransfer.tokens(token, bigAmount));
}
