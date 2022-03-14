// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This example implements 'fairroulette', a simple smart contract that can automatically handle
// an unlimited amount of bets on a number during a timed betting round. Once a betting round
// is over the contract will automatically pay out the winners proportionally to their bet amount.
// The intent is to showcase basic functionality of WasmLib and timed calling of functions
// through a minimal implementation and not to come up with a complete real-world solution.

import * as wasmlib from "wasmlib"
import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "./index";

// Define some default configuration parameters.

// The maximum number one can bet on. The range of numbers starts at 1.
const MAX_NUMBER: u16 = 8;

// The default playing period of one betting round in seconds.
const DEFAULT_PLAY_PERIOD: u32 = 60;

// Enable this if you deploy the contract to an actual node. It will pay out the prize after a certain timeout.
const ENABLE_SELF_POST: boolean = true;

// The number to divide nano seconds to seconds.
const NANO_TIME_DIVIDER: u64 = 1000000000;

// 'placeBet' is used by betters to place a bet on a number from 1 to MAX_NUMBER. The first
// incoming bet triggers a betting round of configurable duration. After the playing period
// expires the smart contract will automatically pay out any winners and start a new betting
// round upon arrival of a new bet.
// The 'placeBet' function takes 1 mandatory parameter:
// - 'number', which must be an Int64 number from 1 to MAX_NUMBER
// The 'member' function will save the number together with the address of the better and
// the amount of incoming iotas as the bet amount in its state.
export function funcPlaceBet(ctx: wasmlib.ScFuncContext, f: sc.PlaceBetContext): void {
    // Get the array of current bets from state storage.
    let bets: sc.ArrayOfMutableBet = f.state.bets();

    const nrOfBets = bets.length();
    for (let i: u32 = 0; i < nrOfBets; i++) {
        let bet: sc.Bet = bets.getBet(i).value();

        if (bet.better.address() == ctx.caller().address()) {
            ctx.panic("Bet already placed for this round");
        }
    }

    // Since we are sure that the 'number' parameter actually exists we can
    // retrieve its actual value into an i64.
    let number: u16 = f.params.number().value();

    // Require that the number is a valid number to bet on, otherwise panic out.
    ctx.require(number >= 1 && number <= MAX_NUMBER, "invalid number");

    // Create ScBalances proxy to the incoming balances for this request.
    // Note that ScBalances wraps an ScImmutableMap of token color/amount combinations
    // in a simpler to use interface.
    let incoming: wasmlib.ScBalances = ctx.incoming();

    // Retrieve the amount of plain iota tokens that are part of the incoming balance.
    let amount: u64 = incoming.balance(wasmtypes.IOTA);

    // Require that there are actually some plain iotas there
    ctx.require(amount > 0, "empty bet");

    // Now we gather all information together into a single serializable struct
    // Note that we use the caller() method of the function context to determine
    // the agent id of the better. This is where a potential pay-out will be sent.
    let bet = new sc.Bet();
    bet.better = ctx.caller();
    bet.amount = amount;
    bet.number = number;

    // Append the bet data to the bets array. The bet array will automatically take care
    // of serializing the bet struct into a bytes representation.
    bets.appendBet().setValue(bet);

    f.events.bet(bet.better.address(), bet.amount, bet.number);

    // Was this the first bet of this round?
    if (nrOfBets == 0) {
        // Yes it was, query the state for the length of the playing period in seconds by
        // retrieving the playPeriod value from state storage
        let playPeriod: u32 = f.state.playPeriod().value();

        // if the play period is less than 10 seconds we override it with the default duration.
        // Note that this will also happen when the play period was not set yet because in that
        // case a zero value was returned.
        if (playPeriod < 10) {
            playPeriod = DEFAULT_PLAY_PERIOD;
            f.state.playPeriod().setValue(playPeriod);
        }

        if (ENABLE_SELF_POST) {
            f.state.roundStatus().setValue(1);

            // timestamp is nanotime, divide by NANO_TIME_DIVIDER to get seconds => common unix timestamp
            let timestamp = (ctx.timestamp() / NANO_TIME_DIVIDER) as u32;
            f.state.roundStartedAt().setValue(timestamp);

            f.events.start();

            let roundNumber = f.state.roundNumber();
            roundNumber.setValue(roundNumber.value() + 1);

            f.events.round(roundNumber.value());

            // And now for our next trick we post a delayed request to ourselves on the Tangle.
            // We are requesting to call the 'payWinners' function, but delay it for the playPeriod
            // amount of seconds. This will lock in the playing period, during which more bets can
            // be placed. Once the 'payWinners' function gets triggered by the ISCP it will gather all
            // bets up to that moment as the ones to consider for determining the winner.
            sc.ScFuncs.payWinners(ctx).func.delay(playPeriod).post();
        }
    }
}

