// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This example implements 'dividend', a simple smart contract that can automatically disperse
// iota tokens which are sent to the contract to a group of member addresses according
// to predefined division factors. The intent is to showcase basic functionality of WasmLib
// through a minimal implementation and not to come up with a complete real-world solution.

use wasmlib::*;

use crate::*;

// 'member' is a function that can be used by the entity that created the 'dividend' smart
// contract on the chain to define the group of member addresses and dispersal factors prior
// to starting to send tokens to the smart contract's 'divide' function. The 'member' function
// takes 2 parameters, that are both required:
// - 'address', which is an Address to use as member in the group, and
// - 'factor', which is an Int64 relative dispersal factor
// The 'member' function will save the address/factor combination in its state and also calculate
// and store a running sum of all factors so that the 'divide' function can simply use these values
pub fn func_member(ctx: &ScFuncContext) {

    // Log the fact that we have initiated the 'member' Func in the log on the host.
    ctx.log("dividend.member");

    // Only the smart contract creator can add members, so we require that the caller agent id
    // is equal to the contract creator's agent id. Otherwise we panic out with an error message.
    ctx.require(ctx.caller() == ctx.contract_creator(), "no permission");

    // Now it is time to check the parameters.
    // First we create an ScImmutableMap proxy to the params map on the host.
    let p = ctx.params();

    // Create an ScImmutableAddress proxy to the 'address' parameter that is still stored
    // in the map on the host. Note that we use constants defined in consts.rs to prevent
    // typos in name strings. This is good practice and will save time in the long run.
    let param_address = p.get_address(PARAM_ADDRESS);

    // Require that the mandatory 'address' parameter actually exists in the map on the host.
    // If it doesn't we panic out with an error message.
    ctx.require(param_address.exists(), "missing mandatory address");

    // Now that we are sure that the 'address' parameter actually exists we can
    // retrieve its actual value into an ScAddress value object
    let address = param_address.value();

    // Create an ScImmutableInt64 proxy to the 'factor' parameter that is still stored
    // in the map on the host. Note how the get_xxx() method defines what type we expect.
    let param_factor = p.get_int64(PARAM_FACTOR);

    // Require that the mandatory 'factor' parameter actually exists in the map on the host.
    // If it doesn't we panic out with an error message.
    ctx.require(param_factor.exists(), "missing mandatory factor");

    // Now that we are sure that the 'factor' parameter actually exists we can
    // retrieve its actual value into an i64. Note that we use Rust's built-in
    // data types when manipulating Int64, String, or Bytes value objects.
    let factor = param_factor.value();

    // As an extra requirement we check that the 'factor' parameter value is not negative.
    // If it is, we panic out with an error message. Note how we use an if expression here.
    // We could have achieved the same in a single line by using the require() method instead:
    // ctx.require(factor >= 0, "negative factor");
    // Using the require() method reduces typing and enhances readability.
    if factor < 0 {
        ctx.panic("negative factor");
    }

    // Now that we have sorted out the parameters we will start using the state storage on the host.
    // First we create an ScMutableMap proxy to the state storage map on the host.
    let state = ctx.state();

    // We will store the address/factor combinations in a key/value sub-map inside the state map
    // Create an ScMutableMap proxy to a map named 'members' in the state storage. If there is no
    // 'members' map present yet this will automatically create an empty map on the host.
    let members = state.get_map(VAR_MEMBERS);

    // Now we create sn ScMutableInt64 proxy for the value stored in the 'members' map under the
    // key defined by the 'address' parameter we retrieved earlier.
    let current_factor = members.get_int64(&address);

    // Next we check to see if this key/value combination already exists in the 'members' map
    if !current_factor.exists() {
        // If it does not exist yet then we have to add this new address to the 'memberList' array
        // We create an ScMutableAddressArray proxy to an Address array named 'memberList' in the
        // state storage. Again, if the array did not exist yet is will automatically be created.
        let member_list = state.get_address_array(VAR_MEMBER_LIST);

        // Now we will append the new address to the memberList array.
        // First we determine the current length of the array.
        let length = member_list.length();

        // Next we create an ScMutableAddress proxy to the Address value that lives at that index
        // in the memberList array.
        let new_address = member_list.get_address(length);

        // And finally we do the actual appending of the new address to the array by telling the
        // proxy to set the value it refers to to the value of the 'address' parameter.
        new_address.set_value(&address);
    }

    // Create an ScMutable Int64 proxy named 'totalFactor' to the value in state storage.
    // Note that we don't care whether this value exists or not, because if it doesn't
    // WasmLib will act as if it has the default value of zero.
    let total_factor = state.get_int64(VAR_TOTAL_FACTOR);

    // Now we calculate the new running total sum of factors by first getting the current
    // value of 'totalFactor' from the state storage, then subtracting the current value
    // of the factor associated with the 'address' parameter, if any exists. Again, if the
    // associated value doesn't exist, WasmLib will assume zero. Finally we add the factor
    // retrieved from the parameters, resulting in the new totalFactor.
    let new_total_factor = total_factor.value() - current_factor.value() + factor;

    // Now we store the new totalFactor in the state storage
    total_factor.set_value(new_total_factor);

    // And we also store the factor from the parameters under the address from the parameters
    // in the state storage the proxy refers to
    current_factor.set_value(factor);

    // Finally, we log the fact that we have successfully completed execution
    // of the 'member' Func in the log on the host.
    ctx.log("dividend.member ok");
}

