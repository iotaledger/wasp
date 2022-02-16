// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib"
import * as sc from "./index";

export function funcNow(ctx: wasmlib.ScFuncContext, f: sc.NowContext): void {
}

export function viewGetTimestamp(ctx: wasmlib.ScViewContext, f: sc.GetTimestampContext): void {
    f.results.timestamp().setValue(ctx.timestamp());
}
