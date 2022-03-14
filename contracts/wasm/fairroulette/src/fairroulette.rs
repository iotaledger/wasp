// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This example implements 'fairroulette', a simple smart contract that can automatically handle
// an unlimited amount of bets on a number during a timed betting round. Once a betting round
// is over the contract will automatically pay out the winners proportionally to their bet amount.
// The intent is to showcase basic functionality of WasmLib and timed calling of functions
// through a minimal implementation and not to come up with a complete real-world solution.

use wasmlib::*;

use crate::*;
use crate::contract::*;
use crate::structs::*;

// Define some default configuration parameters.

// The maximum number one can bet on. The range of numbers starts at 1.
const MAX_NUMBER: u16 = 8;

// The default playing period of one betting round in seconds.
const DEFAULT_PLAY_PERIOD: u32 = 60;

// Enable this if you deploy the contract to an actual node. It will pay out the prize after a certain timeout.
const ENABLE_SELF_POST: bool = true;

// The number to divide nano seconds to seconds.
const NANO_TIME_DIVIDER: u64 = 1_000_000_000;

// 'placeBet' is used by betters to place a bet on a number from 1 to MAX_NUMBER. The first
// incoming bet triggers a betting round of configurable duration. After the playing period
// expires the smart contract will automatically pay out any winners and start a new betting
// round upon arrival of a new bet.
// The 'placeBet' function takes 1 mandatory parameter:
// - 'number', which must be an Int64 number from 1 to MAX_NUMBER
// The 'member' function will save the number together with the address of the better and
// the amount of incoming iotas as the bet amount in its state.
pub fn func_place_bet(ctx: &ScFuncContext, f: &PlaceBetContext) {
    // Get the array of current bets from state storage.
    let bets: ArrayOfMutableBet = f.state.bets();

    let nr_of_bets = bets.length();
    for i in 0..nr_of_bets {
        let bet: Bet = bets.get_bet(i).value();

        if bet.better.address() == ctx.caller().address() {
            ctx.panic("Bet already placed for this round");
        }
    }

    // Since we are sure that the 'number' parameter actually exists we can
    // retrieve its actual value into an u16.
    let number: u16 = f.params.number().value();

    // Require that the number is a valid number to bet on, otherwise panic out.
    ctx.require(number >= 1 && number <= MAX_NUMBER, "invalid number");

    // Create ScBalances proxy to the incoming balances for this request.
    // Note that ScBalances wraps an ScImmutableMap of token color/amount combinations
    // in a simpler to use interface.
    let allowance: ScBalances = ctx.allowance();

    // Retrieve the amount of plain iota tokens that are part of the incoming balance.
    let amount: u64 = allowance.balance(&ScColor::IOTA);

    // Require that there are actually some plain iotas there
    ctx.require(amount > 0, "empty bet");

    // Now we gather all information together into a single serializable struct
    // Note that we use the caller() method of the function context to determine
    // the agent id of the better. This is where a potential pay-out will be sent.
    let bet = Bet {
        better: ctx.caller(),
        amount: amount,
        number: number,
    };

    // Append the bet data to the bets array. The bet array will automatically take care
    // of serializing the bet struct into a bytes representation.
    bets.append_bet().set_value(&bet);

    f.events.bet(&bet.better.address(), bet.amount, bet.number);

    // Was this the first bet of this round?
    if nr_of_bets == 0 {
        // Yes it was, query the state for the length of the playing period in seconds by
        // retrieving the playPeriod value from state storage
        let mut play_period: u32 = f.state.play_period().value();

        // if the play period is less than 10 seconds we override it with the default duration.
        // Note that this will also happen when the play period was not set yet because in that
        // case a zero value was returned.
        if play_period < 10 {
            play_period = DEFAULT_PLAY_PERIOD;
            f.state.play_period().set_value(play_period);
        }

        if ENABLE_SELF_POST {
            f.state.round_status().set_value(1);

            // timestamp is nanotime, divide by NANO_TIME_DIVIDER to get seconds => common unix timestamp
            let timestamp = (ctx.timestamp() / NANO_TIME_DIVIDER) as u32;
            f.state.round_started_at().set_value(timestamp);

            f.events.start();

            let round_number = f.state.round_number();
            round_number.set_value(round_number.value() + 1);

            f.events.round(round_number.value());

            // And now for our next trick we post a delayed request to ourselves on the Tangle.
            // We are requesting to call the 'payWinners' function, but delay it for the play_period
            // amount of seconds. This will lock in the playing period, during which more bets can
            // be placed. Once the 'payWinners' function gets triggered by the ISCP it will gather all
            // bets up to that moment as the ones to consider for determining the winner.
            ScFuncs::pay_winners(ctx).func.delay(play_period).post();
        }
    }
}

