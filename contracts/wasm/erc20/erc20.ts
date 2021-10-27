// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "../wasmlib"
import * as sc from "./index";

export function funcApprove(ctx: wasmlib.ScFuncContext, f: sc.ApproveContext): void {
}

export function funcInit(ctx: wasmlib.ScFuncContext, f: sc.InitContext): void {
}

export function funcTransfer(ctx: wasmlib.ScFuncContext, f: sc.TransferContext): void {
}

export function funcTransferFrom(ctx: wasmlib.ScFuncContext, f: sc.TransferFromContext): void {
}

export function viewAllowance(ctx: wasmlib.ScViewContext, f: sc.AllowanceContext): void {
}

export function viewBalanceOf(ctx: wasmlib.ScViewContext, f: sc.BalanceOfContext): void {
}

export function viewTotalSupply(ctx: wasmlib.ScViewContext, f: sc.TotalSupplyContext): void {
}
