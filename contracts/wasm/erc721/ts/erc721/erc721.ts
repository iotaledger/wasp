// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib";
import * as sc from "./index";

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
const BASE_URI = "my/special/base/uri/";
const ZERO = new wasmlib.ScAgentID()

///////////////////////////  HELPER FUNCTIONS  ////////////////////////////

function approve(state: sc.MutableErc721State, owner: wasmlib.ScAgentID, approved: wasmlib.ScAgentID, tokenID: wasmlib.ScHash): void {
    state.approvedAccounts().getAgentID(tokenID).setValue(approved);
    new sc.Erc721Events().approval(approved, owner, tokenID);
}

// checks if caller is owner, or one of its delegated operators
function canOperate(state: sc.MutableErc721State, caller: wasmlib.ScAgentID, owner: wasmlib.ScAgentID): boolean {
    if (caller.equals(owner)) {
        return true;
    }

    let operators = state.approvedOperators().getOperators(owner);
    return operators.getBool(caller).value();
}

// checks if caller is owner, or one of its delegated operators, or approved account for tokenID
function canTransfer(state: sc.MutableErc721State, caller: wasmlib.ScAgentID, owner: wasmlib.ScAgentID, tokenID: wasmlib.ScHash): boolean {
    if (canOperate(state, caller, owner)) {
        return true;
    }

    let controller = state.approvedAccounts().getAgentID(tokenID);
    return controller.value().equals(caller);
}

// common code for safeTransferFrom and transferFrom
function transfer(ctx: wasmlib.ScFuncContext, state: sc.MutableErc721State, from: wasmlib.ScAgentID, to: wasmlib.ScAgentID, tokenID: wasmlib.ScHash): void {
    let tokenOwner = state.owners().getAgentID(tokenID);
    ctx.require(tokenOwner.exists(), "tokenID does not exist");

    let owner = tokenOwner.value();
    ctx.require(canTransfer(state, ctx.caller(), owner, tokenID),
        "not owner, operator, or approved");

    ctx.require(owner.equals(from), "from is not owner");
    //TODO: ctx.require(to == <check-if-is-a-valid-address> , "invalid 'to' agentid");

    let nftCountFrom = state.balances().getUint64(from);
    let nftCountTo = state.balances().getUint64(to);

    nftCountFrom.setValue(nftCountFrom.value() - 1);
    nftCountTo.setValue(nftCountTo.value() + 1);

    tokenOwner.setValue(to);

    //TODO should probably clear this entry, but for now just set to zero
    approve(state, owner, ZERO, tokenID);

    new sc.Erc721Events().transfer(from, to, tokenID);
}

///////////////////////////  SC FUNCS  ////////////////////////////

// Gives permission to to to transfer tokenID token to another account.
// The approval is cleared when the token is transferred.
export function funcApprove(ctx: wasmlib.ScFuncContext, f: sc.ApproveContext): void {
    let tokenID = f.params.tokenID().value();
    let tokenOwner = f.state.owners().getAgentID(tokenID);
    ctx.require(tokenOwner.exists(), "tokenID does not exist");
    let owner = tokenOwner.value();
    ctx.require(canOperate(f.state, ctx.caller(), owner), "not owner or operator");
    let approved = f.params.approved().value();
    ctx.require(!owner.equals(approved), "approved equals owner");
    approve(f.state, owner, approved, tokenID);
}

// Destroys tokenID. The approval is cleared when the token is burned.
export function funcBurn(ctx: wasmlib.ScFuncContext, f: sc.BurnContext): void {
    let tokenID = f.params.tokenID().value();
    let owner = f.state.owners().getAgentID(tokenID).value();
    ctx.require(!owner.equals(ZERO), "tokenID does not exist");
    ctx.require(ctx.caller().equals(owner), "caller is not owner");

    approve(f.state, owner, ZERO, tokenID);

    let balance = f.state.balances().getUint64(owner);
    balance.setValue(balance.value() - 1);
    //TODO clear this instead of setting to zero
    f.state.owners().getAgentID(tokenID).setValue(ZERO);

    f.events.transfer(owner, ZERO, tokenID);
}

// Initializes the contract by setting a name and a symbol to the token collection.
export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    let name = f.params.name().value();
    let symbol = f.params.symbol().value();

    f.state.name().setValue(name);
    f.state.symbol().setValue(symbol);

    f.events.init(name, symbol);
}