// 'payWinners' is a function whose execution gets initiated by the 'placeBet' function.
// It collects a list of all bets, generates a random number, sorts out the winners and transfers
// the calculated winning sum to each attendee.
export function funcPayWinners(ctx: wasmlib.ScFuncContext, f: sc.PayWinnersContext): void {
    // Use the built-in random number generator which has been automatically initialized by
    // using the transaction hash as initial entropy data. Note that the pseudo-random number
    // generator will use the next 8 bytes from the hash as its random Int64 number and once
    // it runs out of data it simply hashes the previous hash for a next pseudo-random sequence.
    // Here we determine the winning number for this round in the range of 1 thru MAX_NUMBER.
    let winningNumber: u16 = (ctx.random(MAX_NUMBER - 1) + 1) as u16;

    // Save the last winning number in state storage under 'lastWinningNumber' so that there
    // is (limited) time for people to call the 'getLastWinningNumber' View to verify the last
    // winning number if they wish. Note that this is just a silly example. We could log much
    // more extensive statistics information about each playing round in state storage and
    // make that data available through views for anyone to see.
    f.state.lastWinningNumber().setValue(winningNumber);

    // Gather all winners and calculate some totals at the same time.
    // Keep track of the total bet amount, the total win amount, and all the winners.
    // Note how we decided to keep the winners in a local vector instead of creating
    // yet another array in state storage or having to go through lockedBets again.
    let totalBetAmount: u64 = 0;
    let totalWinAmount: u64 = 0;
    let winners: sc.Bet[] = [];

    // Get the 'bets' array in state storage.
    let bets: sc.ArrayOfMutableBet = f.state.bets();

    // Determine the amount of bets in the 'bets' array.
    let nrOfBets: u32 = bets.length();

    // Loop through all indexes of the 'bets' array.
    for (let i: u32 = 0; i < nrOfBets; i++) {
        // Retrieve the bet stored at the next index
        let bet: sc.Bet = bets.getBet(i).value();

        // Add this bet's amount to the running total bet amount
        totalBetAmount += bet.amount;

        // Did this better bet on the winning number?
        if (bet.number == winningNumber) {
            // Yes, add this bet amount to the running total win amount.
            totalWinAmount += bet.amount;

            // And save this bet in the winners vector.
            winners.push(bet);
        }
    }

    // Now that we preprocessed all bets we can get rid of the data in state storage
    // so that the 'bets' array becomes available for when the next betting round ends.
    bets.clear();

    f.events.winner(winningNumber);

    // Did we have any winners at all?
    if (winners.length == 0) {
        // No winners, log this fact to the log on the host.
        ctx.log("Nobody wins!");
    }

    // Pay out the winners proportionally to their bet amount. Note that we could configure
    // a small percentage that would go to the owner of the smart contract as hosting payment.

    // Keep track of the total payout so we can calculate the remainder after truncation.
    let totalPayout: u64 = 0;

    // Loop through all winners.
    let size = winners.length;
    for (let i = 0; i < size; i++) {
        // Get the next winner.
        let bet: sc.Bet = winners[i];

        // Determine the proportional win amount (we could take our percentage here)
        let payout: u64 = totalBetAmount * bet.amount / totalWinAmount;

        // Anything to pay to the winner?
        if (payout != 0) {
            // Yep, keep track of the running total payout
            totalPayout += payout;

            // Set up an ScTransfers proxy that transfers the correct amount of iotas.
            // Note that ScTransfers wraps an ScMutableMap of token color/amount combinations
            // in a simpler to use interface. The constructor we use here creates and initializes
            // a single token color transfer in a single statement. The actual color and amount
            // values passed in will be stored in a new map on the host.
            let transfers: wasmlib.ScTransfers = wasmlib.ScTransfers.iotas(payout);

            // Perform the actual transfer of tokens from the smart contract to the address
            // of the winner. The transferToAddress() method receives the address value and
            // the proxy to the new transfers map on the host, and will call the corresponding
            // host sandbox function with these values.
            ctx.send(bet.better.address(), transfers);
        }

        // Announce who got sent what as event.
        f.events.payout(bet.better.address(), payout);
    }

    // This is where we transfer the remainder after payout to the creator of the smart contract.
    // The bank always wins :-P
    let remainder: u64 = totalBetAmount - totalPayout;
    if (remainder != 0) {
        // We have a remainder. First create a transfer for the remainder.
        let transfers: wasmlib.ScTransfers = wasmlib.ScTransfers.iotas(remainder);

        // Send the remainder to the contract creator.
        ctx.send(ctx.contractCreator().address(), transfers);
    }

    // Set round status to 0, send out event to notify that the round has ended
    f.state.roundStatus().setValue(0);
    f.events.stop();
}

