// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This example implements 'fairroulette', a simple smart contract that can automatically handle
// an unlimited amount of bets bets on a number during a timed betting round. Once a betting round
// is over the contract will automatically pay out the winners proportionally to their bet amount.
// The intent is to showcase basic functionality of WasmLib and timed calling of functions
// through a minimal implementation and not to come up with a complete real-world solution.

use wasmlib::*;

use crate::*;
use crate::types::*;

// define some default configuration parameters

// the maximum number one can bet on. The range of numbers starts at 1.
const MAX_NUMBER: i64 = 5;
// the default playing period of one betting round in minutes
const DEFAULT_PLAY_PERIOD: i64 = 120;

// 'placeBet' is used by betters to place a bet on a number from 1 to MAX_NUMBER. The first
// incoming bet triggers a betting round of configurable duration. After the playing period
// expires the smart contract will automatically pay out any winners and start a new betting
// round upon arrival of a new bet.
// The 'placeBet' function takes 1 mandatory parameter:
// - 'number', which must be s an Int64 number from 1 to MAX_NUMBER
// The 'member' function will save the number together with the address of the better and the amount
// of incoming iotas as the bet amount in its state.
pub fn func_place_bet(ctx: &ScFuncContext) {

    // Log the fact that we have initiated the 'placeBet' Func in the log on the host.
    ctx.log("fairroulette.placeBet");

    // Now it is time to check the parameter.
    // First we create an ScImmutableMap proxy to the params map on the host.
    let p = ctx.params();

    // Create an ScImmutablInt64 proxy to the 'number' parameter that is still stored
    // in the params map on the host using a constant defined in consts.rs.
    let param_number = p.get_int64(PARAM_NUMBER);

    // Require that the mandatory 'number' parameter actually exists in the map on the host.
    // If it doesn't we panic out with an error message.
    ctx.require(param_number.exists(), "missing mandatory number");

    // Now that we are sure that the 'number' parameter actually exists we can
    // retrieve its actual value into an i64.
    let number = param_number.value();
    // require that the number is a valid number to bet on, otherwise panic out.
    ctx.require(number >= 1 && number <= MAX_NUMBER, "invalid number");

    // Create ScBalances proxy to the incoming balances for this Func request.
    // Note that ScBalances wraps an ScImmutableMap of token color/amount combinations
    // in a simpler to use interface.
    let incoming = ctx.incoming();

    // Retrieve the amount of plain iota tokens from the incoming balance
    let amount = incoming.balance(&ScColor::IOTA);

    // require that there are actually some iotas there
    ctx.require(amount > 0, "empty bet");

    // Now we gather all information together into a single serializable struct
    // Note that we use the caller() method of the function context to determine
    // the address of the better. This is the address where a pay-out will be sent.
    let bet = Bet {
        better: ctx.caller(),
        amount: amount,
        number: number,
    };

    // Create an ScMutableMap proxy to the state storage map on the host.
    let state = ctx.state();

    // Create an ScMutableBytesArray proxy to a bytes array named "bets" in the state storage.
    let bets = state.get_bytes_array(VAR_BETS);

    // Determine what the next bet number is by retrieving the length of the bets array.
    let bet_nr = bets.length();

    // Append the bet data to the bets array. We get an ScBytes proxy to the bytes stored
    // using the bet number as index. Then we set the bytes value in the best array on the
    // host to the result of serializing the bet data into a bytes representation.
    bets.get_bytes(bet_nr).set_value(&bet.to_bytes());

    // Was this the first bet of this round?
    if bet_nr == 0 {
        // Yes it was, query the state for the length of the playing period in minutes by
        // retrieving the "playPeriod" from state storage
        let mut play_period = state.get_int64(VAR_PLAY_PERIOD).value();

        // if the play period is less than 10 minutes we override it with the default duration.
        // Note that this will also happen when the play periosd was not set yet because in that
        // case a zero value was returned.
        if play_period < 10 {
            play_period = DEFAULT_PLAY_PERIOD;
        }

        // And now for our next trick we post a request to
        ctx.post_self(HFUNC_LOCK_BETS, None, None, play_period);
    }
    ctx.log("fairroulette.placeBet ok");
}

