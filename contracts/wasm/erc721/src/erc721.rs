// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

use wasmlib::*;

use crate::*;
use crate::typedefs::*;

// Follows ERC-721 standard as closely as possible
// https://eips.ethereum.org/EIPS/eip-721
// Notable changes w.r.t. ERC-721:
// - tokenID is Hash instead of int256
// - balance amounts are Uint64 instead of int256
// - all address accounts are replaced with AgentID accounts
// - for consistency and to reduce confusion:
//     use 'approved' when it is an AgentID
//     use 'approval' when it is a Bool

// set the required base URI, to which the base58 encoded token ID will be concatenated
const BASE_URI: &str = "my/special/base/uri/";

///////////////////////////  HELPER FUNCTIONS  ////////////////////////////

// checks if caller is owner, or one of its delegated operators
fn can_operate(state: &MutableErc721State, caller: &ScAgentID, owner: &ScAgentID) -> bool {
    if *caller == *owner {
        return true;
    }

    let operators = state.approved_operators().get_operators(owner);
    operators.get_bool(&caller).value()
}

// checks if caller is owner, or one of its delegated operators, or approved account for tokenID
fn can_transfer(state: &MutableErc721State, caller: &ScAgentID, owner: &ScAgentID, token_id: &ScHash) -> bool {
    if can_operate(state, caller, owner) {
        return true;
    }

    let controller = state.approved_accounts().get_agent_id(token_id);
    controller.value() == *caller
}

// common code for safeTransferFrom and transferFrom
fn transfer(ctx: &ScFuncContext, state: &MutableErc721State, from: &ScAgentID, to: &ScAgentID, token_id: &ScHash) {
    let token_owner = state.owners().get_agent_id(token_id);
    ctx.require(token_owner.exists(), "tokenID does not exist");

    let owner = token_owner.value();
    ctx.require(can_transfer(state, &ctx.caller(), &owner, token_id),
                "not owner, operator, or approved");

    ctx.require(owner == *from, "from is not owner");
    //TODO: ctx.require(to == <check-if-is-a-valid-address> , "invalid 'to' agentid");

    let balance_from = state.balances().get_uint64(from);
    let balance_to = state.balances().get_uint64(to);

    balance_from.set_value(balance_from.value() - 1);
    balance_to.set_value(balance_to.value() + 1);

    token_owner.set_value(to);

    let events = Erc721Events {};
    // remove approval if it exists
    let current_approved = state.approved_accounts().get_agent_id(&token_id);
    if current_approved.exists() {
        current_approved.delete();
        events.approval(&zero(), &owner, &token_id);
    }

    events.transfer(from, to, token_id);
}

fn zero() -> ScAgentID {
    ScAgentID::from_bytes(&[])
}

///////////////////////////  SC FUNCS  ////////////////////////////

// Gives permission to to to transfer tokenID token to another account.
// The approval is cleared when optional approval account is omitted.
// The approval will be cleared when the token is transferred.
pub fn func_approve(ctx: &ScFuncContext, f: &ApproveContext) {
    let token_id = f.params.token_id().value();
    let token_owner = f.state.owners().get_agent_id(&token_id);
    ctx.require(token_owner.exists(), "tokenID does not exist");
    let owner = token_owner.value();
    ctx.require(can_operate(&f.state, &ctx.caller(), &owner), "not owner or operator");

    let approved = f.params.approved();
    if !approved.exists() {
        // remove approval if it exists
        let current_approved = f.state.approved_accounts().get_agent_id(&token_id);
        if current_approved.exists() {
            current_approved.delete();
            f.events.approval(&zero(), &owner, &token_id);
        }
        return;
    }

    let account = approved.value();
    ctx.require(owner != account, "approved account equals owner");

    f.state.approved_accounts().get_agent_id(&token_id).set_value(&account);
    f.events.approval(&account, &owner, &token_id);
}

// Destroys tokenID. The approval is cleared when the token is burned.
pub fn func_burn(ctx: &ScFuncContext, f: &BurnContext) {
    let token_id = f.params.token_id().value();
    let owner = f.state.owners().get_agent_id(&token_id).value();
    ctx.require(owner != zero(), "tokenID does not exist");
    ctx.require(ctx.caller() == owner, "caller is not owner");

    // remove approval if it exists
    let current_approved = f.state.approved_accounts().get_agent_id(&token_id);
    if current_approved.exists() {
        current_approved.delete();
        f.events.approval(&zero(), &owner, &token_id);
    }

    let balance = f.state.balances().get_uint64(&owner);
    balance.set_value(balance.value() - 1);

    f.state.owners().get_agent_id(&token_id).delete();
    f.events.transfer(&owner, &zero(), &token_id);
}

