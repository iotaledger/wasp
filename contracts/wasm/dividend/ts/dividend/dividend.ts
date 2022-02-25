// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// This example implements 'dividend', a simple smart contract that will
// automatically disperse iota tokens which are sent to the contract to a group
// of member addresses according to predefined division factors. The intent is
// to showcase basic functionality of WasmLib through a minimal implementation
// and not to come up with a complete robust real-world solution.
// Note that we have drawn sometimes out constructs that could have been done
// in a single line over multiple statements to be able to properly document
// step by step what is happening in the code. We also unnecessarily annotate
// all 'let' statements with their assignment type to improve understanding.

import * as wasmlib from "wasmlib"
import * as sc from "./index";

// 'init' is used as a way to initialize a smart contract. It is an optional
// function that will automatically be called upon contract deployment. In this
// case we use it to initialize the 'owner' state variable so that we can later
// use this information to prevent non-owners from calling certain functions.
// The 'init' function takes a single optional parameter:
// - 'owner', which is the agent id of the entity owning the contract.
// When this parameter is omitted the owner will default to the contract creator.
export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    // The schema tool has already created a proper InitContext for this function that
    // allows us to access call parameters and state storage in a type-safe manner.

    // First we set up a default value for the owner in case the optional
    // 'owner' parameter was omitted.
    let owner: wasmlib.ScAgentID = ctx.contractCreator();

    // Now we check if the optional 'owner' parameter is present in the params map.
    if (f.params.owner().exists()) {
        // Yes, it was present, so now we overwrite the default owner with
        // the one specified by the 'owner' parameter.
        owner = f.params.owner().value();
    }

    // Now that we have sorted out which agent will be the owner of this contract
    // we will save this value in the 'owner' variable in state storage on the host.
    // Read the documentation on schema.json to understand why this state variable is
    // supported at compile-time by code generated from schema.json by the schema tool.
    f.state.owner().setValue(owner);
}

// 'member' is a function that can only be used by the entity that owns the
// 'dividend' smart contract. It can be used to define the group of member
// addresses and dispersal factors one by one prior to sending tokens to the
// smart contract's 'divide' function. The 'member' function takes 2 parameters,
// which are both required:
// - 'address', which is an Address to use as member in the group, and
// - 'factor',  which is an Uint64 relative dispersal factor associated with
//              that address
// The 'member' function will save the address/factor combination in state storage
// and also calculate and store a running sum of all factors so that the 'divide'
// function can simply start using these precalculated values when called.
export function funcMember(ctx: wasmlib.ScFuncContext, f: sc.MemberContext): void {
    // Note that the schema tool has already dealt with making sure that this function
    // can only be called by the owner and that the required parameters are present.
    // So once we get to this point in the code we can take that as a given.

    // Since we are sure that the 'factor' parameter actually exists we can
    // retrieve its actual value into an u64. Note that we use Rust's built-in
    // data types when manipulating Uint64, String, or Bytes value objects.
    let factor: u64 = f.params.factor().value();

    // Since we are sure that the 'address' parameter actually exists we can
    // retrieve its actual value into an ScAddress value type.
    let address: wasmlib.ScAddress = f.params.address().value();

    // We will store the address/factor combinations in a key/value sub-map of the
    // state storage named 'members'. The schema tool has generated an appropriately
    // type-checked proxy map for us from the schema.json state storage definition.
    // If there is no 'members' map present yet in state storage an empty map will
    // automatically be created on the host.
    let members: sc.MapAddressToMutableUint64 = f.state.members();

    // Now we create an ScMutableUint64 proxy for the value stored in the 'members'
    // map under the key defined by the 'address' parameter we retrieved earlier.
    let currentFactor: wasmlib.ScMutableUint64 = members.getUint64(address);

    // We check to see if this key/value combination exists in the 'members' map.
    if (!currentFactor.exists()) {
        // If it does not exist yet then we have to add this new address to the
        // 'memberList' array that keeps track of all address keys used in the
        // 'members' map. The schema tool has again created the appropriate type
        // for us already. Here too, if the address array was not present yet it
        // will automatically be created on the host.
        let memberList: sc.ArrayOfMutableAddress = f.state.memberList();

        // Now we will append the new address to the memberList array.
        // We create an ScMutableAddress proxy to an address value that lives
        // at the end of the memberList array (no value yet, since we're appending).
        let newAddress: wasmlib.ScMutableAddress = memberList.appendAddress();

        // And finally we append the new address to the array by telling the proxy
        // to update the value it refers to with the 'address' parameter.
        newAddress.setValue(address);

        // Note that we could have achieved the last 3 lines of code in a single line:
        // memberList.getAddress(memberList.length()).setValue(address);
    }

    // Create an ScMutableUint64 proxy named 'totalFactor' for an Uint64 value in
    // state storage. Note that we don't care whether this value exists or not,
    // because WasmLib will treat it as if it has the default value of zero.
    let totalFactor: wasmlib.ScMutableUint64 = f.state.totalFactor();

    // Now we calculate the new running total sum of factors by first getting the
    // current value of 'totalFactor' from the state storage, then subtracting the
    // current value of the factor associated with the 'address' parameter, if any
    // exists. Again, if the associated value doesn't exist, WasmLib will assume it
    // to be zero. Finally we add the factor retrieved from the parameters,
    // resulting in the new totalFactor.
    let newTotalFactor: u64 = totalFactor.value() - currentFactor.value() + factor;

    // Now we store the new totalFactor in the state storage.
    totalFactor.setValue(newTotalFactor);

    // And we also store the factor from the parameters under the address from the
    // parameters in the state storage that the proxy refers to.
    currentFactor.setValue(factor);
}

