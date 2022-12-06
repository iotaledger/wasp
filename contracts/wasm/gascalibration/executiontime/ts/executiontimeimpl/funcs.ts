// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

import * as wasmlib from "wasmlib";
import * as wasmtypes from "wasmlib/wasmtypes";
import * as sc from "../executiontime/index";

export function funcF(_ctx: wasmlib.ScFuncContext, f: sc.FContext): void {
    const n = f.params.n().value();
    let x: u32 = 0;
    let y: u32 = 0;

    for (let i: u32 = 0; i < n; i++) {
        x += 1;
        y = 3 * (x % 10);
    }
    f.results.n().setValue(y);
}