export function funcForceReset(ctx: wasmlib.ScFuncContext, f: sc.ForceResetContext): void {

    // Get the 'bets' array in state storage.
    let bets: sc.ArrayOfMutableBet = f.state.bets();

    // Clear all bets.
    bets.clear();

    // Set round status to 0, send out event to notify that the round has ended
    f.state.roundStatus().setValue(0);
    f.events.stop();
}

// 'playPeriod' can be used by the contract creator to set the length of a betting round
// to a different value than the default value, which is 120 seconds.
export function funcPlayPeriod(ctx: wasmlib.ScFuncContext, f: sc.PlayPeriodContext): void {
    // Since we are sure that the 'playPeriod' parameter actually exists we can
    // retrieve its actual value into an i32 value.
    let playPeriod: u32 = f.params.playPeriod().value();

    // Require that the play period (in seconds) is not ridiculously low.
    // Otherwise, panic out with an error message.
    ctx.require(playPeriod >= 10, "invalid play period");

    // Now we set the corresponding variable 'playPeriod' in state storage.
    f.state.playPeriod().setValue(playPeriod);
}

export function viewLastWinningNumber(ctx: wasmlib.ScViewContext, f: sc.LastWinningNumberContext): void {
    // Get the 'lastWinningNumber' int64 value from state storage.
    let lastWinningNumber = f.state.lastWinningNumber().value();

    // Set the 'lastWinningNumber' in results to the value from state storage.
    f.results
        .lastWinningNumber()
        .setValue(lastWinningNumber);
}

export function viewRoundNumber(ctx: wasmlib.ScViewContext, f: sc.RoundNumberContext): void {
    // Get the 'roundNumber' uint32 value from state storage.
    let roundNumber = f.state.roundNumber().value();

    // Set the 'roundNumber' in results to the value from state storage.
    f.results.roundNumber().setValue(roundNumber);
}

export function viewRoundStatus(ctx: wasmlib.ScViewContext, f: sc.RoundStatusContext): void {
    // Get the 'roundStatus' uint16 value from state storage.
    let roundStatus = f.state.roundStatus().value();

    // Set the 'roundStatus' in results to the value from state storage.
    f.results.roundStatus().setValue(roundStatus);
}

export function viewRoundStartedAt(ctx: wasmlib.ScViewContext, f: sc.RoundStartedAtContext): void {
    // Get the 'roundStartedAt' uint32 value from state storage.
    let roundStartedAt = f.state.roundStartedAt().value();

    // Set the 'roundStartedAt' in results to the value from state storage.
    f.results.roundStartedAt().setValue(roundStartedAt);
}

export function funcForcePayout(ctx: wasmlib.ScFuncContext, f: sc.ForcePayoutContext): void {
    sc.ScFuncs.payWinners(ctx).func.call();
}
