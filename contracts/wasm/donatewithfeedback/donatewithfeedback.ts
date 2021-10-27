// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "../wasmlib"
import * as sc from "./index";

export function funcDonate(ctx: wasmlib.ScFuncContext, f: sc.DonateContext): void {
}

export function funcWithdraw(ctx: wasmlib.ScFuncContext, f: sc.WithdrawContext): void {
}

export function viewDonation(ctx: wasmlib.ScViewContext, f: sc.DonationContext): void {
}

export function viewDonationInfo(ctx: wasmlib.ScViewContext, f: sc.DonationInfoContext): void {
}