// Initializes the contract by setting a name and a symbol to the token collection.
pub fn func_init(_ctx: &ScFuncContext, f: &InitContext) {
    let name = f.params.name().value();
    let symbol = f.params.symbol().value();

    f.state.name().set_value(&name);
    f.state.symbol().set_value(&symbol);

    f.events.init(&name, &symbol);
}

// Mints tokenID and transfers it to caller as new owner.
pub fn func_mint(ctx: &ScFuncContext, f: &MintContext) {
    let token_id = f.params.token_id().value();
    let token_owner = f.state.owners().get_agent_id(&token_id);
    ctx.require(!token_owner.exists(), "tokenID already minted");

    // save optional token uri
    let token_uri = f.params.token_uri();
    if token_uri.exists() {
        f.state.token_ur_is().get_string(&token_id).set_value(&token_uri.value());
    }

    let owner = ctx.caller();
    token_owner.set_value(&owner);
    let balance = f.state.balances().get_uint64(&owner);
    balance.set_value(balance.value() + 1);

    f.events.transfer(&zero(), &owner, &token_id);
    // if !owner.is_address() {
    //     //TODO interpret to as SC address and call its onERC721Received() function
    // }
}

// Safely transfers tokenID token from from to to, checking first that contract
// recipients are aware of the ERC721 protocol to prevent tokens from being forever locked.
pub fn func_safe_transfer_from(ctx: &ScFuncContext, f: &SafeTransferFromContext) {
    let from = f.params.from().value();
    let to = f.params.to().value();
    let token_id = f.params.token_id().value();
    transfer(&ctx, &f.state, &from, &to, &token_id);
    // if !to.is_address() {
    //     //TODO interpret to as SC address and call its onERC721Received() function
    // }
}

// Approve or remove operator as an operator for the caller.
pub fn func_set_approval_for_all(ctx: &ScFuncContext, f: &SetApprovalForAllContext) {
    let owner = ctx.caller();
    let operator = f.params.operator().value();
    ctx.require(owner != operator, "owner equals operator");

    let approval = f.params.approval().value();
    let operators_for_caller = f.state.approved_operators().get_operators(&owner);
    operators_for_caller.get_bool(&operator).set_value(approval);

    f.events.approval_for_all(approval, &operator, &owner);
}

// Transfers tokenID token from from to to.
pub fn func_transfer_from(ctx: &ScFuncContext, f: &TransferFromContext) {
    let from = f.params.from().value();
    let to = f.params.to().value();
    let token_id = f.params.token_id().value();
    transfer(ctx, &f.state, &from, &to, &token_id);
}

///////////////////////////  SC VIEWS  ////////////////////////////

// Returns the number of tokens in owner's account if the owner exists.
pub fn view_balance_of(_ctx: &ScViewContext, f: &BalanceOfContext) {
    let owner = f.params.owner().value();
    let balance = f.state.balances().get_uint64(&owner);
    if balance.exists() {
        f.results.amount().set_value(balance.value());
    }
}

// Returns the approved account for tokenID token if there is one.
pub fn view_get_approved(_ctx: &ScViewContext, f: &GetApprovedContext) {
    let token_id = f.params.token_id().value();
    let approved = f.state.approved_accounts().get_agent_id(&token_id);
    if approved.exists() {
        f.results.approved().set_value(&approved.value());
    }
}

// Returns if the operator is allowed to manage all the assets of owner.
pub fn view_is_approved_for_all(_ctx: &ScViewContext, f: &IsApprovedForAllContext) {
    let owner = f.params.owner().value();
    let operator = f.params.operator().value();
    let operators = f.state.approved_operators().get_operators(&owner);
    let approval = operators.get_bool(&operator);
    if approval.exists() {
        f.results.approval().set_value(approval.value());
    }
}

// Returns the token collection name.
pub fn view_name(_ctx: &ScViewContext, f: &NameContext) {
    f.results.name().set_value(&f.state.name().value());
}

// Returns the owner of the tokenID token if the token exists.
pub fn view_owner_of(_ctx: &ScViewContext, f: &OwnerOfContext) {
    let token_id = f.params.token_id().value();
    let owner = f.state.owners().get_agent_id(&token_id);
    if owner.exists() {
        f.results.owner().set_value(&owner.value());
    }
}

// Returns the token collection symbol.
pub fn view_symbol(_ctx: &ScViewContext, f: &SymbolContext) {
    f.results.symbol().set_value(&f.state.symbol().value());
}

// Returns the Uniform Resource Identifier (URI) for tokenID token if the token exists.
pub fn view_token_uri(_ctx: &ScViewContext, f: &TokenURIContext) {
    let token_id = f.params.token_id();
    if token_id.exists() {
        let mut token_uri = BASE_URI.to_owned() + &token_id.to_string();
        let saved_uri = f.state.token_ur_is().get_string(&token_id.value());
        if saved_uri.exists() {
            token_uri = saved_uri.value();
        }
        f.results.token_uri().set_value(&token_uri);
    }
}