// 'divide' is a function that will take any iotas it receives and properly
// disperse them to the addresses in the member list according to the dispersion
// factors associated with these addresses.
// Anyone can send iota tokens to this function and they will automatically be
// divided over the member list. Note that this function does not deal with
// fractions. It simply truncates the calculated amount to the nearest lower
// integer and keeps any remaining iotas in its own account. They will be added
// to any next round of tokens received prior to calculation of the new
// dividend amounts.
export function funcDivide(ctx: wasmlib.ScFuncContext, f: sc.DivideContext): void {

    // Create an ScBalances map proxy to the account balances for this
    // smart contract. Note that ScBalances wraps an ScImmutableMap of
    // token color/amount combinations in a simpler to use interface.
    let allowance: wasmlib.ScBalances = ctx.allowance();

    // Retrieve the amount of plain iota tokens from the account balance.
    let amount: u64 = allowance.balance(wasmlib.IOTA);

    // Retrieve the pre-calculated totalFactor value from the state storage.
    let totalFactor: u64 = f.state.totalFactor().value();

    // Get the proxy to the 'members' map in the state storage.
    let members: sc.MapAddressToMutableUint64 = f.state.members();

    // Get the proxy to the 'memberList' array in the state storage.
    let memberList: sc.ArrayOfMutableAddress = f.state.memberList();

    // Determine the current length of the memberList array.
    let size: u32 = memberList.length();

    // Loop through all indexes of the memberList array.
    for (let i: u32 = 0; i < size; i++) {
        // Retrieve the next indexed address from the memberList array.
        let address: wasmlib.ScAddress = memberList.getAddress(i).value();

        // Retrieve the factor associated with the address from the members map.
        let factor: u64 = members.getUint64(address).value();

        // Calculate the fair share of iotas to disperse to this member based on the
        // factor we just retrieved. Note that the result will be truncated.
        let share: u64 = amount * factor / totalFactor;

        // Is there anything to disperse to this member?
        if (share > 0) {
             // Yes, so let's set up an ScTransfers map proxy that transfers the
            // calculated amount of iotas. Note that ScTransfers wraps an
            // ScMutableMap of token color/amount combinations in a simpler to use
            // interface. The constructor we use here creates and initializes a
            // single token color transfer in a single statement. The actual color
            // and amount values passed in will be stored in a new map on the host.
            let transfers: wasmlib.ScTransfers = wasmlib.ScTransfers.iotas(share);

            // Perform the actual transfer of tokens from the smart contract to the
            // member address. The transferToAddress() method receives the address
            // value and the proxy to the new transfers map on the host, and will
            // call the corresponding host sandbox function with these values.
            ctx.transferAllowed(address.asAgentID(), transfers, true);
        }
    }
}

// 'setOwner' is used to change the owner of the smart contract.
// It updates the 'owner' state variable with the provided agent id.
// The 'setOwner' function takes a single mandatory parameter:
// - 'owner', which is the agent id of the entity that will own the contract.
// Only the current owner can change the owner.
export function funcSetOwner(ctx: wasmlib.ScFuncContext, f: sc.SetOwnerContext): void {
    // Note that the schema tool has already dealt with making sure that this function
    // can only be called by the owner and that the required parameter is present.
    // So once we get to this point in the code we can take that as a given.

    // Save the new owner parameter value in the 'owner' variable in state storage.
    f.state.owner().setValue(f.params.owner().value());
}

// 'getFactor' is a simple View function. It will retrieve the factor
// associated with the (mandatory) address parameter it was provided with.
export function viewGetFactor(ctx: wasmlib.ScViewContext, f: sc.GetFactorContext): void {

    // Since we are sure that the 'address' parameter actually exists we can
    // retrieve its actual value into an ScAddress value type.
    let address: wasmlib.ScAddress = f.params.address().value();

    // Create an ScImmutableMap proxy to the 'members' map in the state storage.
    // Note that for views this is an *immutable* map as opposed to the *mutable*
    // map we can access from the *mutable* state that gets passed to funcs.
    let members: sc.MapAddressToImmutableUint64 = f.state.members();

    // Retrieve the factor associated with the address parameter.
    let factor: u64 = members.getUint64(address).value();

    // Set the factor in the results map of the function context.
    // The contents of this results map is returned to the caller of the function.
    f.results.factor().setValue(factor);
}

// 'getOwner' can be used to retrieve the current owner of the dividend contract
export function viewGetOwner(ctx: wasmlib.ScViewContext, f: sc.GetOwnerContext): void {
    f.results.owner().setValue(f.state.owner().value());
}
