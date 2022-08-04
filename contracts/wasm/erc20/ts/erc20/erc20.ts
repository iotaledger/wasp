// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// implementation of ERC-20 smart contract for ISC
// following https://ethereum.org/en/developers/tutorials/understand-the-erc-20-token-smart-contract/

import * as wasmlib from "wasmlib"
import * as sc from "./index";

// Sets the allowance value for delegated account
// inputs:
//  - PARAM_DELEGATION: agentID
//  - PARAM_AMOUNT: u64
export function funcApprove(ctx: wasmlib.ScFuncContext, f: sc.ApproveContext): void {
    let delegation = f.params.delegation().value();
    let amount = f.params.amount().value();

    // all allowances are in the map under the name of he owner
    let allowances = f.state.allAllowances().getAllowancesForAgent(ctx.caller());
    allowances.getUint64(delegation).setValue(amount);
    f.events.approval(amount, ctx.caller(), delegation)
}

// onInit is a constructor entry point. It initializes the smart contract with the
// initial value of the token supply and the owner of that supply
// - input:
//   -- PARAM_SUPPLY must be nonzero positive integer. Mandatory
//   -- PARAM_CREATOR is the AgentID where initial supply is placed. Mandatory
export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
    let supply = f.params.supply().value();
    ctx.require(supply > 0, "erc20.onInit.fail: wrong 'supply' parameter");
    f.state.supply().setValue(supply);

    // we cannot use 'caller' here because onInit is always called from the 'root'
    // so, owner of the initial supply must be provided as a parameter PARAM_CREATOR to constructor (onInit)
    // assign the whole supply to creator
    let creator = f.params.creator().value();
    f.state.balances().getUint64(creator).setValue(supply);

    let t = "erc20.onInit.success. Supply: " + supply.toString() +
        ", creator:" + creator.toString();
    ctx.log(t);
}

// transfer moves tokens from caller's account to target account
// This function emits the Transfer event.
// Input:
// - PARAM_ACCOUNT: agentID
// - PARAM_AMOUNT: u64
export function funcTransfer(ctx: wasmlib.ScFuncContext, f: sc.TransferContext): void {
    let amount = f.params.amount().value();

    let balances = f.state.balances();
    let sourceAgent = ctx.caller();
    let sourceBalance = balances.getUint64(sourceAgent);
    ctx.require(sourceBalance.value() >= amount, "erc20.transfer.fail: not enough funds");

    let targetAgent = f.params.account().value();
    let targetBalance = balances.getUint64(targetAgent);

    sourceBalance.setValue(sourceBalance.value() - amount);
    targetBalance.setValue(targetBalance.value() + amount);

    f.events.transfer(amount, sourceAgent, targetAgent)
}

// Moves the amount of tokens from sender to recipient using the allowance mechanism.
// Amount is then deducted from the caller’s allowance.
// This function emits the Transfer event.
// Input:
// - PARAM_ACCOUNT: agentID   the spender
// - PARAM_RECIPIENT: agentID   the target
// - PARAM_AMOUNT: u64
export function funcTransferFrom(ctx: wasmlib.ScFuncContext, f: sc.TransferFromContext): void {
    // validate parameters
    let amount = f.params.amount().value();

    // allowances are in the map under the name of the account
    let sourceAgent = f.params.account().value();
    let allowances = f.state.allAllowances().getAllowancesForAgent(sourceAgent);
    let allowance = allowances.getUint64(ctx.caller());
    ctx.require(allowance.value() >= amount, "erc20.transferFrom.fail: not enough allowance");

    let balances = f.state.balances();
    let sourceBalance = balances.getUint64(sourceAgent);
    ctx.require(sourceBalance.value() >= amount, "erc20.transferFrom.fail: not enough funds");

    let targetAgent = f.params.recipient().value();
    let recipientBalance = balances.getUint64(targetAgent);

    sourceBalance.setValue(sourceBalance.value() - amount);
    recipientBalance.setValue(recipientBalance.value() + amount);
    allowance.setValue(allowance.value() - amount);

    f.events.transfer(amount, sourceAgent, targetAgent)
}

// the view returns max number of tokens the owner PARAM_ACCOUNT of the account
// allowed to retrieve to another party PARAM_DELEGATION
// Input:
// - PARAM_ACCOUNT: agentID
// - PARAM_DELEGATION: agentID
// Output:
// - PARAM_AMOUNT: u64
export function viewAllowance(ctx: wasmlib.ScViewContext, f: sc.AllowanceContext): void {
    // all allowances of the address 'owner' are stored in the map of the same name
    let allowances = f.state.allAllowances().getAllowancesForAgent(f.params.account().value());
    let allow = allowances.getUint64(f.params.delegation().value()).value();
    f.results.amount().setValue(allow);
}

// the view returns balance of the token held in the account
// Input:
// - PARAM_ACCOUNT: agentID
export function viewBalanceOf(ctx: wasmlib.ScViewContext, f: sc.BalanceOfContext): void {
    let balances = f.state.balances();
    let balance = balances.getUint64(f.params.account().value());
    f.results.amount().setValue(balance.value());
}

// the view returns total supply set when creating the contract (a constant).
// Output:
// - PARAM_SUPPLY: u64
export function viewTotalSupply(ctx: wasmlib.ScViewContext, f: sc.TotalSupplyContext): void {
    f.results.supply().setValue(f.state.supply().value());
}