// Mints tokenID and transfers it to caller as new owner.
export function funcMint(ctx: wasmlib.ScFuncContext, f: sc.MintContext): void {
    let tokenID = f.params.tokenID().value();
    let tokenOwner = f.state.owners().getAgentID(tokenID);
    ctx.require(!tokenOwner.exists(), "tokenID already minted");

    let owner = ctx.caller();
    tokenOwner.setValue(owner);
    let balance = f.state.balances().getUint64(owner);
    balance.setValue(balance.value() + 1);

    f.events.transfer(ZERO, owner, tokenID);
    if (!owner.isAddress()) {
        //TODO interpret to as SC address and call its onERC721Received() function
    }
}

// Safely transfers tokenID token from from to to, checking first that contract
// recipients are aware of the ERC721 protocol to prevent tokens from being forever locked.
export function funcSafeTransferFrom(ctx: wasmlib.ScFuncContext, f: sc.SafeTransferFromContext): void {
    let from = f.params.from().value();
    let to = f.params.to().value();
    let tokenID = f.params.tokenID().value();
    transfer(ctx, f.state, from, to, tokenID);
    if (!to.isAddress()) {
        //TODO interpret to as SC address and call its onERC721Received() function
    }
}

// Approve or remove operator as an operator for the caller.
export function funcSetApprovalForAll(ctx: wasmlib.ScFuncContext, f: sc.SetApprovalForAllContext): void {
    let owner = ctx.caller();
    let operator = f.params.operator().value();
    ctx.require(!owner.equals(operator), "owner equals operator");

    let approval = f.params.approval().value();
    let approvalsByCaller = f.state.approvedOperators().getOperators(owner);
    approvalsByCaller.getBool(operator).setValue(approval);

    f.events.approvalForAll(approval, operator, owner);
}

// Transfers tokenID token from from to to.
export function funcTransferFrom(ctx: wasmlib.ScFuncContext, f: sc.TransferFromContext): void {
    let from = f.params.from().value();
    let to = f.params.to().value();
    let tokenID = f.params.tokenID().value();
    transfer(ctx, f.state, from, to, tokenID);
}

///////////////////////////  SC VIEWS  ////////////////////////////

// Returns the number of tokens in owner's account if the owner exists.
export function viewBalanceOf(ctx: wasmlib.ScViewContext, f: sc.BalanceOfContext): void {
    let owner = f.params.owner().value();
    let nftCount = f.state.balances().getUint64(owner);
    if (nftCount.exists()) {
        f.results.amount().setValue(nftCount.value());
    }
}

// Returns the approved account for tokenID token if there is one.
export function viewGetApproved(ctx: wasmlib.ScViewContext, f: sc.GetApprovedContext): void {
    let tokenID = f.params.tokenID().value();
    let approved = f.state.approvedAccounts().getAgentID(tokenID).value();
    if (!approved.equals(ZERO)) {
        f.results.approved().setValue(approved);
    }
}

// Returns if the operator is allowed to manage all the assets of owner.
export function viewIsApprovedForAll(ctx: wasmlib.ScViewContext, f: sc.IsApprovedForAllContext): void {
    let owner = f.params.owner().value();
    let operator = f.params.operator().value();
    let operators = f.state.approvedOperators().getOperators(owner);
    let approval = operators.getBool(operator);
    if (approval.exists()) {
        f.results.approval().setValue(approval.value());
    }
}

// Returns the token collection name.
export function viewName(ctx: wasmlib.ScViewContext, f: sc.NameContext): void {
    f.results.name().setValue(f.state.name().value());
}

// Returns the owner of the tokenID token if the token exists.
export function viewOwnerOf(ctx: wasmlib.ScViewContext, f: sc.OwnerOfContext): void {
    let tokenID = f.params.tokenID().value();
    let owner = f.state.owners().getAgentID(tokenID);
    if (owner.exists()) {
        f.results.owner().setValue(owner.value());
    }
}

// Returns the token collection symbol.
export function viewSymbol(ctx: wasmlib.ScViewContext, f: sc.SymbolContext): void {
    f.results.symbol().setValue(f.state.symbol().value());
}

// Returns the Uniform Resource Identifier (URI) for tokenID token if the token exists.
export function viewTokenURI(ctx: wasmlib.ScViewContext, f: sc.TokenURIContext): void {
    let tokenID = f.params.tokenID();
    if (tokenID.exists()) {
        f.results.tokenURI().setValue(BASE_URI + tokenID.toString());
    }
}
