// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "../wasmlib"
import * as sc from "./index";

export function funcForcePayout(ctx: wasmlib.ScFuncContext, f: sc.ForcePayoutContext): void {
}

export function funcForceReset(ctx: wasmlib.ScFuncContext, f: sc.ForceResetContext): void {
}

export function funcPayWinners(ctx: wasmlib.ScFuncContext, f: sc.PayWinnersContext): void {
}

export function funcPlaceBet(ctx: wasmlib.ScFuncContext, f: sc.PlaceBetContext): void {
}

export function funcPlayPeriod(ctx: wasmlib.ScFuncContext, f: sc.PlayPeriodContext): void {
}

export function viewLastWinningNumber(ctx: wasmlib.ScViewContext, f: sc.LastWinningNumberContext): void {
}

export function viewRoundNumber(ctx: wasmlib.ScViewContext, f: sc.RoundNumberContext): void {
}

export function viewRoundStartedAt(ctx: wasmlib.ScViewContext, f: sc.RoundStartedAtContext): void {
}

export function viewRoundStatus(ctx: wasmlib.ScViewContext, f: sc.RoundStatusContext): void {
}