pub fn func_lock_bets(ctx: &ScFuncContext) {
    ctx.log("fairroulette.lockBets");
    // only SC itself can invoke this function
    ctx.require(ctx.caller() == ctx.contract_id().as_agent_id(), "no permission");

    // move all current bets to the locked_bets array
    let state = ctx.state();
    let bets = state.get_bytes_array(VAR_BETS);
    let locked_bets = state.get_bytes_array(VAR_LOCKED_BETS);
    let nr_bets = bets.length();
    for i in 0..nr_bets {
        let bytes = bets.get_bytes(i).value();
        locked_bets.get_bytes(i).set_value(&bytes);
    }
    bets.clear();

    ctx.post_self(HFUNC_PAY_WINNERS, None, None, 0);
    ctx.log("fairroulette.lockBets ok");
}

pub fn func_pay_winners(ctx: &ScFuncContext) {
    ctx.log("fairroulette.payWinners");
    // only SC itself can invoke this function
    ctx.require(ctx.caller() == ctx.contract_id().as_agent_id(), "no permission");

    let sc_id = ctx.contract_id().as_agent_id();
    let winning_number = ctx.utility().random(5) + 1;
    let state = ctx.state();
    state.get_int64(VAR_LAST_WINNING_NUMBER).set_value(winning_number);

    // gather all winners and calculate some totals
    let mut total_bet_amount = 0_i64;
    let mut total_win_amount = 0_i64;
    let locked_bets = state.get_bytes_array(VAR_LOCKED_BETS);
    let mut winners: Vec<Bet> = Vec::new();
    let nr_bets = locked_bets.length();
    for i in 0..nr_bets {
        let bet = Bet::from_bytes(&locked_bets.get_bytes(i).value());
        total_bet_amount += bet.amount;
        if bet.number == winning_number {
            total_win_amount += bet.amount;
            winners.push(bet);
        }
    }
    locked_bets.clear();

    if winners.is_empty() {
        ctx.log("Nobody wins!");
        // compact separate bet deposit UTXOs into a single one
        ctx.transfer_to_address(&sc_id.address(), ScTransfers::new(&ScColor::IOTA, total_bet_amount));
        return;
    }

    // pay out the winners proportionally to their bet amount
    let mut total_payout = 0_i64;
    let size = winners.len();
    for i in 0..size {
        let bet = &winners[i];
        let payout = total_bet_amount * bet.amount / total_win_amount;
        if payout != 0 {
            total_payout += payout;
            ctx.transfer_to_address(&bet.better.address(), ScTransfers::new(&ScColor::IOTA, payout));
        }
        let text = "Pay ".to_string() + &payout.to_string() +
            " to " + &bet.better.to_string();
        ctx.log(&text);
    }

    // any truncation left-overs are fair picking for the smart contract
    if total_payout != total_bet_amount {
        let remainder = total_bet_amount - total_payout;
        let text = "Remainder is ".to_string() + &remainder.to_string();
        ctx.log(&text);
        ctx.transfer_to_address(&sc_id.address(), ScTransfers::new(&ScColor::IOTA, remainder));
    }
    ctx.log("fairroulette.payWinners ok");
}

pub fn func_play_period(ctx: &ScFuncContext) {
    ctx.log("fairroulette.playPeriod");
    // only SC creator can update the play period
    ctx.require(ctx.caller() == ctx.contract_creator(), "no permission");

    let p = ctx.params();
    let param_play_period = p.get_int64(PARAM_PLAY_PERIOD);

    ctx.require(param_play_period.exists(), "missing mandatory playPeriod");

    let play_period = param_play_period.value();
    if play_period < 10 {
        ctx.panic("Invalid play period...");
    }

    ctx.state().get_int64(VAR_PLAY_PERIOD).set_value(play_period);
    ctx.log("fairroulette.playPeriod ok");
}