// 'payWinners' is a function whose execution gets initiated by the 'placeBet' function.
// It collects a list of all bets, generates a random number, sorts out the winners and transfers
// the calculated winning sum to each attendee.
pub fn func_pay_winners(ctx: &ScFuncContext, f: &PayWinnersContext) {
    // Use the built-in random number generator which has been automatically initialized by
    // using the transaction hash as initial entropy data. Note that the pseudo-random number
    // generator will use the next 8 bytes from the hash as its random Int64 number and once
    // it runs out of data it simply hashes the previous hash for a next pseudo-random sequence.
    // Here we determine the winning number for this round in the range of 1 thru MAX_NUMBER.
    let winning_number: u16 = (ctx.random(MAX_NUMBER as u64 - 1) + 1) as u16;

    // Save the last winning number in state storage under 'lastWinningNumber' so that there
    // is (limited) time for people to call the 'getLastWinningNumber' View to verify the last
    // winning number if they wish. Note that this is just a silly example. We could log much
    // more extensive statistics information about each playing round in state storage and
    // make that data available through views for anyone to see.
    f.state.last_winning_number().set_value(winning_number);

    // Gather all winners and calculate some totals at the same time.
    // Keep track of the total bet amount, the total win amount, and all the winners.
    // Note how we decided to keep the winners in a local vector instead of creating
    // yet another array in state storage or having to go through lockedBets again.
    let mut total_bet_amount: u64 = 0_u64;
    let mut total_win_amount: u64 = 0_u64;
    let mut winners: Vec<Bet> = Vec::new();

    // Get the 'bets' array in state storage.
    let bets: ArrayOfMutableBet = f.state.bets();

    // Determine the amount of bets in the 'bets' array.
    let nr_of_bets: u32 = bets.length();

    // Loop through all indexes of the 'bets' array.
    for i in 0..nr_of_bets {
        // Retrieve the bet stored at the next index
        let bet: Bet = bets.get_bet(i).value();

        // Add this bet's amount to the running total bet amount
        total_bet_amount += bet.amount;

        // Did this better bet on the winning number?
        if bet.number == winning_number {
            // Yes, add this bet amount to the running total win amount.
            total_win_amount += bet.amount;

            // And save this bet in the winners vector.
            winners.push(bet);
        }
    }

    // Now that we preprocessed all bets we can get rid of the data in state storage
    // so that the 'bets' array becomes available for when the next betting round ends.
    bets.clear();

    f.events.winner(winning_number);

    // Did we have any winners at all?
    if winners.is_empty() {
        // No winners, log this fact to the log on the host.
        ctx.log("Nobody wins!");
    }

    // Pay out the winners proportionally to their bet amount. Note that we could configure
    // a small percentage that would go to the owner of the smart contract as hosting payment.

    // Keep track of the total payout so we can calculate the remainder after truncation.
    let mut total_payout: u64 = 0_u64;

    // Loop through all winners.
    let size: usize = winners.len();
    for i in 0..size {
        // Get the next winner.
        let bet: &Bet = &winners[i];

        // Determine the proportional win amount (we could take our percentage here)
        let payout: u64 = total_bet_amount * bet.amount / total_win_amount;

        // Anything to pay to the winner?
        if payout != 0 {
            // Yep, keep track of the running total payout
            total_payout += payout;

            // Set up an ScTransfers proxy that transfers the correct amount of iotas.
            // Note that ScTransfers wraps an ScMutableMap of token color/amount combinations
            // in a simpler to use interface. The constructor we use here creates and initializes
            // a single token color transfer in a single statement. The actual color and amount
            // values passed in will be stored in a new map on the host.
            let transfers: ScTransfers = ScTransfers::iotas(payout);

            // Perform the actual transfer of tokens from the smart contract to the address
            // of the winner. The transfer_to_address() method receives the address value and
            // the proxy to the new transfers map on the host, and will call the corresponding
            // host sandbox function with these values.
            ctx.send(&bet.better.address(), &transfers);
        }

        // Announce who got sent what as event.
        f.events.payout(&bet.better.address(), payout);
    }

    // This is where we transfer the remainder after payout to the creator of the smart contract.
    // The bank always wins :-P
    let remainder: u64 = total_bet_amount - total_payout;
    if remainder != 0 {
        // We have a remainder. First create a transfer for the remainder.
        let transfers: ScTransfers = ScTransfers::iotas(remainder);

        // Send the remainder to the contract creator.
        ctx.send(&ctx.contract_creator().address(), &transfers);
    }

    // Set round status to 0, send out event to notify that the round has ended
    f.state.round_status().set_value(0);
    f.events.stop();
}

