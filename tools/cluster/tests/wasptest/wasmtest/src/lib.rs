use wasp::ScContext;

mod wasp;
mod utils;

enum RequestCode {
    RequestNop = 1,
    RequestInc,
    RequestIncRepeat1,
    RequestIncMany,
    RequestPlaceBet,
    RequestLockBets,
    RequestPayWinners,
    RequestPlayPeriod = 0x4008,
}

const NUM_COLORS: i64 = 5;
const PLAY_PERIOD: i64 = 120;

struct BetInfo {
    req_hash: String,
    sender: String,
    color: i64,
    amount: i64,
}

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

#[no_mangle]
pub fn no_op() {
    let ctx = ScContext::new();
    ctx.log("Doing nothing as requested. Oh, wait...");
}

#[no_mangle]
pub fn increment() {
    let ctx = ScContext::new();
    ctx.log("Increment...");
    let counter = ctx.state().get_int("counter");
    counter.set_value(counter.value() + 1);
}

#[no_mangle]
pub fn incrementRepeat1() {
    let ctx = ScContext::new();
    ctx.log("incrementRepeat1...");
    let counter = ctx.state().get_int("counter");
    let value = counter.value();
    counter.set_value(value + 1);
    if value == 0 {
        let request = ctx.requests().get_map(0);
        request.get_int("reqCode").set_value(RequestCode::RequestInc as i64);
        request.get_int("reqDelay").set_value(5);
    }
}

#[no_mangle]
pub fn incrementRepeatMany() {
    let ctx = ScContext::new();
    ctx.log("incrementRepeatMany...");
    let counter = ctx.state().get_int("counter");
    let value = counter.value();
    counter.set_value(value + 1);
    let mut repeats = ctx.params().get_int("numrepeats").value();
    let state_repeats = ctx.state().get_int("numrepeats");
    if repeats == 0 {
        repeats = state_repeats.value();
        if repeats == 0 {
            return;
        }
    }
    state_repeats.set_value(repeats - 1);
    let request = ctx.requests().get_map(0);
    request.get_int("reqCode").set_value(RequestCode::RequestIncMany as i64);
    request.get_int("reqDelay").set_value(3);
}

#[no_mangle]
pub fn placeBet() {
    let ctx = ScContext::new();
    ctx.log("Place bet...");
    let amount = ctx.request_balance("iota");
    if amount == 0 {
        ctx.log("Empty bet...");
        return;
    }
    let color = ctx.params().get_int("color").value();
    if color == 0 {
        ctx.log("No color...");
        return;
    }
    if color < 1 || color > NUM_COLORS {
        ctx.log("Invalid color...");
        return;
    }

    let bet = BetInfo {
        req_hash: ctx.request_hash(),
        sender: ctx.sender(),
        color: color,
        amount: amount,
    };

    let state = ctx.state();
    let bets = state.get_string_array("bets");
    let bet_nr = bets.length();
    let bet_data = bet_to_string(&bet);
    bets.get_string(bet_nr).set_value(&bet_data);
    if bet_nr == 0 {
        let mut play_period = state.get_int("playPeriod").value();
        if play_period < 10 {
            play_period = PLAY_PERIOD;
        }
        let request = ctx.requests().get_map(0);
        request.get_int("reqCode").set_value(RequestCode::RequestLockBets as i64);
        request.get_int("reqDelay").set_value(play_period);
    }
}

#[no_mangle]
pub fn lockBets() {
    let ctx = ScContext::new();
    ctx.log("Lock bets...");

    // can only be sent by SC itself
    if ctx.sender() != ctx.sc_address() {
        ctx.log("Cancel spoofed request");
        return;
    }

    let state = ctx.state();
    let bets = state.get_string_array("bets");
    let locked_bets = state.get_string_array("lockedBets");
    for i in 0..bets.length() {
        let bet = bets.get_string(i).value();
        locked_bets.get_string(i).set_value(&bet);
    }
    bets.clear();

    let request = ctx.requests().get_map(0);
    request.get_int("reqCode").set_value(RequestCode::RequestPayWinners as i64);
}

#[no_mangle]
pub fn payWinners() {
    let ctx = ScContext::new();
    ctx.log("Pay winners...");

    // can only be sent by SC itself
    let sc_address = ctx.sc_address();
    if ctx.sender() != sc_address {
        ctx.log("Cancel spoofed request");
        return;
    }

    let winning_color = ctx.random(5) + 1;
    let state = ctx.state();
    state.get_int("lastWinningColor").set_value(winning_color);

    let mut total_bet_amount: i64 = 0;
    let mut total_win_amount: i64 = 0;
    let locked_bets = state.get_string_array("lockedBets");
    let mut winners: Vec<BetInfo> = Vec::new();
    for i in 0..locked_bets.length() {
        let bet_data = locked_bets.get_string(i).value();
        let bet = string_to_bet(&bet_data);
        total_bet_amount += bet.amount;
        if bet.color == winning_color {
            total_win_amount += bet.amount;
            winners.push(bet);
        }
    }
    locked_bets.clear();

    if winners.is_empty() {
        ctx.log("Nobody wins!");
        // compact separate UTXOs into a single one
        let transfer = ctx.transfers().get_map(0);
        transfer.get_string("xferAddress").set_value(&sc_address);
        transfer.get_int("xferAmount").set_value(total_bet_amount);
        return;
    }

    let mut total_payout: i64 = 0;
    for i in 0..winners.len() {
        let bet = &winners[i];
        let payout = total_bet_amount * bet.amount / total_win_amount;
        if payout != 0 {
            total_payout += payout;
            let transfer = ctx.transfers().get_map(i as i32);
            transfer.get_string("xferAddress").set_value(&bet.sender);
            transfer.get_int("xferAmount").set_value(payout);
        }
        let text = "Pay ".to_string() + &payout.to_string() + " to " + &bet.sender;
        ctx.log(&text);
    }

    if total_payout != total_bet_amount {
        let remainder = total_bet_amount - total_payout;
        let text = "Remainder is ".to_string() + &remainder.to_string();
        ctx.log(&text);
    }
}

#[no_mangle]
pub fn playPeriod() {
    let ctx = ScContext::new();
    ctx.log("Play period...");

    // can only be sent by SC itself
    if ctx.sender() != ctx.owner() {
        ctx.log("Cancel spoofed request");
        return;
    }

    let play_period = ctx.params().get_int("playPeriod").value();
    if play_period < 10 {
        ctx.log("Invalid play period...");
        return;
    }

    ctx.state().get_int("playPeriod").set_value(play_period);
}


struct TokenInfo {
    supply: i64,
    mintedBy: String,
    owner: String,
    created: i64,
    updated: i64,
    description: String,
    user_defined: String,
}

#[no_mangle]
pub fn tokenMint() {
    let ctx = ScContext::new();
    ctx.log("Token mint...");
    //TBD
}


fn bet_to_string(bet: &BetInfo) -> String {
    String::new() +
        &bet.req_hash + "|" +
        &bet.sender + "|" +
        &bet.color.to_string() + "|" +
        &bet.amount.to_string()
}

fn string_to_bet(data: &str) -> BetInfo {
    let parts: Vec<&str> = data.split("|").collect();
    BetInfo {
        req_hash: parts[0].to_string(),
        sender: parts[1].to_string(),
        color: parts[2].parse::<i64>().unwrap(),
        amount: parts[3].parse::<i64>().unwrap(),
    }
}