// 'divide' is a function that will take any iotas it receives and properly disperse them
// to the addresses in the member list according to the associated dispersion factors.
// Anyone can send iota tokens to this function and they will automatically be passed on
// to the member list. Note that this function does not deal with fractions. It simply
// truncates the calculated amount to the nearest lower integer and keeps any remaining
// iotas in its own account. They will be added to any next round of tokens received
// prior to calculation the new dispersion amounts.
pub fn func_divide(ctx: &ScFuncContext) {

    // Log the fact that we have initiated the 'divide' Func in the log on the host.
    ctx.log("dividend.divide");

    // Create ScBalances proxy to the total account balances for this smart contract
    // Note that ScBalances wraps an ScImmutableMap of token color/amount combinations
    // in a simpler to use interface. Note that we are using the balances() method
    // here instead of the incoming() method because there could still me some iotas
    // remaining in our account from a previous round.
    let balances = ctx.balances();

    // Retrieve the amount of plain iota tokens from the account balance
    let amount = balances.balance(&ScColor::IOTA);

    // Create an ScMutableMap proxy to the state storage map on the host.
    let state = ctx.state();

    // retrieve the totalFactor value from the state storage through an
    // ScmutableInt64 proxy
    let total_factor = state.get_int64(VAR_TOTAL_FACTOR).value();

    // note that it is useless to try to divide anything less than totalFactor
    // iotas because every member would receive zero iotas
    if amount < total_factor {
        // log the fact that we have nothing to do to the host log
        ctx.log("dividend.divide: nothing to divide");

        // And exit the function. Note that we could not have used a require()
        // statement here, because that would have indicated an error and caused
        // a panic out of the function, returning any amount of tokens that was
        // intended to be dispersed to the members. By gracefully exiting instead
        // we keep these tokens into our account ready for dispersal in a next round.
        return;
    }

    // Create an ScMutableMap proxy to the 'members' map in the state storage.
    let members = state.get_map(VAR_MEMBERS);

    // Create an ScMutableAddressArray proxy to the 'memberList' Address array
    // in the state storage.
    let member_list = state.get_address_array(VAR_MEMBER_LIST);

    // Determine the current length of the memberList array.
    let size = member_list.length();

    // loop through all indexes of the memberList array
    for i in 0..size {
        // Retrieve the next address from the memberList array through
        // an ScMutableAddress proxy referring the value at the required index.
        let address = member_list.get_address(i).value();

        // Retrieve the factor associated with the address from the members map
        // through an ScMutableInt64 proxy referring the value in the map.
        let factor = members.get_int64(&address).value();

        // calculate the fair share of iotas to disperse to this member based
        // on the factor we just retrieved. Note that the result will been truncated.
        let share = amount * factor / total_factor;

        // is there anything to disperse to this member?
        if share > 0 {
            // Yes, so let's set up an ScTransfers proxy that transfers the calculated
            // amount of iotas. Note that ScTransfers wraps an ScMutableMap of token
            // color/amount combinations in a simpler to use interface. The constructor
            // we use here creates and initializes a single token color transfer in a
            // single statement. The actual color and amount values passed in will be
            // stored in a new map on the host.
            let transfers = ScTransfers::new(&ScColor::IOTA, share);

            // Perform the actual transfer of tokens from the smart contract to the member
            // address. The transfer_to_address() method receives the address value and
            // the proxy to the new transfers map on the host, and will call the
            // corresponding host sandbox function with these values.
            ctx.transfer_to_address(&address, transfers);
        }
    }

    // Finally, we log the fact that we have successfully completed execution
    // of the 'divide' Func in the log on the host.
    ctx.log("dividend.divide ok");
}

// 'getFactor' is a simple example of a View function. It will retrieve the
// factor associated with the (mandatory) address parameter it was provided with.
pub fn view_get_factor(ctx: &ScViewContext) {

    // Log the fact that we have initiated the 'getFactor' View in the log on the host.
    ctx.log("dividend.getFactor");

    // Now it is time to check the parameter.
    // First we create an ScImmutableMap proxy to the params map on the host.
    let p = ctx.params();

    // Create an ScImmutableAddress proxy to the 'address' parameter that is still stored
    // in the map on the host.
    let param_address= p.get_address(PARAM_ADDRESS);

    // Require that the mandatory 'address' parameter actually exists in the map on the host.
    // If it doesn't we panic out with an error message.
    ctx.require(param_address.exists(), "missing mandatory address");

    // Now that we are sure that the 'address' parameter actually exists we can
    // retrieve its actual value into an ScAddress value object
    let address = param_address.value();

    // Now that we have sorted out the parameter we will access the state storage on the host.
    // First we create an ScImmutableMap proxy to the state storage map on the host.
    // Note that this is an *immutable* map, as opposed to the mutable map we get when
    // we call the state() method on an ScFuncContext.
    let state = ctx.state();

    // Create an ScImmutableMap proxy to the 'members' map in the state storage.
    // Note that again, this is an *immutable* map as opposed to the mutable map we
    // get from the mutable state map we get through ScFuncContext.
    let members = state.get_map(VAR_MEMBERS);

    // Retrieve the factor associated with the address parameter through
    // an ScImmutableInt64 proxy to the value stored in the 'members' map.
    let factor = members.get_int64(&address).value();

    // Create an ScMutableMap proxy to the map on the host that will store the
    // key/value pairs that we want to return from this View function
    let results = ctx.results();

    // Set the value associated with the 'factor' key to the factor we got from
    // the members map through an ScMutableInt64 proxy to the results map.
    results.get_int64(VAR_FACTOR).set_value(factor);

    // Finally, we log the fact that we have successfully completed execution
    // of the 'getFactor' View in the log on the host.
    ctx.log("dividend.getFactor ok");
}