pub fn func_force_reset(_ctx: &ScFuncContext, f: &ForceResetContext) {

    // Get the 'bets' array in state storage.
    let bets: ArrayOfMutableBet = f.state.bets();

    // Clear all bets.
    bets.clear();

    // Set round status to 0, send out event to notify that the round has ended
    f.state.round_status().set_value(0);
    f.events.stop();
}

// 'playPeriod' can be used by the contract creator to set the length of a betting round
// to a different value than the default value, which is 120 seconds.
pub fn func_play_period(ctx: &ScFuncContext, f: &PlayPeriodContext) {
    // Since we are sure that the 'playPeriod' parameter actually exists we can
    // retrieve its actual value into an i32 value.
    let play_period: u32 = f.params.play_period().value();

    // Require that the play period (in seconds) is not ridiculously low.
    // Otherwise, panic out with an error message.
    ctx.require(play_period >= 10, "invalid play period");

    // Now we set the corresponding variable 'playPeriod' in state storage.
    f.state.play_period().set_value(play_period);
}

pub fn view_last_winning_number(_ctx: &ScViewContext, f: &LastWinningNumberContext) {
    // Get the 'last_winning_number' int64 value from state storage.
    let last_winning_number = f.state.last_winning_number().value();

    // Set the 'last_winning_number' in results to the value from state storage.
    f.results
        .last_winning_number()
        .set_value(last_winning_number);
}

pub fn view_round_number(_ctx: &ScViewContext, f: &RoundNumberContext) {
    // Get the 'round_number' int64 value from state storage.
    let round_number = f.state.round_number().value();

    // Set the 'round_number' in results to the value from state storage.
    f.results.round_number().set_value(round_number);
}

pub fn view_round_status(_ctx: &ScViewContext, f: &RoundStatusContext) {
    // Get the 'roundStatus' int16 value from state storage.
    let round_status = f.state.round_status().value();

    // Set the 'round_status' in results to the value from state storage.
    f.results.round_status().set_value(round_status);
}

pub fn view_round_started_at(_ctx: &ScViewContext, f: &RoundStartedAtContext) {
    // Get the 'round_started_at' int32 value from state storage.
    let round_started_at = f.state.round_started_at().value();

    // Set the 'round_started_at' in results to the value from state storage.
    f.results.round_started_at().set_value(round_started_at);
}

pub fn func_force_payout(ctx: &ScFuncContext, _f: &ForcePayoutContext) {
    ScFuncs::pay_winners(ctx).func.call();
}
